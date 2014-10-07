# goscrape

[![Godoc](http://img.shields.io/badge/godoc-reference-blue.svg?style=flat)](https://godoc.org/github.com/andrew-d/goscrape)

goscrape is a structured scraper for Go.  What does a "structured scraper" mean?
In this case, it means that you define what you want to extract from a page in
a structures, hierarchical manner, and then goscrape takes care of splitting up
the input page.

The architecture of goscrape is roughly as follows:

- A single request to start scraping (from a given URL) is called a scrape.
- Each scrape consists of some number of pages
- Inside each page, there's 1 or more "blocks" - some logical method of splitting
  up a page into subcomponents.  By default, there's a single block that consists
	of the pages's `<body>` element, but you can change this fairly easily.
- Inside each block, you define some number of "pieces" of data that you wish
  to extract.  Each piece consists of a name, a selector, and what data to
	extract from the current block.

This all sounds rather complicated, but in practice it's quite simple.  Here's
a short example of how to get a list of all the headlined articles from Wired
and dump them as JSON to the screen:

```go
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
```

As you can see, the entire example, including proper error handling, only takes
35 lines of code - short and sweet.

For more usage examples, see the
[examples directory](https://github.com/andrew-d/goscrape/tree/master/_examples).

## Roadmap

- [ ] Add support for using PhantomJS to grab a website's content

## License

MIT
