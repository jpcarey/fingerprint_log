package lib

import (
  "strings"
  "fmt"
  "log"

  "github.com/blevesearch/segment"
  es "fingerprint_log/elasticsearch"
)

var (
  // indexToEs  = true
  indexToEs   = false
	printTokens = false
)

type opt struct {
	TokenType   int    `json:"type"`
	Startoffset int    `json:"start_offset"`
	Endoffset   int    `json:"end_offset"`
	Position    int    `json:"position"`
	Token       string `json:"token"`
}

// Analyze segments tokens
func Analyze(message []string) {
	line := strings.Join(message, "\n")

	if indexToEs {
		es.Index(line, es.Client)
	}

	line = StripIndetifyingData(line)
	checkDuplicate(line)

	startOffset := 0
	var s = []opt{}
	if len(line) > 25 {
		thing := strings.NewReader(strings.ToLower(line[25:]))
		// thing := strings.NewReader(strings.ToLower(line))

		segmenter := segment.NewWordSegmenter(thing)

		for segmenter.Segment() {
			endOffset := startOffset + len(segmenter.Bytes())

			if segmenter.Type() > 1 {
				test := opt{
					TokenType:   segmenter.Type(),
					Startoffset: startOffset,
					Endoffset:   endOffset,
					Position:    len(s),
					Token:       segmenter.Text(),
				}

				s = append(s, test)

				if printTokens {
					fmt.Printf("|%6d|%4d|%4d|%10s|\n", test.TokenType, test.Startoffset,
						test.Endoffset, test.Token)
				}
			}

			// update start position
			startOffset = endOffset
		}
		if err := segmenter.Err(); err != nil {
			log.Fatal(err)
		}

		// dump to json
		// fmt.Printf("%+v\n", s)
		// b, err := json.Marshal(s)
		// if err != nil {
		// 	fmt.Println("error:", err)
		// }
		// fmt.Println(string(b))
		// os.Stdout.Write(b)

	} else {
		// These are multiline messages that I'm not currently capturing
		// for now, do nothing.
		// fmt.Printf("short line: %s", line)
	}
}
