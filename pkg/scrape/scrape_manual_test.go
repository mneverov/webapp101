package scrape

import (
	"io"
	"io/ioutil"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

type clientMock struct {
	doMock func(req *http.Request) (*http.Response, error)
}

func (c *clientMock) Do(req *http.Request) (*http.Response, error) {
	return c.doMock(req)
}

func TestScraperManual_Scrape(t *testing.T) {
	const testURL = "https://example.com"

	t.Run("should return error on invalid url", func(t *testing.T) {
		s := newHTTPScraper(
			&http.Client{},
			"http://.invalid url/",
		)
		_, err := s.scrape()

		assert.Error(t, err)
	})

	t.Run("should return error on request fail", func(t *testing.T) {
		client := clientMock{
			doMock: func(req *http.Request) (*http.Response, error) {
				return nil, assert.AnError
			},
		}
		s := newHTTPScraper(&client, testURL)
		_, err := s.scrape()

		assert.Error(t, err)
		assert.Regexp(t, testURL, err)
	})

	t.Run("should return error when fail to read response body", func(t *testing.T) {
		client := getClientWithStatusAndBody(http.StatusOK, brokenReadCloser{})

		s := newHTTPScraper(client, testURL)
		_, err := s.scrape()

		assert.Error(t, err)
		assert.Regexp(t, testURL, err)
	})

	t.Run("should return metric when service is unavailable", func(t *testing.T) {
		testStartTime := time.Now()
		client := getClientWithStatusAndBody(
			http.StatusServiceUnavailable,
			ioutil.NopCloser(strings.NewReader("")),
		)
		s := newHTTPScraper(client, testURL)
		res, err := s.scrape()

		assert.NoError(t, err)
		assert.Equal(t, http.StatusServiceUnavailable, res.StatusCode)
		assert.Zero(t, res.ResponseSizeBytes)
		assert.True(t, res.CreatedAt.After(testStartTime))
		assert.Greater(t, res.ResponseTimeMs, 0)
	})

	t.Run("success", func(t *testing.T) {
		testStartTime := time.Now()
		client := getClientWithStatusAndBody(
			http.StatusOK,
			ioutil.NopCloser(strings.NewReader("7 bytes")),
		)
		s := newHTTPScraper(client, testURL)
		res, err := s.scrape()

		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, res.StatusCode)
		assert.Equal(t, int64(7), res.ResponseSizeBytes)
		assert.True(t, res.CreatedAt.After(testStartTime))
		assert.Greater(t, res.ResponseTimeMs, 0)
	})
}

func getClientWithStatusAndBody(status int, body io.ReadCloser) httpClient {
	return &clientMock{
		doMock: func(req *http.Request) (*http.Response, error) {
			// add sleep here because the mock returns the response immediately,
			// and the response time is otherwise always zero.
			time.Sleep(1 * time.Millisecond)
			return &http.Response{
				StatusCode: status,
				Body:       body,
			}, nil
		},
	}
}

type brokenReadCloser struct {
	io.ReadCloser
}

func (brokenReadCloser) Read(_ []byte) (n int, err error) {
	return 0, assert.AnError
}

func (brokenReadCloser) Close() error {
	return nil
}
