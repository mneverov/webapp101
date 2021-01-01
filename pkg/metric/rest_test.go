package metric

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/go-chi/chi"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const metricsPath = "/metrics"
const timestampString = "2006-01-02T15:04:05Z"

func TestMetricHandler_Get(t *testing.T) {
	t.Run("should return BadRequest on unknown parameter", func(t *testing.T) {
		query := "unknown=something"
		r := httptest.NewRequest(
			http.MethodGet, fmt.Sprintf("%s?%s", metricsPath, query), nil,
		)
		w := httptest.NewRecorder()

		router, _ := createTestRouter()
		router.ServeHTTP(w, r)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		assert.Regexp(t, "unknown", w.Body)
	})

	t.Run("should return BadRequest on empty name", func(t *testing.T) {
		query := "name=&timestamp=2006-01-02T15:04:05Z"
		r := httptest.NewRequest(
			http.MethodGet, fmt.Sprintf("%s?%s", metricsPath, query), nil,
		)
		w := httptest.NewRecorder()

		router, _ := createTestRouter()
		router.ServeHTTP(w, r)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("should return BadRequest on invalid date format", func(t *testing.T) {
		query := "name=test_name&since=2006-01-02T15:04:05"
		r := httptest.NewRequest(
			http.MethodGet, fmt.Sprintf("%s?%s", metricsPath, query), nil,
		)
		w := httptest.NewRecorder()

		router, _ := createTestRouter()
		router.ServeHTTP(w, r)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		assert.Regexp(t, "invalid timestamp format", w.Body)
	})

	t.Run("should propagate service error", func(t *testing.T) {
		timestamp, err := time.Parse(time.RFC3339, timestampString)
		require.NoError(t, err)

		query := fmt.Sprintf(
			"name=%s&since=%s", testMetrics[0].Name, timestampString,
		)
		r := httptest.NewRequest(
			http.MethodGet, fmt.Sprintf("%s?%s", metricsPath, query), nil,
		)
		w := httptest.NewRecorder()

		router, metricService := createTestRouter()
		metricService.
			On("Get", Filter{Name: testMetrics[0].Name, Since: timestamp}).
			Return(Metrics{}, assert.AnError).
			Once()

		router.ServeHTTP(w, r)
		assert.Equal(t, http.StatusInternalServerError, w.Code)
		metricService.AssertExpectations(t)
	})

	t.Run("should return found metrics", func(t *testing.T) {
		timestamp, err := time.Parse(time.RFC3339, timestampString)
		require.NoError(t, err)

		query := fmt.Sprintf(
			"name=%s&since=%s", testMetrics[0].Name, timestampString,
		)
		r := httptest.NewRequest(
			http.MethodGet, fmt.Sprintf("%s?%s", metricsPath, query), nil,
		)
		w := httptest.NewRecorder()

		router, metricService := createTestRouter()
		metricService.
			On("Get", Filter{Name: testMetrics[0].Name, Since: timestamp}).
			Return(Metrics{Data: []Metric{testMetrics[0]}}, nil).
			Once()

		router.ServeHTTP(w, r)

		testMetricJSON, err := json.Marshal(&Metrics{Data: []Metric{testMetrics[0]}})
		require.NoError(t, err)
		assert.JSONEq(t, string(testMetricJSON), w.Body.String())
		assert.Equal(t, http.StatusOK, w.Code)
		metricService.AssertExpectations(t)
	})

	t.Run("should return empty array when no metric found", func(t *testing.T) {
		timestamp, err := time.Parse(time.RFC3339, timestampString)
		require.NoError(t, err)

		query := fmt.Sprintf(
			"name=%s&since=%s", testMetrics[0].Name, timestampString,
		)
		r := httptest.NewRequest(
			http.MethodGet, fmt.Sprintf("%s?%s", metricsPath, query), nil,
		)
		w := httptest.NewRecorder()

		router, metricService := createTestRouter()
		metricService.
			On("Get", Filter{Name: testMetrics[0].Name, Since: timestamp}).
			Return(Metrics{Data: []Metric{}}, nil).
			Once()

		router.ServeHTTP(w, r)

		assert.JSONEq(t, `{"data":[]}`, w.Body.String())
		assert.Equal(t, http.StatusOK, w.Code)
		metricService.AssertExpectations(t)
	})
}

func createTestRouter() (http.Handler, *mockMetricService) {
	router := chi.NewRouter()
	svc := mockMetricService{}
	handler := NewHandler(&svc)
	router.Get(metricsPath, handler.Get)

	return router, &svc
}
