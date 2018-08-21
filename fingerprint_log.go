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

	es "fingerprint_log/esclient"

	"fingerprint_log/esbulk"
)

var start = time.Now()

var (
	filepath   = flag.String("f", "", "path to log `file`")
	cpuprofile = flag.String("cpuprofile", "", "write cpu profile to `file`")
	memprofile = flag.String("memprofile", "", "write memory profile to `file`")
	// indexToEs  = true
	indexToEs  = false
	printDebug = true
)

// COUNTERS
var (
	mux   sync.Mutex
	count = make(map[string]int)
)

func counter(bucket string, value int) {
	mux.Lock()
	count[bucket] += value
	mux.Unlock()
}

// CACHE used for checking for duplicate messages
var CACHE = make(map[string]string)

func main() {

	// if err != nil {
	// 	log.Fatal(err)
	// }
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

	// readFile(*filepath, writer)
	readFile(*filepath)
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

// func readFile(filepath string, writer *bufio.Writer) {
func readFile(filepath string) {
	file, err := os.Open(filepath)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)

	var message []string
	for scanner.Scan() {
		line := scanner.Text()
		counter("lines", 1)

		// process multiline messages
		if !lineStart(line) {
			// append this line into prior matching start line
			message = append(message, line)
		} else {
			// make sure the message is flushed for processing
			flush(message)
			// this is the start of a new line / message
			message = []string{line}
		}
	}
	if err := scanner.Err(); err != nil {
		log.Fatal(err)
	}
}

func flush(message []string) {
	if len(message) > 0 {
		counter("messages", 1)
		// fmt.Println(len(message))
		analyze(message)
	}
}

type opt struct {
	TokenType   int    `json:"type"`
	Startoffset int    `json:"start_offset"`
	Endoffset   int    `json:"end_offset"`
	Position    int    `json:"position"`
	Token       string `json:"token"`
}

func stripIndetifyingData(line string) string {

	if printDebug {
		fmt.Println(line)
	}

	match := nodeName.FindStringSubmatch(line)
	result := make(map[string]string)
	for i, name := range nodeName.SubexpNames() {
		if i != 0 && name != "" {
			result[name] = match[i]
		}
	}
	line = strings.Replace(line, result["NodeName"], "node_name", -1)

	// fmt.Printf("by name: %s\n", result["NodeName"])

	// name := regexp.MustCompile(result["NodeName"])
	// line = name.ReplaceAllString(line, "")

	line = ipv4.ReplaceAllString(line, "")
	// line = uid.ReplaceAllString(line, "")
	// line = node.ReplaceAllString(line, "")
	line = node.ReplaceAllString(line, "node [node_id]")
	line = hexString.ReplaceAllString(line, "0x00000000")
	line = node2.ReplaceAllString(line, "")
	line = indexAndShard.ReplaceAllString(line, "[index][shard]")

	if printDebug {
		fmt.Println(line)
	}

	return line
}

func checkDuplicate(line string) {
	key := hashString(line[25:])
	if val, ok := CACHE[key]; ok {
		// duplicate stacktrace
		fmt.Println(val)
		counter("matched", 1)
	} else {
		// new stacktrace. store in CACHE & write modified event that includes
		// the hash in the message
		counter("stacktraces", 1)
		CACHE[key] = ""
	}
}

func analyze(message []string) {
	line := strings.Join(message, "\n")

	if indexToEs {
		esbulk.Index(line, es.Client)
	}

	line = stripIndetifyingData(line)
	checkDuplicate(line)

	startOffset := 0
	var s = []opt{}
	if len(line) > 25 {
		thing := strings.NewReader(strings.ToLower(line[25:]))
		// thing := strings.NewReader(strings.ToLower(line))

		segmenter := segment.NewWordSegmenter(thing)

		for segmenter.Segment() {
			endOffset := startOffset + len(segmenter.Bytes())

			if segmenter.Type() > 1 {
				test := opt{
					TokenType:   segmenter.Type(),
					Startoffset: startOffset,
					Endoffset:   endOffset,
					Position:    len(s),
					Token:       segmenter.Text(),
				}
				s = append(s, test)

				// if printDebug {
				// 	fmt.Printf("|%6d|%4d|%4d|%10s|\n", test.TokenType, test.Startoffset,
				// 		test.Endoffset, test.Token)
				// }
			}

			// update start position
			startOffset = endOffset
		}
		if err := segmenter.Err(); err != nil {
			log.Fatal(err)
		}

		// dump to json
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

func hashString(s string) string {
	h := xxhash.New64()
	h.Write([]byte(s))
	return fmt.Sprintf("%x", h.Sum64())
	// return fmt.Sprintf("%x", md5.Sum([]byte(s)))
}

// regex match that line starts with `[YYYY.MM.dd`
var linestart = regexp.MustCompile(`^\[\d{4}-\d{2}-\d{2}`)

func lineStart(line string) bool {
	return linestart.Match([]byte(line))
}

func closeFile(f *os.File) {
	f.Close()
}

var ipv4 = regexp.MustCompile(`(?m)((?:(?:25[0-5]|2[0-4][0-9]|[0-1]?[0-9]{1,2})` +
	`\.(?:25[0-5]|2[0-4][0-9]|[0-1]?[0-9]{1,2})` +
	`\.(?:25[0-5]|2[0-4][0-9]|[0-1]?[0-9]{1,2})` +
	`\.(?:25[0-5]|2[0-4][0-9]|[0-1]?[0-9]{1,2}))(?::\d+)?)`,
)
var uid = regexp.MustCompile(`(?m)[0-9a-z]{32}`)
var node = regexp.MustCompile(`node \[?([A-Za-z0-9-_]{22})\]?`)
var node2 = regexp.MustCompile(`((?:{.*?}){6})`)
var nodeName = regexp.MustCompile(`^(?:\[.*?\]){3} \[(?P<NodeName>.*?)\]`)
var hexString = regexp.MustCompile(`0x(?:[A-Fa-f0-9]{8})`)
var indexAndShard = regexp.MustCompile(`\[([^ "\*\\<|,>/?]+)\]\[(\d+)\]`)
