package scrape

import (
	"errors"
	"net/http"
	"net/http/cookiejar"

	"code.google.com/p/go.net/publicsuffix"
	"github.com/PuerkitoBio/goquery"
)

var (
	ErrNoPieces = errors.New("no pieces in the config")
	ErrNoOrigin = errors.New("no OriginURL provided")
)

// The NextPageFunc type is used to extract the next page during a scrape.  For
// more information, please see the documentation on the ScrapeConfig type.
type NextPageFunc func(*goquery.Selection) string

// The DividePageFunc type is used to extract a page's blocks during a scrape.
// For more information, please see the documentation on the ScrapeConfig type.
type DividePageFunc func(*goquery.Selection) []*goquery.Selection

// This interface represents something that can extract data from a selection.
type PieceExtractor interface {
	// Extract some data from the given Selection and return it.  The returned
	// data should be encodable - i.e. passing it to json.Marshal should succeed.
	// If the returned data is nil, then the output from this piece will not be
	// included.
	//
	// If this function returns an error, then the scrape is aborted.
	Extract(*goquery.Selection) (interface{}, error)
}

// A Piece represents a given chunk of data that is to be extracted from every
// block in each page of a scrape.
type Piece struct {
	// The name of this piece.  Required, and will be used to aggregate results.
	Name string

	// A sub-selector within the given block to process.  Pass in "." to use
	// the root block's selector with no modification.
	// TODO(andrew-d): Consider making this an interface too.
	Selector string

	// What to extract from the Selector (above).
	Extract PieceExtractor
}

// The main configuration for a scrape.  Pass this to the New() function.
type ScrapeConfig struct {
	// PrepareClient prepares this scraper's http.Client for usage.  Use this
	// function to do things like logging in.  If the function returns an error,
	// the scrape is aborted.
	PrepareClient func(*http.Client) error

	// PrepareRequest prepares each request that will be sent, prior to sending.
	// This is useful for, e.g. setting custom HTTP headers, changing the User-
	// Agent, and so on.  If the function returns an error, then the scrape will
	// be aborted.
	//
	// Note: this function does NOT apply to requests made during the
	// PrepareClient function (above0.
	PrepareRequest func(*http.Request) error

	// ProcessResponse modifies a response that is returned from the server before
	// it is handled by the scraper.  If the function returns an error, then the
	// scrape will be aborted.
	ProcessResponse func(*http.Response) error

	// NextPage controls the progress of the scrape.  It is called for each input
	// page, starting with the origin URL, and is expected to return the URL of
	// the next page to process.  Note that order matters - calling 'NextPage' on
	// page 1 should return page 2, not page 3.  The function should return an
	// empty string when there are no more pages to process.
	//
	// If NextPage is nil, then no pagination is performed and it is assumed that
	// the OriginURL is the only page.
	// TODO(andrew-d): should this return a string, a url.URL, ???
	NextPage NextPageFunc

	// OriginURL is the URL of the first page to process.
	OriginURL string

	// DividePage splits a page into individual 'blocks'.  When scraping, we treat
	// each page as if it contains some number of 'blocks', each of which can be
	// further subdivided into what actually needs to be extracted.
	//
	// If the DividePage function is nil, then no division is performed and the
	// page is assumed to contain a single block containing the entire <body>
	// element.
	DividePage DividePageFunc

	// Pieces contains the list of data that is extracted for each block.  For
	// every block that is the result of the DividePage function (above), all of
	// the Pieces entries receives the selector representing the block, and can
	// return a result.  If the returned result is nil, then the Piece is
	// considered not to exist in this block, and is not included.
	//
	// Note: if a Piece returns an error, it results in the scrape being aborted -
	// this can be useful if you need to ensure that a given Piece is required,
	// for example.
	Pieces []Piece
}

func (c *ScrapeConfig) clone() *ScrapeConfig {
	ret := &ScrapeConfig{
		PrepareClient:   c.PrepareClient,
		PrepareRequest:  c.PrepareRequest,
		ProcessResponse: c.ProcessResponse,
		NextPage:        c.NextPage,
		OriginURL:       c.OriginURL,
		DividePage:      c.DividePage,
		Pieces:          c.Pieces,
	}
	return ret
}

type Scraper struct {
	client *http.Client
	config *ScrapeConfig
}

func New(c *ScrapeConfig) (*Scraper, error) {
	// Validate config
	if len(c.OriginURL) == 0 {
		return nil, ErrNoOrigin
	}
	if len(c.Pieces) == 0 {
		return nil, ErrNoPieces
	}

	// Set up the HTTP client
	jarOpts := &cookiejar.Options{PublicSuffixList: publicsuffix.List}
	jar, err := cookiejar.New(jarOpts)
	if err != nil {
		return nil, err
	}
	client := &http.Client{Jar: jar}

	// Clone the configuration and fill in the defaults.
	config := c.clone()
	if config.PrepareClient == nil {
		config.PrepareClient = func(*http.Client) error {
			return nil
		}
	}
	if config.PrepareRequest == nil {
		config.PrepareRequest = func(*http.Request) error {
			return nil
		}
	}
	if config.ProcessResponse == nil {
		config.ProcessResponse = func(*http.Response) error {
			return nil
		}
	}
	if config.NextPage == nil {
		config.NextPage = func(*goquery.Selection) string {
			return ""
		}
	}
	if config.DividePage == nil {
		config.DividePage = DividePageBySelector("body")
	}

	// All set!
	ret := &Scraper{
		client: client,
		config: config,
	}
	return ret, nil
}

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
		doc.Find(sel).Each(func(int, s *goquery.Selection) {
			sels = append(sels, s)
		})

		return sels
	}
	return ret
}

