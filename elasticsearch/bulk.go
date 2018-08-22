// Package elasticsearch is used for writing documents to elasticsearch
package elasticsearch

import (
	"context"
	"errors"
	"fmt"
	"log"
	"math/rand"
	"sync/atomic"
	"time"

	"github.com/olivere/elastic"
	"golang.org/x/sync/errgroup"
)

var (
	index    = "test"
	typ      = "_doc"
	bulkSize = 100
	n        = 1000
)

// Index This func must be Exported, Capitalized, and comment added.
func Index(text string, client *elastic.Client) {

	// Setup a group of goroutines from the excellent errgroup package
	g, ctx := errgroup.WithContext(context.TODO())

	// The first goroutine will emit documents and send it to the second goroutine
	// via the docsc channel.
	// The second Goroutine will simply bulk insert the documents.
	type doc struct {
		Message   string    `json:"message"`
		Timestamp time.Time `json:"@timestamp"`
	}
	docsc := make(chan doc)

	begin := time.Now()

	// Goroutine to create documents
	g.Go(func() error {
		defer close(docsc)

		buf := make([]byte, 32)
		// for i := 0; i < n; i++ {
		// Generate a random ID
		_, err := rand.Read(buf)
		if err != nil {
			return err
		}
		// id := base64.URLEncoding.EncodeToString(buf)

		// Construct the document
		d := doc{
			Message:   text,
			Timestamp: time.Now(),
		}

		// Send over to 2nd goroutine, or cancel
		select {
		case docsc <- d:
		case <-ctx.Done():
			return ctx.Err()
		}
		// }
		return nil
	})

	// Second goroutine will consume the documents sent from the first and bulk insert into ES
	var total uint64
	g.Go(func() error {
		bulk := client.Bulk().Index(index).Type(typ)
		for d := range docsc {
			// Simple progress
			current := atomic.AddUint64(&total, 1)
			dur := time.Since(begin).Seconds()
			sec := int(dur)
			pps := int64(float64(current) / dur)
			fmt.Printf("%10d | %6d req/s | %02d:%02d\r", current, pps, sec/60, sec%60)

			// Enqueue the document
			bulk.Add(elastic.NewBulkIndexRequest().Doc(d))
			if bulk.NumberOfActions() >= bulkSize {
				// Commit
				res, err := bulk.Do(ctx)
				if err != nil {
					return err
				}
				if res.Errors {
					// Look up the failed documents with res.Failed(), and e.g. recommit
					return errors.New("bulk commit failed")
				}
				// "bulk" is reset after Do, so you can reuse it
			}

			select {
			default:
			case <-ctx.Done():
				return ctx.Err()
			}
		}

		// Commit the final batch before exiting
		if bulk.NumberOfActions() > 0 {
			_, err := bulk.Do(ctx)
			if err != nil {
				return err
			}
		}
		return nil
	})

	// Wait until all goroutines are finished
	if err := g.Wait(); err != nil {
		log.Fatal(err)
	}

	// Final results
	dur := time.Since(begin).Seconds()
	sec := int(dur)
	pps := int64(float64(total) / dur)
	fmt.Printf("%10d | %6d req/s | %02d:%02d\n", total, pps, sec/60, sec%60)
}