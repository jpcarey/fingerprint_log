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
var causedby = regexp.MustCompile(`^Caused by:`)
var causedbyCleanup = regexp.MustCompile(`(\[.*\](?:\s|$))`)

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
	stackTrace := false
	startFrom := 0
	// indentedLine := 0
	for i, m := range message {
		if i == 0 {
			continue
		}
		if javaclass.MatchString(m) {
			stackTrace = true
			// sometimes the stack trace is the last line
			// this is common with a large query dumped in the log
			if i+1 == size {
				startFrom = i + 1
			}
		}
		if stackTrace && startFrom == 0 && indented.MatchString(m) {
			startFrom = i + 1
		}
		if causedby.MatchString(m) {
			message[i] = causedbyCleanup.ReplaceAllString(m, "")
		}
	}
	return stackTrace, startFrom
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
	}
}
