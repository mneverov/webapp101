package scrape

import (
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"time"

	"github.com/pkg/errors"
)

// Result represent a result of a single web page scrape.
type Result struct {
	StatusCode        int
	ResponseSizeBytes int64
	ResponseTimeMs    int
	CreatedAt         time.Time
}

// scraper defines methods to work with a web page scraper.
type scraper interface {
	scrape() (Result, error)
}

// HTTPScraper represents a web page scraper that access the page by the given
// URL via http and gathers metrics from it.
type HTTPScraper struct {
	client httpClient
	url    string
}

// newHTTPScraper returns a new HTTPScraper with the given params.
func newHTTPScraper(client httpClient, url string) *HTTPScraper {
	return &HTTPScraper{
		client: client,
		url:    url,
	}
}

// scrape retrieves ranks and returns the ranks or an error not
// longer than the configured timeout.
func (c *HTTPScraper) scrape() (Result, error) {
	// 1. Create a new http request with the scraper url
	req, err := http.NewRequest(http.MethodGet, c.url, nil)
	if err != nil {
		return Result{},
			errors.Wrapf(err, "failed to create request for %s", c.url)
	}

	start := time.Now()
	// 2. Use the scraper client to do the request
	resp, err := c.client.Do(req)
	if err != nil {
		return Result{}, errors.Wrapf(err, "request failed for %s", c.url)
	}

	defer func() {
		if err = resp.Body.Close(); err != nil {
			log.Printf("failed to close response, %s", err)
		}
	}()

	// 3. Calculate the result body size
	bytes, err := io.Copy(ioutil.Discard, resp.Body)
	if err != nil {
		return Result{},
			errors.Wrapf(err, "failed to read response for %s", c.url)
	}

	// 4. Measure the time and assemble the Metric
	responseTime := int(time.Since(start).Milliseconds())
	m := Result{
		StatusCode:        resp.StatusCode,
		ResponseSizeBytes: bytes,
		ResponseTimeMs:    responseTime,
		CreatedAt:         time.Now(),
	}

	return m, err
}
