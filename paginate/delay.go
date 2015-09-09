package paginate

import (
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/andrew-d/goscrape"
)

type withDelayPaginator struct {
	delay time.Duration
	p     scrape.Paginator
}

// WithDelay returns a Paginator that will wait the given duration whenever the
// next page is requested, and will then dispatch to the underling Paginator.
func WithDelay(delay time.Duration, p scrape.Paginator) scrape.Paginator {
	return &withDelayPaginator{
		delay: delay,
		p:     p,
	}
}

func (p *withDelayPaginator) NextPage(uri string, doc *goquery.Selection) (string, error) {
	time.Sleep(p.delay)
	return p.p.NextPage(uri, doc)
}
