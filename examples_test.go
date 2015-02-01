package tvdb_test

import (
	"fmt"

	"github.com/nemith/tvdb"
)

func ExampleSearch() {
	t := tvdb.NewClient("90D7DF3AE9E4841E")
	res, err := t.SearchSeries("The Simpsons", "en")
	if err != nil {
		panic(err)
	}

	fmt.Printf("Found '%d' matches.\n", len(res))
	fmt.Printf("Name:     %s (%d)\n", res[0].Name, res[0].FirstAired.Year())
	fmt.Printf("Overview: %s\n\n", res[0].Overview)
}
