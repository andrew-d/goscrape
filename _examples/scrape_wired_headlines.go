package main

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/andrew-d/goscrape"
	"github.com/andrew-d/goscrape/extract"
)

func main() {
	config := &scrape.ScrapeConfig{
		DividePage: scrape.DividePageBySelector("#secondary_package .headline"),

		Pieces: []scrape.Piece{
			{Name: "type", Selector: "h5", Extractor: extract.Text{}},
			{Name: "title", Selector: "h2 > a", Extractor: extract.Text{}},
			{Name: "link", Selector: "h2 > a", Extractor: extract.Attr{Attr: "href"}},
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
