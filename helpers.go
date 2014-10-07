package scrape

import (
	"github.com/PuerkitoBio/goquery"
)

// NextPageBySelector returns a function that extracts the next page from a
// document by querying a given CSS selector and extracting the given HTML
// attribute from the resulting element.
func NextPageBySelector(sel, attr string) NextPageFunc {
	ret := func(doc *goquery.Selection) string {
		if val, found := doc.Find(sel).Attr(attr); found {
			return val
		}
		return ""
	}
	return ret
}

// DividePageBySelector returns a function that divides a page into blocks by
// CSS selector.  Each element in the page with the given selector is treated
// as a new block.
func DividePageBySelector(sel string) DividePageFunc {
	ret := func(doc *goquery.Selection) []*goquery.Selection {
		sels := []*goquery.Selection{}
		doc.Find(sel).Each(func(i int, s *goquery.Selection) {
			sels = append(sels, s)
		})

		return sels
	}
	return ret
}
