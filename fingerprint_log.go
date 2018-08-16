package main

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	_ "net/http/pprof"
	"os"
	"path"
	"regexp"
	"runtime"
	"runtime/pprof"
	"strings"
	"sync"
	"time"

	"github.com/OneOfOne/xxhash"
	"github.com/blevesearch/segment"
)

var start = time.Now()

var filepath = flag.String("f", "", "path to log `file`")
var cpuprofile = flag.String("cpuprofile", "", "write cpu profile to `file`")
var memprofile = flag.String("memprofile", "", "write memory profile to `file`")

// COUNTERS
var mux sync.Mutex
var count = make(map[string]int)

func counter(bucket string, value int) {
	mux.Lock()
	count[bucket] += value
	mux.Unlock()
}

// CACHE used for storing hashed value of stack traces
var CACHE = make(map[string]string)

// Init empty Event. Used for aggregating multilines
var event []string

type opt struct {
	TokenType   int    `json:"type"`
	Startoffset int    `json:"start_offset"`
	Endoffset   int    `json:"end_offset"`
	Position    int    `json:"position"`
	Token       string `json:"token"`
}

func hashString(s string) string {
	h := xxhash.New64()
	h.Write([]byte(s))
	return fmt.Sprintf("%x", h.Sum64())
	// return fmt.Sprintf("%x", md5.Sum([]byte(s)))
}

// regex match that line starts with `[YYYY.MM.dd`
var linestart, _ = regexp.Compile("^\\[\\d{4}-\\d{2}-\\d{2}")

func lineStart(line string) bool {
	return linestart.Match([]byte(line))
}

var multiline, _ = regexp.Compile("^(?:[a-zA-Z0-9-]+\\.)+[A-Za-z0-9$]+")

func checkMultiLine(secondline string) bool {
	return multiline.Match([]byte(secondline))
}

func writeLines(e []string, writer *bufio.Writer) {
	counter("call_write", 1)
	byteArray := []byte(strings.Join(e, ""))

	if _, err := writer.Write(byteArray); err != nil {
		log.Fatal(err)
	}
}

func dstFile(outPath string) (*bufio.Writer, *os.File) {
	f, err := os.OpenFile(outPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Fatal(err)
	}
	// writer := bufio.NewWriter(f)
	// 4MiB, stat -f "st_blksize: %k" = st_blksize: 4194304
	// TODO: This could be the wrong optimal size on other drives, should figure
	// how to find this value automatically.
	writer := bufio.NewWriterSize(f, 4194000)
	return writer, f
}

func flushevent(writer *bufio.Writer) {
	// Flush event happens when a new starting line is detected.
	// We can now test if the event is actual multiline, and has what is
	// believed to be a stacktrace.
	// This essentially skips anything without a java class starting the second
	// line. This probably is not correct, and will need adjusted.
	if len(event) >= 2 && checkMultiLine(event[1]) {
		// create hash of the event, skipping the first line.
		key := hashString(strings.Join(event[2:len(event)], ""))
		timestamp := event[0][:25]
		// check if the hash already exists in the CACHE map.
		if val, ok := CACHE[key]; ok {
			// duplicate stacktrace
			counter("matched", 1)
			counter("lines_reduced", len(event[1:len(event)]))
			StackTrace := fmt.Sprintf("StackTrace: %s, %s\n", key, val)
			writeLines([]string{event[0], StackTrace}, writer)
		} else {
			// new stacktrace. store in CACHE & write modified event that includes
			// the hash in the message
			counter("stacktraces", 1)
			CACHE[key] = timestamp
			dup := fmt.Sprintf("StackTrace: %x\n", key)
			writeLines(append([]string{event[0], dup}, event[1:len(event)]...), writer)
		}
	} else if len(event) > 0 {
		// skip empty event (initial global variable is empty)
		// write the non-stacktrace event to file.
		writeLines(event, writer)
	}
}

var r = regexp.MustCompile(`(?m)((?:(?:25[0-5]|2[0-4][0-9]|[0-1]?[0-9]{1,2})\.(?:25[0-5]|2[0-4][0-9]|[0-1]?[0-9]{1,2})\.(?:25[0-5]|2[0-4][0-9]|[0-1]?[0-9]{1,2})\.(?:25[0-5]|2[0-4][0-9]|[0-1]?[0-9]{1,2}))(?::\d+)?)`)
var uid = regexp.MustCompile(`(?m)[0-9a-z]{32}`)
var node = regexp.MustCompile(`(node \[?[A-Za-z0-9-_]{22}\]?)`)
var node2 = regexp.MustCompile(`((?:{.*?}){6})`)
var nodeName = regexp.MustCompile(`^(?:\[.*?\]){3} \[(?P<NodeName>.*?)\]`)
var indexAndShard = regexp.MustCompile(`\[(.*?)\]\[(\d+)\]`)

