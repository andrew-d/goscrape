package main

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/andrew-d/goscrape"
)

func main() {
	fetcher, err := scrape.NewPhantomJSFetcher()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error creating fetcher: %s\n", err)
		os.Exit(1)
	}

	config := &scrape.ScrapeConfig{
		Fetcher: fetcher,

		DividePage: scrape.DividePageBySelector(".linklisting > div.thing"),

		Pieces: []scrape.Piece{
			{Name: "title", Selector: "p.title > a", Extractor: scrape.TextExtractor{}},
			{Name: "link", Selector: "p.title > a", Extractor: scrape.AttrExtractor{Attr: "href"}},
			{Name: "score", Selector: "div.score.unvoted", Extractor: scrape.TextExtractor{}},
			{Name: "rank", Selector: "span.rank", Extractor: scrape.TextExtractor{}},
			{Name: "author", Selector: "a.author", Extractor: scrape.TextExtractor{}},
			{Name: "subreddit", Selector: "a.subreddit", Extractor: scrape.TextExtractor{}},

			// Note: if a self post is edited, then this will be an array with two elements.
			{Name: "date", Selector: "time", Extractor: scrape.AttrExtractor{Attr: "datetime"}},
		},
	}

	scraper, err := scrape.New(config)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error creating scraper: %s\n", err)
		os.Exit(1)
	}

	results, err := scraper.Scrape("https://www.reddit.com")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error scraping: %s\n", err)
		os.Exit(1)
	}

	json.NewEncoder(os.Stdout).Encode(results)
}
