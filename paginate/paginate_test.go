package paginate

import (
	"strings"
	"testing"

	"github.com/PuerkitoBio/goquery"
	"github.com/andrew-d/goscrape"
	"github.com/stretchr/testify/assert"
)

func selFrom(s string) *goquery.Selection {
	r := strings.NewReader(s)
	doc, err := goquery.NewDocumentFromReader(r)
	if err != nil {
		panic(err)
	}

	return doc.Selection
}

func TestBySelector(t *testing.T) {
	sel := selFrom(`<a href="http://www.google.com">foo</a>`)

	pg, err := BySelector("a", "href").NextPage("", sel)
	assert.NoError(t, err)
	assert.Equal(t, pg, "http://www.google.com")

	pg, err = BySelector("div", "xxx").NextPage("", sel)
	assert.NoError(t, err)
	assert.Equal(t, pg, "")
}

func TestByQueryParam(t *testing.T) {
	pg, err := ByQueryParam("foo").NextPage("http://www.google.com?foo=1", nil)
	assert.NoError(t, err)
	assert.Equal(t, pg, "http://www.google.com?foo=2")

	pg, err = ByQueryParam("bad").NextPage("http://www.google.com", nil)
	assert.NoError(t, err)
	assert.Equal(t, pg, "")

	pg, err = ByQueryParam("bad").NextPage("http://www.google.com?bad=asdf", nil)
	assert.NoError(t, err)
	assert.Equal(t, pg, "")
}

func TestLimitPages(t *testing.T) {
	collect := func(url string, p scrape.Paginator) []string {
		results := []string{}
		for {
			n, err := p.NextPage(url, nil)
			assert.NoError(t, err)

			if n == "" {
				return results
			}

			results = append(results, n)
			url = n
		}
	}

	tests := []struct {
		Start   string
		Limit   int
		Results []string
	}{
		{"http://www.google.com?foo=1", 0, []string{}},
		{"http://www.google.com?foo=1", 1, []string{"http://www.google.com?foo=2"}},
		{"http://www.google.com?foo=1", 2, []string{
			"http://www.google.com?foo=2",
			"http://www.google.com?foo=3",
		}},
	}

	for _, curr := range tests {
		check := collect(curr.Start, LimitPages(curr.Limit, ByQueryParam("foo")))
		assert.Equal(t, check, curr.Results)
	}
}
