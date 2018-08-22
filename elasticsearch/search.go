package elasticsearch

import (
	"context"
	"fmt"
	"reflect"
	"time"

	"github.com/olivere/elastic"
)

type tweet struct {
	User     string                `json:"user"`
	Message  string                `json:"message"`
	Retweets int                   `json:"retweets"`
	Image    string                `json:"image,omitempty"`
	Created  time.Time             `json:"created,omitempty"`
	Tags     []string              `json:"tags,omitempty"`
	Location string                `json:"location,omitempty"`
	Suggest  *elastic.SuggestField `json:"suggest_field,omitempty"`
}

// Search for search and stuff
func Search(line string, client *elastic.Client) {
	// Search with a match query
	ctx := context.Background()
	matchQuery := elastic.NewMatchQuery("message", line)
	searchResult, err := Client.Search().
		Index("test").
		Query(matchQuery).
		// Sort("user", true). // sort by "user" field, ascending
		From(0).Size(1).
		// Pretty(true).
		Do(ctx) // execute
	if err != nil {
		if err.Error() == "elastic: Error 404 (Not Found): no such index [type=index_not_found_exception]" {
		} else {
			fmt.Println(err.Error())
			// Handle error
			panic(err)
		}
	}

	// searchResult is of type SearchResult and returns hits, suggestions,
	// and all kinds of other information from Elasticsearch.
	fmt.Printf("Query took %d milliseconds\n", searchResult.TookInMillis)
	fmt.Printf("Total Hits: %d\n", searchResult.TotalHits())

	// Each is a convenience function that iterates over hits in a search result.
	// It makes sure you don't need to check for nil values in the response.
	// However, it ignores errors in serialization. If you want full control
	// over the process, see below.
	var ttyp tweet
	for _, item := range searchResult.Each(reflect.TypeOf(ttyp)) {
		t := item.(tweet)
		fmt.Printf("ES Result Doc:\n%s: %s\n", t.User, t.Message)
	}

	// TotalHits is another convenience function that works even when something goes wrong.

	// Here's how you iterate through the search results with full control over each step.
	// if searchResult.Hits.TotalHits > 0 {
	// 	fmt.Printf("Found a total of %d tweets\n", searchResult.Hits.TotalHits)
	//
	// 	// Iterate through results
	// 	for _, hit := range searchResult.Hits.Hits {
	// 		fmt.Println(hit)
	// 		// hit.Index contains the name of the index
	//
	// 		// Deserialize hit.Source into a Tweet (could also be just a map[string]interface{}).
	// 		// var t Tweet
	// 		// err := json.Unmarshal(*hit.Source, &t)
	// 		// if err != nil {
	// 		// 	// Deserialization failed
	// 		// }
	// 		//
	// 		// // Work with tweet
	// 		// fmt.Printf("Tweet by %s: %s\n", t.User, t.Message)
	// 	}
	// } else {
	// 	// No hits
	// 	fmt.Print("Found no tweets\n")
	// }
}
