package elasticsearch

import (
	"time"

	"github.com/olivere/elastic"
)

var (
	// Client is a connection handle
	// for elasticsearch
	Client, err = elastic.NewClient(
		elastic.SetURL("https://5084f6be56d6407b80dcc1fa9c1628f9.us-central1.gcp.cloud.es.io:9243"),
		elastic.SetBasicAuth("elastic", "NiFgH8cFEdpGeoHb9FDAAjV1"),
		elastic.SetSniff(false),
		elastic.SetHealthcheckInterval(10*time.Second),
		elastic.SetGzip(true),
	)
)

// Create an Elasticsearch client
// client, err := elastic.NewClient(elastic.SetURL(*url), elastic.SetSniff(*sniff))
