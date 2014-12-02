# goscrape

[![Godoc](http://img.shields.io/badge/godoc-reference-blue.svg?style=flat)](https://godoc.org/github.com/andrew-d/goscrape) [![Build Status](https://travis-ci.org/andrew-d/goscrape.svg?branch=master)](https://travis-ci.org/andrew-d/goscrape)

goscrape is a extensible structured scraper for Go.  What does a "structured
scraper" mean?  In this case, it means that you define what you want to extract
from a page in a structured, hierarchical manner, and then goscrape takes care
of pagination, splitting the input page, and calling the code to extract chunks
of data.  However, goscrape is *extensible*, allowing you to customize nearly
every step of this process.

The architecture of goscrape is roughly as follows:

- A single request to start scraping (from a given URL) is called a *scrape*.
- Each scrape consists of some number of *pages*.
- Inside each page, there's 1 or more *blocks* - some logical method of splitting
  up a page into subcomponents.  By default, there's a single block that consists
	of the pages's `<body>` element, but you can change this fairly easily.
- Inside each block, you define some number of *pieces* of data that you wish
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
```

As you can see, the entire example, including proper error handling, only takes
36 lines of code - short and sweet.

For more usage examples, see the
[examples directory](https://github.com/andrew-d/goscrape/tree/master/_examples).

## Roadmap

Here's the rough roadmap of things that I'd like to add.  If you have a feature
request, please let me know by [opening an issue](https://github.com/andrew-d/goscrape/issues/new)!

- [ ] Allow deduplication of Pieces (a custom callback?)
- [ ] Improve parallelization (scrape multiple pages at a time, but maintain order)

## License

MIT
