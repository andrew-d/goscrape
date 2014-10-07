package main

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/andrew-d/goscrape"
)

func main() {
	config := &scrape.ScrapeConfig{
		DividePage: scrape.DividePageBySelector("#secondary_package .headline"),

		Pieces: []scrape.Piece{
			{Name: "type", Selector: "h5", Extractor: scrape.TextExtractor{}},
			{Name: "title", Selector: "h2 > a", Extractor: scrape.TextExtractor{}},
			{Name: "link", Selector: "h2 > a", Extractor: scrape.AttrExtractor{Attr: "href"}},
		},
	}

	scraper, err := scrape.New(config)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error creating scraper: %s\n", err)
		os.Exit(1)
	}

	results, err := scraper.Scrape("http://www.wired.com")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error scraping: %s\n", err)
		os.Exit(1)
	}

	json.NewEncoder(os.Stdout).Encode(results)
}
