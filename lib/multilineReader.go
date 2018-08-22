package lib

import (
	"bufio"
	"log"
	"regexp"
)

// regex match that line starts with `[YYYY.MM.dd`
var linestart = regexp.MustCompile(`^\[\d{4}-\d{2}-\d{2}`)

func lineStart(line string) bool {
	return linestart.Match([]byte(line))
}

// ReadLines iterates each line and groups multiline messages together
func ReadLines(scanner *bufio.Scanner) {
	var message []string
	for scanner.Scan() {
		line := scanner.Text()
		Counter("lines", 1)

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
		Counter("messages", 1)
		// fmt.Println(len(message))
		Analyze(message)
	}
}
