package lib

// CACHE used for checking for duplicate messages
var CACHE = make(map[string]string)

// func checkDuplicate(line string) string {
func duplicate(h string) bool {
	// line := dup.line
	// key := HashString(line)
	if _, ok := CACHE[h]; ok {
		// duplicate stacktrace
		// fmt.Printf("matched: %s\n", val)
		Counter("matched", 1)
		return true
	} else {
		// new stacktrace. store in CACHE & write modified event that includes
		// the hash in the message
		Counter("unique_messages", 1)
		CACHE[h] = ""
		return false
	}
}