func analyze(message []string) {
	line := strings.Join(message, "\n")

	match := nodeName.FindStringSubmatch(line)
	result := make(map[string]string)
	for i, name := range nodeName.SubexpNames() {
		if i != 0 && name != "" {
			result[name] = match[i]
		}
	}
	// name := regexp.MustCompile(result["NodeName"])

	fmt.Printf("by name: %s\n", result["NodeName"])

	// line := scanner.Text()

	fmt.Println(line)
	// fmt.Printf("%#v\n", nodeName.FindStringSubmatch(line))
	// fmt.Printf("%q\n", nodeName.SubexpNames())
	// reversed := fmt.Sprintf("${%s} ${%s}", nodeName.SubexpNames()[2], nodeName.SubexpNames()[1])
	// fmt.Println(reversed)
	line = strings.Replace(line, result["NodeName"], "", -1)
	// line = name.ReplaceAllString(line, "")

	line = r.ReplaceAllString(line, "")
	line = uid.ReplaceAllString(line, "")
	line = node.ReplaceAllString(line, "")
	line = node2.ReplaceAllString(line, "")
	line = indexAndShard.ReplaceAllString(line, "")

	fmt.Println(line)

	startOffset := 0
	var s = []opt{}
	if len(line) > 25 {
		thing := strings.NewReader(strings.ToLower(line[25:]))
		// thing := strings.NewReader(strings.ToLower(line))

		segmenter := segment.NewWordSegmenter(thing)

		for segmenter.Segment() {
			endOffset := startOffset + len(segmenter.Bytes())

			if segmenter.Type() > 1 {
				// tokenBytes := segmenter.Bytes()
				// tokenType := segmenter.Type()

				// fmt.Printf("|%6d|%4d|%4d|%10s|\n", tokenType, startOffset, endOffset, string(tokenBytes))

				test := opt{
					TokenType:   segmenter.Type(),
					Startoffset: startOffset,
					Endoffset:   endOffset,
					Position:    len(s),
					Token:       segmenter.Text(),
				}
				s = append(s, test)

				fmt.Printf("|%6d|%4d|%4d|%10s|\n", test.TokenType, test.Startoffset, test.Endoffset, test.Token)
			}
			// update start position
			startOffset = endOffset
		}
		if err := segmenter.Err(); err != nil {
			log.Fatal(err)
		}
		// fmt.Printf("%+v\n", s)

		// b, err := json.Marshal(s)
		// if err != nil {
		// 	fmt.Println("error:", err)
		// }
		// fmt.Println(string(b))
		// os.Stdout.Write(b)

	} else {
		// These are multiline messages that I'm not currently capturing
		// for now, do nothing.
		// fmt.Printf("short line: %s", line)
	}
}

func flush(message []string) {
	if len(message) > 0 {
		fmt.Println(len(message))
		analyze(message)
	}
}

func readFile(filepath string, writer *bufio.Writer) {
	file, err := os.Open(filepath)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	//scanner.Split(segment.SplitWords)

	var message []string
	for scanner.Scan() {
		line := scanner.Text()

		// process multiline messages
		if !lineStart(line) {
			message = append(message, line)
			// write prior event to file
			// flushevent(writer)
			// empty event
			// event = event[:0]
			// event = append(event, line)
		} else {
			flush(message)
			message = []string{line}
		}
	}
	if err := scanner.Err(); err != nil {
		log.Fatal(err)
	}
}

func closeFile(f *os.File) {
	f.Close()
}

func main() {

	// START CPU
	flag.Parse()
	if *cpuprofile != "" {
		f, err := os.Create(*cpuprofile)
		if err != nil {
			log.Fatal("could not create CPU profile: ", err)
		}
		if err := pprof.StartCPUProfile(f); err != nil {
			log.Fatal("could not start CPU profile: ", err)
		}
		defer pprof.StopCPUProfile()
	}
	// END CPU

	// defer close in main, pass buffered write.
	// ugly, need to find a cleaner way.
	var outFilename = path.Base(*filepath)
	var outExtension = path.Ext(outFilename)
	var outFname = outFilename[0 : len(outFilename)-len(outExtension)]
	var outFilepath = path.Join(
		path.Dir(*filepath),
		fmt.Sprintf("%s-reduced%s", outFname, outExtension))

	writer, f := dstFile(outFilepath)
	defer closeFile(f)

	readFile(*filepath, writer)
	log.Println(count)
	writer.Flush()

	// START MEM
	if *memprofile != "" {
		f, err := os.Create(*memprofile)
		if err != nil {
			log.Fatal("could not create memory profile: ", err)
		}
		runtime.GC() // get up-to-date statistics
		if err := pprof.WriteHeapProfile(f); err != nil {
			log.Fatal("could not write memory profile: ", err)
		}
		f.Close()
	}
	// END MEM
	elapsed := time.Since(start)
	log.Printf("go_collapse_log took %s", elapsed)
}
