package scrape

import (
	"io"
	"strings"
)

type StringReadCloser struct {
	*strings.Reader
}

func NewStringReadCloser(s string) *StringReadCloser {
	return &StringReadCloser{strings.NewReader(s)}
}

func (s *StringReadCloser) Close() error {
	return nil
}

var _ io.ReadCloser = &StringReadCloser{}
