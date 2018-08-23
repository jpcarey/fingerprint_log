// Package elasticsearch is used for writing documents to elasticsearch
package elasticsearch

import (
	"context"
	"time"

	"github.com/olivere/elastic"
)

var (
	// IndexToEs enable es indexing.
	IndexToEs = true
	index     = "test"
	typ       = "_doc"
	bulkSize  = 50
	n         = 200
)

// Doc holds the documents for indexing
type Doc struct {
	Message   string    `json:"message"`
	Timestamp time.Time `json:"@timestamp"`
	Hash      string    `json:"fingerprint"`
}

// Ch is used to receive data for indexing.
var Ch = make(chan Doc, n)

// Done signal the end
var Done = make(chan bool)

// Index This func must be Exported, Capitalized, and comment added.
// func Index(d Doc, client *elastic.Client) {
// 	// go func() { ch <- d }()
// 	Ch <- d
// 	bulk(client)
// }

// Bulk provides bulk indexing to elasticsearch
func Bulk(client *elastic.Client) {
	bulk := client.Bulk().Index(index).Type(typ)
	ctx := context.TODO()

	go func() {
		for {
			d, more := <-Ch
			if more {
				bulk.Add(elastic.NewBulkIndexRequest().Doc(d))
				if bulk.NumberOfActions() >= bulkSize {
					// Commit
					// res, _ := bulk.Do(ctx)
					bulk.Do(ctx)
					// "bulk" is reset after Do, so you can reuse it
				}
			} else {
				bulk.Do(ctx)
				Done <- true
				return
			}
		}
	}()
}
