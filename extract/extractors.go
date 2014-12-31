package extract

import (
	"bytes"
	"errors"
	"fmt"
	"regexp"

	"github.com/PuerkitoBio/goquery"
	"github.com/andrew-d/goscrape"
	"golang.org/x/net/html"
)

// Const is a PieceExtractor that returns a constant value.
type Const struct {
	// The value to return when the Extract() function is called.
	Val interface{}
}

func (e Const) Extract(sel *goquery.Selection) (interface{}, error) {
	return e.Val, nil
}

var _ scrape.PieceExtractor = Const{}

// Text is a PieceExtractor that returns the combined text contents of
// the given selection.
type Text struct{}

func (e Text) Extract(sel *goquery.Selection) (interface{}, error) {
	return sel.Text(), nil
}

var _ scrape.PieceExtractor = Text{}

// MultipleText is a PieceExtractor that extracts the text from each element
// in the given selection and returns the texts as an array.
type MultipleText struct {
	// If there are no items in the selection, then return 'nil' from Extract,
	// instead of the empty list.  This signals that the result of this Piece
	// should be omitted entirely from the results, as opposed to including the
	// empty list.
	OmitIfEmpty bool
}

func (e MultipleText) Extract(sel *goquery.Selection) (interface{}, error) {
	results := []string{}

	sel.Each(func(i int, s *goquery.Selection) {
		results = append(results, s.Text())
	})

	if len(results) == 0 && e.OmitIfEmpty {
		return nil, nil
	}

	return results, nil
}

// Html extracts and returns the HTML from inside each element of the
// given selection, as a string.
//
// Note that this results in what is effectively the innerHTML of the element -
// i.e. if our selection consists of ["<p><b>ONE</b></p>", "<p><i>TWO</i></p>"]
// then the output will be: "<b>ONE</b><i>TWO</i>".
//
// The return type is a string of all the inner HTML joined together.
type Html struct{}

func (e Html) Extract(sel *goquery.Selection) (interface{}, error) {
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

var _ scrape.PieceExtractor = Html{}

// OuterHtml extracts and returns the HTML of each element of the
// given selection, as a string.
//
// To illustrate, if our selection consists of
// ["<div><b>ONE</b></div>", "<p><i>TWO</i></p>"] then the output will be:
// "<div><b>ONE</b></div><p><i>TWO</i></p>".
//
// The return type is a string of all the outer HTML joined together.
type OuterHtml struct{}

func (e OuterHtml) Extract(sel *goquery.Selection) (interface{}, error) {
	output := bytes.NewBufferString("")
	for _, node := range sel.Nodes {
		if err := html.Render(output, node); err != nil {
			return nil, err
		}
	}

	return output.String(), nil
}

var _ scrape.PieceExtractor = OuterHtml{}

// Regex runs the given regex over the contents of each element in the
// given selection, and, for each match, extracts the given subexpression.
// The return type of the extractor is a list of string matches (i.e. []string).
type Regex struct {
	// The regular expression to match.  This regular expression must define
	// exactly one parenthesized subexpression (sometimes known as a "capturing
	// group"), which will be extracted.
	Regex *regexp.Regexp

	// The subexpression of the regex to match.  If this value is not set, and if
	// the given regex has more than one subexpression, an error will be thrown.
	Subexpression int

	// When OnlyText is true, only run the given regex over the text contents of
	// each element in the selection, as opposed to the HTML contents.
	OnlyText bool

	// By default, if there is only a single match, Regex will return
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

func (e Regex) Extract(sel *goquery.Selection) (interface{}, error) {
	if e.Regex == nil {
		return nil, errors.New("no regex given")
	}
	if e.Regex.NumSubexp() == 0 {
		return nil, errors.New("regex has no subexpressions")
	}

	var subexp int
	if e.Subexpression == 0 {
		if e.Regex.NumSubexp() != 1 {
			e := fmt.Errorf(
				"regex has more than one subexpression (%d), but which to "+
					"extract was not specified",
				e.Regex.NumSubexp())
			return nil, e
		}

		subexp = 1
	} else {
		subexp = e.Subexpression
	}

	results := []string{}

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
			// The 0th entry will be the match of the entire string.  The 1st
			// entry will be the first capturing group, which is what we want to
			// extract.
			if len(submatches) > 1 {
				results = append(results, submatches[subexp])
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
		return results[0], nil
	}

	return results, nil
}

var _ scrape.PieceExtractor = Regex{}

// Attr extracts the value of a given HTML attribute from each element
// in the selection, and returns them as a list.
// The return type of the extractor is a list of attribute valueus (i.e. []string).
type Attr struct {
	// The HTML attribute to extract from each element.
	Attr string

	// By default, if there is only a single attribute extracted, AttrExtractor
	// will return the match itself (as opposed to an array containing the single
	// match).  Set AlwaysReturnList to true to disable this behaviour, ensuring
	// that the Extract function always returns an array.
	AlwaysReturnList bool

	// If no elements with this attribute are found, then return 'nil' from
	// Extract, instead of the empty list.  This signals that the result of this
	// Piece should be omitted entirely from the results, as opposed to including
	// the empty list.
	OmitIfEmpty bool
}

func (e Attr) Extract(sel *goquery.Selection) (interface{}, error) {
	if len(e.Attr) == 0 {
		return nil, errors.New("no attribute provided")
	}

	results := []string{}

	sel.Each(func(i int, s *goquery.Selection) {
		if val, found := s.Attr(e.Attr); found {
			results = append(results, val)
		}
	})

	if len(results) == 0 && e.OmitIfEmpty {
		return nil, nil
	}
	if len(results) == 1 && !e.AlwaysReturnList {
		return results[0], nil
	}

	return results, nil
}

var _ scrape.PieceExtractor = Attr{}

// Count extracts the count of elements that are matched and returns it.
type Count struct {
	// If no elements with this attribute are found, then return 'nil' from
	// Extract, instead of a number.  This signals that the result of this
	// Piece should be omitted entirely from the results, as opposed to including
	// the empty list.
	OmitIfEmpty bool
}

func (e Count) Extract(sel *goquery.Selection) (interface{}, error) {
	l := sel.Length()
	if l == 0 && e.OmitIfEmpty {
		return nil, nil
	}

	return l, nil
}
