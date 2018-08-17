package esclient

import (
	"time"

	"github.com/olivere/elastic"
)

var (
	// client is a connection handle
	// for elasticsearch
	Client, err = elastic.NewClient(
		elastic.SetURL("https://node01:9200"),
		elastic.SetBasicAuth("elastic", "changeme"),
		elastic.SetSniff(false),
		elastic.SetHealthcheckInterval(10*time.Second),
		elastic.SetGzip(true),
	)
)

// Create an Elasticsearch client
// client, err := elastic.NewClient(elastic.SetURL(*url), elastic.SetSniff(*sniff))
