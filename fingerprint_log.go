package main

import (
	"bufio"
	"fingerprint_log/lib"
	"flag"
	"fmt"
	"log"
	_ "net/http/pprof"
	"os"
	"path"
	"runtime"
	"runtime/pprof"
	"strings"
	"time"
)

var start = time.Now()

var (
	filepath   = flag.String("f", "", "path to log `file`")
	cpuprofile = flag.String("cpuprofile", "", "write cpu profile to `file`")
	memprofile = flag.String("memprofile", "", "write memory profile to `file`")
)

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

	// readFile(*filepath, writer)
	readFile(*filepath)
	log.Println(lib.Count)
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
	lib.ReadLines(scanner)
}

func writeLines(e []string, writer *bufio.Writer) {
	lib.Counter("call_write", 1)
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

func closeFile(f *os.File) {
	f.Close()
}
