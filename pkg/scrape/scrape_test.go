package scrape

import (
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/jarcoal/httpmock"
	"github.com/stretchr/testify/assert"
)

func TestScraperHttpMock_Scrape(t *testing.T) {
	const (
		testURL        = "https://example.com"
		testMetricName = "test_metric_name"
	)

	t.Run("should return error on invalid url", func(t *testing.T) {
		s := NewHTTPScraper(
			&http.Client{},
			"http://.invalid url/",
			testMetricName,
		)
		_, err := s.Scrape()

		assert.Error(t, err)
	})

	t.Run("should return error on request fail", func(t *testing.T) {
		httpmock.Activate()
		defer httpmock.DeactivateAndReset()

		httpmock.RegisterResponder(
			http.MethodGet,
			testURL,
			httpmock.NewErrorResponder(assert.AnError),
		)

		c := NewHTTPScraper(&http.Client{}, testURL, testMetricName)
		_, err := c.Scrape()

		assert.Error(t, err)
		assert.Regexp(t, testMetricName, err)
	})

	t.Run("should return metric when service is unavailable", func(t *testing.T) {
		testStartTime := time.Now()
		httpmock.Activate()
		defer httpmock.DeactivateAndReset()

		httpmock.RegisterResponder(
			http.MethodGet,
			testURL,
			func(*http.Request) (*http.Response, error) {
				time.Sleep(1 * time.Millisecond)
				return &http.Response{
					Status:        strconv.Itoa(http.StatusServiceUnavailable),
					StatusCode:    http.StatusServiceUnavailable,
					Body:          ioutil.NopCloser(strings.NewReader("")),
					Header:        http.Header{},
					ContentLength: -1,
				}, nil
			},
		)

		s := NewHTTPScraper(&http.Client{}, testURL, testMetricName)
		res, err := s.Scrape()

		assert.NoError(t, err)
		assert.Equal(t, testMetricName, res.Name)
		assert.Equal(t, http.StatusServiceUnavailable, res.StatusCode)
		assert.Zero(t, res.ResponseSizeBytes)
		assert.True(t, res.CreatedAt.After(testStartTime))
		assert.Greater(t, res.ResponseTimeMs, 0)
	})

	t.Run("success", func(t *testing.T) {
		testStartTime := time.Now()

		httpmock.Activate()
		defer httpmock.DeactivateAndReset()

		httpmock.RegisterResponder(
			http.MethodGet,
			testURL,
			func(*http.Request) (*http.Response, error) {
				time.Sleep(1 * time.Millisecond)
				return &http.Response{
					Status:        strconv.Itoa(http.StatusOK),
					StatusCode:    http.StatusOK,
					Body:          ioutil.NopCloser(strings.NewReader("7 bytes")),
					Header:        http.Header{},
					ContentLength: -1,
				}, nil
			},
		)

		s := NewHTTPScraper(&http.Client{}, testURL, testMetricName)
		res, err := s.Scrape()

		assert.NoError(t, err)
		assert.Equal(t, testMetricName, res.Name)
		assert.Equal(t, http.StatusOK, res.StatusCode)
		assert.Equal(t, int64(7), res.ResponseSizeBytes)
		assert.True(t, res.CreatedAt.After(testStartTime))
		assert.Greater(t, res.ResponseTimeMs, 0)
	})
}
