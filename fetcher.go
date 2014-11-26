package scrape

import (
	"io"
	"net/http"
	"net/http/cookiejar"

	"code.google.com/p/go.net/publicsuffix"
)

// Fetcher is the interface that must be satisfied by things that can fetch
// remote URLs and return their contents.
type Fetcher interface {
	// Prepare is called once at the beginning of the scrape.
	Prepare() error

	// Fetch is called to retrieve each document from the remote server.
	Fetch(method, url string) (io.ReadCloser, error)

	// Close is called when the scrape is finished, and can be used to clean up
	// allocated resources or perform other cleanup actions.
	Close()
}

// HttpClientFetcher is a Fetcher that uses the Go standard library's http
// client to fetch URLs.
type HttpClientFetcher struct {
	client *http.Client

	// PrepareClient prepares this fetcher's http.Client for usage.  Use this
	// function to do things like logging in.  If the function returns an error,
	// the scrape is aborted.
	PrepareClient func(*http.Client) error

	// PrepareRequest prepares each request that will be sent, prior to sending.
	// This is useful for, e.g. setting custom HTTP headers, changing the User-
	// Agent, and so on.  If the function returns an error, then the scrape will
	// be aborted.
	//
	// Note: this function does NOT apply to requests made during the
	// PrepareClient function (above).
	PrepareRequest func(*http.Request) error

	// ProcessResponse modifies a response that is returned from the server before
	// it is handled by the scraper.  If the function returns an error, then the
	// scrape will be aborted.
	ProcessResponse func(*http.Response) error
}

func NewHttpClientFetcher() (*HttpClientFetcher, error) {
	// Set up the HTTP client
	jarOpts := &cookiejar.Options{PublicSuffixList: publicsuffix.List}
	jar, err := cookiejar.New(jarOpts)
	if err != nil {
		return nil, err
	}
	client := &http.Client{Jar: jar}

	ret := &HttpClientFetcher{
		client: client,
	}
	return ret, nil
}

func (hf *HttpClientFetcher) Prepare() error {
	if hf.PrepareClient != nil {
		return hf.PrepareClient(hf.client)
	}
	return nil
}

func (hf *HttpClientFetcher) Fetch(method, url string) (io.ReadCloser, error) {
	req, err := http.NewRequest(method, url, nil)
	if err != nil {
		return nil, err
	}

	if hf.PrepareRequest != nil {
		if err = hf.PrepareRequest(req); err != nil {
			return nil, err
		}
	}

	resp, err := hf.client.Do(req)
	if err != nil {
		return nil, err
	}

	if hf.ProcessResponse != nil {
		if err = hf.ProcessResponse(resp); err != nil {
			return nil, err
		}
	}

	return resp.Body, nil
}

func (hf *HttpClientFetcher) Close() {
	return
}

// Static type assertion
var _ Fetcher = &HttpClientFetcher{}
