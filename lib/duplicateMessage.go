package lib

import (
	es "fingerprint_log/elasticsearch"
	"fmt"
)

// CACHE used for checking for duplicate messages
var CACHE = make(map[string]string)

func checkDuplicate(line string) {
	key := HashString(line[25:])
	if val, ok := CACHE[key]; ok {
		// duplicate stacktrace
		fmt.Println(val)
		Counter("matched", 1)
	} else {
		// new stacktrace. store in CACHE & write modified event that includes
		// the hash in the message
		Counter("stacktraces", 1)
		CACHE[key] = ""
		es.Search(line[25:], es.Client)
	}
}
