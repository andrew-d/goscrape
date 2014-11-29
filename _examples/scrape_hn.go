package main

import (
	"encoding/json"
	"fmt"
	"os"
	"regexp"

	"github.com/PuerkitoBio/goquery"
	"github.com/andrew-d/goscrape"
	"github.com/andrew-d/goscrape/extract"
)

func main() {
	numPages := 0

	config := &scrape.ScrapeConfig{
		DividePage: scrape.DividePageBySelector("tr:nth-child(3) tr:nth-child(3n-2):not([style='height:10px'])"),

		Pieces: []scrape.Piece{
			{Name: "title", Selector: "td.title > a", Extractor: extract.Text{}},
			{Name: "link", Selector: "td.title > a", Extractor: extract.Attr{Attr: "href"}},
			{Name: "rank", Selector: "td.title[align='right']",
				Extractor: extract.Regex{Regex: regexp.MustCompile(`(\d+)`)}},
		},

		// Extract the first 3 pages of results
		NextPage: func(doc *goquery.Selection) string {
			val, found := doc.Find("a[rel='nofollow']:last-child").Attr("href")

			numPages++
			if !found || numPages >= 3 {
				return ""
			}

			return "https://news.ycombinator.com/" + val
		},
	}

	scraper, err := scrape.New(config)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error creating scraper: %s\n", err)
		os.Exit(1)
	}

	results, err := scraper.Scrape("https://news.ycombinator.com")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error scraping: %s\n", err)
		os.Exit(1)
	}

	json.NewEncoder(os.Stdout).Encode(results)
}
