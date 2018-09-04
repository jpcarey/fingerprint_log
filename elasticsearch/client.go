package elasticsearch

import (
	"time"

	"github.com/olivere/elastic"
)

var (
	// Client is a connection handle
	// for elasticsearch
	Client, err = elastic.NewClient(
		elastic.SetURL("http://localhost:9200"),
		elastic.SetBasicAuth("elastic", "changeme"),
		elastic.SetSniff(false),
		elastic.SetHealthcheckInterval(10*time.Second),
		elastic.SetGzip(true),
	)
)

// Create an Elasticsearch client
// client, err := elastic.NewClient(elastic.SetURL(*url), elastic.SetSniff(*sniff))