// TextExtractor is a PieceExtractor that returns the combined text contents of
// the given selection.
type TextExtractor struct{}

func (e *TextExtractor) Extract(sel *goquery.Selection) (interface{}, error) {
	return sel.Text(), nil
}

// HtmlExtractor extracts and returns the HTML from inside each element of the
// given selection, as a string.
//
// Note that this results in what is effectively the innerHTML of the element -
// i.e. if our selection consists of ["<p><b>ONE</b></p>", "<p><i>TWO</i></p>"]
// then the output will be: "<b>ONE</b><i>TWO</i>".
type HtmlExtractor struct{}

func (e *HtmlExtractor) Extract(sel *goquery.Selection) (interface{}, error) {
	var ret string

	sel.Each(func(int, s *Selection) {
		s.Each(func(int, s *Selection) {
			ret += s.Html()
		})
	})

	return ret, nil
}

// OuterHtmlExtractor extracts and returns the HTML of each element of the
// given selection, as a string.
//
// To illustrate, if our selection consists of
// ["<div><b>ONE</b></div>", "<p><i>TWO</i></p>"] then the output will be:
// "<div><b>ONE</b></div><p><i>TWO</i></p>".
type OuterHtmlExtractor struct{}

func (e *HtmlExtractor) Extract(sel *goquery.Selection) (interface{}, error) {
	var ret string

	sel.Each(func(int, s *Selection) {
		ret += s.Html()
	})

	return ret, nil
}

// RegexExtractor runs the given regex over the contents of each element in the
// given selection, and, for each match, extracts the first subexpression.
type RegexExtractor struct {
	// The regular expression to match.  This regular expression must define
	// exactly one parenthesized subexpressions (sometimes known as a "capturing
	// group"), which will be extracted and joined.
	Regex *regexp.Regexp

	// When OnlyText is true, only run the given regex over the text contents of
	// each element in the selection, as opposed to the HTML contents.
	OnlyText bool

	// By default, if there is only a single match of, RegexExtractor will return
	// the match itself (as opposed to an array containing the single match).
	// Set AlwaysReturnList to true to disable this behaviour, ensuring that the
	// Extract function always returns an array.
	AlwaysReturnList bool

	// If no matches of the provided regex could be extracted, then return 'nil'
	// from Extract, instead of the empty list.  This signals that the result of
	// this Piece should be omitted entirely from the results, as opposed to
	// including the empty list.
	OmitIfEmpty bool
}

func (e *RegexExtractor) Extract(sel *goquery.Selection) (interface{}, error) {
	if e.Regex == nil {
		return nil, errors.New("no regex given")
	}
	if e.Regex.NumSubexp() != 1 {
		return nil, fmt.Errorf("regex has an invalid number of subexpressions (%d != 1)",
			e.Regex.NumSubexp())
	}

	var results []string

	// For each element in the selector...
	sel.Each(func(int, s *Selection) {
		var contents string
		if e.OnlyText {
			contents = s.Text()
		} else {
			contents = s.Html()
		}

		ret := e.Regex.FindAllStringSubmatch()

		// For each regex match...
		for _, submatches := range ret {
			// The 0th entry will be the match of the entire string.  The 1st entry will
			// be the first capturing group, which is what we want to extract.
			if len(submatches) > 1 {
				results = append(results, submatches[1])
			}
		}
	})

	if len(results) == 0 && e.OmitIfEmpty {
		return nil, nil
	}
	if len(results) == 1 && !e.AlwaysReturnList {
		return results[1], nil
	}

	return results, nil
}

// AttrExtractor extracts the value of a given HTML attribute from each element
// in the selection, and returns them as a list.
type AttrExtractor struct {
	// The HTML attribute to extract from each element.
	Attr string

	// If no elements with this attribute are found, then return 'nil' from
	// Extract, instead of the empty list.  This signals that the result of this
	// Piece should be omitted entirely from the results, as opposed to including
	// the empty list.
	OmitIfEmpty bool
}

func (e *AttrExtractor) Extract(sel *goquery.Selection) (interface{}, error) {
	if len(e.Attr) == 0 {
		return errors.New("no attribute provided")
	}

	var results []string

	sel.Each(func(int, s *Selection) {
		if val, found := s.Attr(e.Attr); found {
			results = append(results, val)
		}
	})

	if len(results) == 0 && e.OmitIfEmpty {
		return nil, nil
	}

	return results, nil
}
