package elasticsearch

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"github.com/olivere/elastic"
)

var (
	// SearchES blah
	SearchES    = true
	msearchSize = 20
)

// MsearchCh channel for submitting msearch requests
var MsearchCh = make(chan *elastic.SearchRequest, 200)

// MsearchDone for closing the MsearchCh channel
var MsearchDone = make(chan bool)

// type tweet struct {
// 	Fingerprint string                `json:"fingerprint"`
// 	Message     string                `json:"message"`
// 	Retweets    int                   `json:"retweets"`
// 	Image       string                `json:"image,omitempty"`
// 	Created     time.Time             `json:"created,omitempty"`
// 	Tags        []string              `json:"tags,omitempty"`
// 	Location    string                `json:"location,omitempty"`
// 	Suggest     *elastic.SuggestField `json:"suggest_field,omitempty"`
// }

// Msearch for search and stuff
func Msearch(client *elastic.Client) {
	ctx := context.TODO()
	msearch := client.MultiSearch().Index(index)
	go func() {
		i := 0
		for {
			// Read queries from channel
			s, more := <-MsearchCh
			if more {
				i++
				msearch.Add(s)
				// Only submit when desired batch size has been reached
				if i >= msearchSize {
					fmt.Printf("Going to search %d\n", i)
					searchResult, err := msearch.Do(ctx)
					msearch = client.MultiSearch().Index(index)
					i = 0
					if err != nil {
						log.Fatal(err)
					}
					fmt.Println(searchResult.Responses)
					for _, it := range searchResult.Responses {
						fmt.Printf("Took: %d\n", it.TookInMillis)

						if it.TotalHits() > 0 {
							for _, hit := range it.Hits.Hits {
								item := make(map[string]interface{})
								err := json.Unmarshal(*hit.Source, &item)
								if err != nil {
									log.Fatal(err)
								}
								fmt.Printf("%v\n", item)
							}
							fmt.Println(it.Hits, it.TookInMillis)
						}
					}
				}
			} else {
				fmt.Println(i)
				msearch.Do(ctx)
				MsearchDone <- true
				return
			}
		}
	}()
}
