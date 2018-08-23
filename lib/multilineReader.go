package lib

import (
	"bufio"
	"errors"
	es "fingerprint_log/elasticsearch"
	"log"
	"regexp"
	"time"
)

// regex match that line starts with `[YYYY.MM.dd`
var linestart = regexp.MustCompile(`^\[\d{4}-\d{2}-\d{2}`)

func lineStart(line string) bool {
	return linestart.Match([]byte(line))
}

// ReadLines iterates each line and groups multiline messages together
func ReadLines(scanner *bufio.Scanner) {
	if es.IndexToEs {
		go es.Bulk(es.Client)
	}

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

	// close out es channel and wait for done.
	if es.IndexToEs {
		close(es.Ch)
		<-es.Done
	}

	if err := scanner.Err(); err != nil {
		log.Fatal(err)
	}
}

func flush(message []string) (string, error) {
	if len(message) > 0 {
		Counter("messages", 1)
		key, line := Analyze(message)
		if es.IndexToEs {
			d := es.Doc{Message: line, Timestamp: time.Now(), Hash: key}
			Counter("es_docs", 1)
			es.Ch <- d
		}
		return line, nil
	}
	return "", errors.New("Blank line")
}
