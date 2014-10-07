package scrape

import (
	"errors"
	"fmt"
	"regexp"

	"github.com/PuerkitoBio/goquery"
)

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
	var ret, h string
	var err error

	sel.EachWithBreak(func(i int, s *goquery.Selection) bool {
		s.EachWithBreak(func(i int, s *goquery.Selection) bool {
			h, err = s.Html()
			if err != nil {
				return false
			}

			ret += h
			return true
		})

		return err == nil
	})

	if err != nil {
		return nil, err
	}
	return ret, nil
}

// OuterHtmlExtractor extracts and returns the HTML of each element of the
// given selection, as a string.
//
// To illustrate, if our selection consists of
// ["<div><b>ONE</b></div>", "<p><i>TWO</i></p>"] then the output will be:
// "<div><b>ONE</b></div><p><i>TWO</i></p>".
type OuterHtmlExtractor struct{}

func (e *OuterHtmlExtractor) Extract(sel *goquery.Selection) (interface{}, error) {
	var ret, h string
	var err error

	sel.EachWithBreak(func(i int, s *goquery.Selection) bool {
		h, err = s.Html()
		if err != nil {
			return false
		}

		ret += h
		return true
	})

	if err != nil {
		return nil, err
	}
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
	var err error
	sel.EachWithBreak(func(i int, s *goquery.Selection) bool {
		var contents string
		if e.OnlyText {
			contents = s.Text()
		} else {
			contents, err = s.Html()
			if err != nil {
				return false
			}
		}

		ret := e.Regex.FindAllStringSubmatch(contents, -1)

		// A return value of nil == no match
		if ret == nil {
			return true
		}

		// For each regex match...
		for _, submatches := range ret {
			// The 0th entry will be the match of the entire string.  The 1st entry will
			// be the first capturing group, which is what we want to extract.
			if len(submatches) > 1 {
				results = append(results, submatches[1])
			}
		}

		return true
	})

	if err != nil {
		return nil, err
	}
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
		return nil, errors.New("no attribute provided")
	}

	var results []string

	sel.Each(func(i int, s *goquery.Selection) {
		if val, found := s.Attr(e.Attr); found {
			results = append(results, val)
		}
	})

	if len(results) == 0 && e.OmitIfEmpty {
		return nil, nil
	}

	return results, nil
}
