package lib

import (
	"regexp"
	"strings"

	"github.com/olivere/elastic"
)

var (
	// indexToEs   = false
	printTokens = false
)

type String string

type Message struct {
	org           string
	clean         string
	hash          string
	stacktrace    bool
	query         *elastic.SearchRequest
	numberoflines int
}

var javaclass = regexp.MustCompile(`^(?:[a-zA-Z0-9-]+\.)+[A-Za-z0-9$]+`)
var indented = regexp.MustCompile(`^\s+`)

// func (s String) tolower() String {
// 	return String(strings.ToLower(string(s)))
// }
//
// func (s String) toupper() String {
// 	return String(strings.ToUpper(string(s)))
// }
//
// func (s String) removeTimestamp() String {
// 	return String(string(s)[10:])
// }
//
// func (s String) tostring() string {
// 	return string(s)
// }

// func removeTimestamp(line string) string {
// 	return line[25:]
// }

func checkJavaClass(message []string) (bool, int) {
	size := len(message)
	for i, m := range message {
		if i == 0 {
			continue
		}
		if javaclass.MatchString(m) {
			if i+1 == size {
				return true, i + 1
			} else if i+1 < size && indented.MatchString(message[i+1]) {
				return true, i + 1
			} else if i+2 < size && indented.MatchString(message[i+2]) {
				return true, i + 2
			} else if i+3 < size && indented.MatchString(message[i+3]) {
				return true, i + 3
			}
			// } else if (i+2) < size && strings.HasPrefix(message[i+2], "  ") {
			// 	return true, i + 2
			// } else if (i+3) < size && strings.HasPrefix(message[i+3], "  ") {
			// 	return true, i + 3
			// 	fmt.Println("WTF!!")
			// }

			// } else if i < size-1 && strings.HasPrefix(message[i+1], "	") {
			// 	return true, i + 1
			// } else if i < size-2 && strings.HasPrefix(message[i+2], " ") {
			// 	return true, i + 2
			// }
		}
	}
	return false, 0
}

func analyze(message []string) {
	m := Message{
		org:           strings.Join(message, "\n"),
		stacktrace:    false,
		numberoflines: len(message),
	}

	// message = filter(message)
	// m.stripped = String(m.org).filter().removeTimestamp().tolower().tostring()

	// Check if the log message contains a stack trace
	stacktr, stacktrStart := checkJavaClass(message)
	if stacktr {
		m.stacktrace = true
		m.hash = HashArray(message[stacktrStart:])
		m.clean = strings.Join(message[stacktrStart:], "\n")

	} else {
		message = filter(message)
		m.hash = HashArray(message)
		m.clean = strings.Join(message, "\n")

	}

	if !(duplicate(m.hash)) {
		processQueue <- m

		// fmt.Println("<|><|><|><|><|><|><|><|><|><|><|><|><|><|><|><|><|><|><|><|>")
		// data := dup{line: m.stripped, queryEnd: 0}
		// if len(tokens) > 1024 {
		// 	// fmt.Printf(">>>>>>> size: %d\n", len(s))
		// 	data.queryEnd = tokens[1024].Endoffset
		// }

		// if dup.queryEnd > 0 {
		// 	line = line[:dup.queryEnd]
		// }

		// Build ES Search Request
		// matchQuery := elastic.NewMatchQuery("message", line)
		// triGram := elastic.NewMatchQuery("message.tri", line)
		// r := elastic.NewSearchRequest().Index("test").
		// 	Source(elastic.NewSearchSource().Query(matchQuery).Query(triGram).From(0).Size(1))

		// es.MsearchCh <- r

		//
		// es.Search(line, key, es.Client)
		// matched := es.Search(line, key, es.Client)
		// if matched {
		// 	Counter("es_matched", 1)
		// } else {
		// 	fmt.Println(line)
		// }

	}

	//
	// if err := segmenter.Err(); err != nil {
	// 	log.Fatal(err)
	// }

	// key := checkDuplicate(data)

	// if indexToEs {
	// 	d := es.Doc{
	// 		Message:   line,
	// 		Timestamp: time.Now(),
	// 		Hash:      key,
	// 	}
	//
	// 	go es.Index(d, es.Client)
	// }
	// return key, m.stripped
}
