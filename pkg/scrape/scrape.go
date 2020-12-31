package scrape

import (
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"time"

	"github.com/pkg/errors"

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
	// 1. Create a new http request with the scraper url
	req, err := http.NewRequest(http.MethodGet, c.url, nil)
	if err != nil {
		return metric.Metric{},
			errors.Wrapf(err, "failed to create request for %s", c.name)
	}

	start := time.Now()
	// 2. Use the scraper client to do the request
	resp, err := c.client.Do(req)
	if err != nil {
		return metric.Metric{}, errors.Wrapf(err, "request failed for %s", c.name)
	}

	defer func() {
		if err = resp.Body.Close(); err != nil {
			log.Printf("failed to close response, %s", err)
		}
	}()

	// 3. Calculate the result body size
	bytes, err := io.Copy(ioutil.Discard, resp.Body)
	if err != nil {
		return metric.Metric{},
			errors.Wrapf(err, "failed to read response for %s", c.name)
	}

	// 4. Measure the time and assemble the Metric
	responseTime := int(time.Since(start).Milliseconds())
	m := metric.Metric{
		Name:              c.name,
		StatusCode:        resp.StatusCode,
		ResponseSizeBytes: bytes,
		ResponseTimeMs:    responseTime,
		CreatedAt:         time.Now(),
	}

	return m, err
}
