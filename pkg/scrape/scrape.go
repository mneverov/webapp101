package scrape

import (
	"net/http"

	"github.com/mneverov/webapp101/pkg/metric"
)

// Scraper defines methods to work with a web page scraper.
type Scraper interface {
	Scrape() (metric.Metric, error)
}

type httpClient interface {
	Do(req *http.Request) (*http.Response, error)
}

// HTTPScraper represents a web page scraper that access the page by the given
// URL via http and gathers metrics from it.
type HTTPScraper struct {
	client httpClient
	url    string
	name   string
}

// NewHTTPScraper returns a new HTTPScraper with the given params.
func NewHTTPScraper(client httpClient, url, name string) *HTTPScraper {
	return &HTTPScraper{
		client: client,
		url:    url,
		name:   name,
	}
}

// Scrape retrieves ranks and returns the ranks or an error not
// longer than the configured timeout.
func (c *HTTPScraper) Scrape() (metric.Metric, error) {
	m := metric.Metric{}
	// 1. Create a new http request with the scraper url
	// 2. Use the scraper client to do the request
	// 3. Calculate the result body size
	// 4. measure the time and assemble the Metric

	return m, nil
}
