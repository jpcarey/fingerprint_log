package lib

import (
	es "fingerprint_log/elasticsearch"
	"fmt"
)

// CACHE used for checking for duplicate messages
var CACHE = make(map[string]string)

func checkDuplicate(line string) string {
	key := HashString(line)
	if val, ok := CACHE[key]; ok {
		// duplicate stacktrace
		fmt.Println(val)
		Counter("matched", 1)
	} else {
		// new stacktrace. store in CACHE & write modified event that includes
		// the hash in the message
		Counter("stacktraces", 1)
		CACHE[key] = ""
		matched := es.Search(line, key, es.Client)

		// es.Search(line, key, es.Client)

		if matched {
			Counter("es_matched", 1)
		} else {
			fmt.Println(line)
		}
	}
	return key
}
