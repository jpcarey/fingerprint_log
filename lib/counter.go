package lib

import "sync"

// COUNTERS
var (
	mux   sync.Mutex
	Count = make(map[string]int)
)

// Counter is a map that allows any arbitary key to be incrimented.
func Counter(bucket string, value int) {
	mux.Lock()
	Count[bucket] += value
	mux.Unlock()
}
