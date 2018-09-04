package lib

import (
	"fmt"
	"strings"

	"github.com/blevesearch/segment"
)

type token struct {
	TokenType   int    `json:"type"`
	Startoffset int    `json:"start_offset"`
	Endoffset   int    `json:"end_offset"`
	Position    int    `json:"position"`
	Token       string `json:"token"`
}

type tokens []token

func tokenize(s string) tokens {
	t := make(tokens, 0)
	startOffset := 0
	// if len(s) > 0 {
	segmenter := segment.NewWordSegmenter(strings.NewReader(s))

	for segmenter.Segment() {
		endOffset := startOffset + len(segmenter.Bytes())

		if segmenter.Type() > 0 {
			n := token{
				TokenType:   segmenter.Type(),
				Startoffset: startOffset,
				Endoffset:   endOffset,
				Position:    len(t),
				Token:       segmenter.Text(),
			}

			t = append(t, n)

			if printTokens {
				fmt.Printf("|%6d|%4d|%4d|%10s|\n", n.TokenType, n.Startoffset,
					n.Endoffset, n.Token)
			}
		}

		// update start position
		startOffset = endOffset
	}

	// dump to json
	// fmt.Printf("%+v\n", s)
	// b, err := json.Marshal(s)
	// if err != nil {
	// 	fmt.Println("error:", err)
	// }
	// fmt.Println(string(b))
	// os.Stdout.Write(b)

	// } else {
	// 	// These are multiline messages that I'm not currently capturing
	// 	// for now, do nothing.
	// 	// fmt.Printf("short line: %s", line)
	// }
	return t
}
