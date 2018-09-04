package lib

import (
	"bufio"
	es "fingerprint_log/elasticsearch"
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
	go process(es.Client)
	if es.SearchES {
		go es.Msearch(es.Client)
	}
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

	close(processQueue)
	<-processDone

	// if es.SearchES {
	// 	close(es.MsearchCh)
	// 	<-es.MsearchDone
	// }

	// close out es channel and wait for done.
	if es.IndexToEs {
		close(es.Ch)
		<-es.Done
	}

	if err := scanner.Err(); err != nil {
		log.Fatal(err)
	}
}

func flush(message []string) {
	if len(message) > 0 {
		Counter("messages", 1)
		// key, line := Analyze(message)
		analyze(message)
		// if es.IndexToEs {
		// 	d := es.Doc{Message: line, Timestamp: time.Now(), Hash: key}
		// 	Counter("es_docs", 1)
		// 	es.Ch <- d
		// }
		// return line, nil
	}
	// return "", errors.New("Blank line")
}
