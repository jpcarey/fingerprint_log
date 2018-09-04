package lib

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	es "fingerprint_log/elasticsearch"

	"github.com/olivere/elastic"
)

var (
	processQueue = make(chan Message, 200)
	processDone  = make(chan bool)
	msearchSize  = 20
	index        = "test"
)

type batch []Message

func process(client *elastic.Client) {
	// initalize client msearch
	ctx := context.TODO()
	msearch := client.MultiSearch().Index(index)

	b := make(batch, 0)
	go func() {
		for {
			m, more := <-processQueue
			if more {
				tokens := tokenize(m.clean)
				var matchQuery *elastic.MatchQuery
				var triGram *elastic.MatchQuery
				if len(tokens) > 1024 {
					stringEnd := tokens[1024].Endoffset
					matchQuery = elastic.NewMatchQuery("message", m.clean[:stringEnd])
					triGram = elastic.NewMatchQuery("message.tri", m.clean[:stringEnd])
				} else {
					matchQuery = elastic.NewMatchQuery("message", m.clean)
					triGram = elastic.NewMatchQuery("message.tri", m.clean)
				}

				r := elastic.NewSearchRequest().Index("test").
					Source(elastic.NewSearchSource().Query(matchQuery).Query(triGram).From(0).Size(2))
				m.query = r
				msearch.Add(r)

				b = append(b, m)
				if len(b) >= msearchSize {
					// submit msearch
					searchResult, err := msearch.Do(ctx)
					if err != nil {
						log.Fatal(err)
					}
					parseSearchResponse(searchResult, b)

					// clear batch and reset search builder
					b = nil
					msearch = client.MultiSearch().Index(index)
				}

				// es.MsearchCh <- r

				// Build ES Search Request

				// triGram := elastic.NewMatchQuery("message.tri", line)
				// r := elastic.NewSearchRequest().Index("test").
				// 	Source(elastic.NewSearchSource().Query(matchQuery).Query(triGram).From(0).Size(1))

			} else {
				searchResult, err := msearch.Do(ctx)
				if err != nil {
					log.Fatal(err)
				}
				parseSearchResponse(searchResult, b)

				processDone <- true

				close(es.MsearchCh)
				<-es.MsearchDone

				return
			}
		}
	}()
}

func parseSearchResponse(r *elastic.MultiSearchResult, b batch) {
	for i, it := range r.Responses {
		m := b[i]
		fmt.Println("<|><|><|><|><|><|><|><|><|><|><|><|><|><|><|><|><|><|><|><|>")
		fmt.Println(m.clean)

		fmt.Printf("<|> Took: %d\n", it.TookInMillis)
		if it.TotalHits() > 0 {
			for _, hit := range it.Hits.Hits {
				item := make(map[string]interface{})
				err := json.Unmarshal(*hit.Source, &item)
				if err != nil {
					log.Fatal(err)
				}
				fmt.Printf("%v\n", item["message"])

			}
			fmt.Println(it.Hits, it.TookInMillis)
		} else {
			d := es.Doc{Message: m.clean, Timestamp: time.Now(), Hash: m.hash}
			Counter("es_docs", 1)
			es.Ch <- d
		}
	}
}
