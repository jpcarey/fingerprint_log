package lib

import (
	"fmt"

	"github.com/OneOfOne/xxhash"
)

// HashString returns xxhash base64 string
func HashString(s string) string {
	h := xxhash.New64()
	h.Write([]byte(s))
	return fmt.Sprintf("%x", h.Sum64())
	// return fmt.Sprintf("%x", md5.Sum([]byte(s)))
}
