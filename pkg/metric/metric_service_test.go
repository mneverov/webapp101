package metric

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/mneverov/webapp101/pkg/scrape"
)

var f = Filter{
	Name:  "test_metric",
	Since: time.Now(),
}

func TestMetricService_Get(t *testing.T) {
	t.Run("should propagate error from db", func(t *testing.T) {
		svc, db := createTestService()
		db.On("Get", f).
			Return(nil, assert.AnError).
			Once()

		_, err := svc.Get(f)

		require.Error(t, err)
		assert.Equal(t, assert.AnError, err)
		db.AssertExpectations(t)
	})

	t.Run("should return found metrics", func(t *testing.T) {
		svc, db := createTestService()
		db.On("Get", f).
			Return(testMetrics, nil).
			Once()

		res, err := svc.Get(f)

		assert.NoError(t, err)
		assert.Equal(t, testMetrics, res.Data)
		db.AssertExpectations(t)
	})
}

func TestMetricService_Consume(t *testing.T) {
	ch := make(chan scrape.Result, 2)
	r1 := scrape.Result{
		StatusCode:        testMetrics[0].StatusCode,
		ResponseSizeBytes: testMetrics[0].ResponseSizeBytes,
		ResponseTimeMs:    testMetrics[0].ResponseTimeMs,
		CreatedAt:         testMetrics[0].CreatedAt,
	}
	r2 := scrape.Result{
		StatusCode:        testMetrics[1].StatusCode,
		ResponseSizeBytes: testMetrics[1].ResponseSizeBytes,
		ResponseTimeMs:    testMetrics[1].ResponseTimeMs,
		CreatedAt:         testMetrics[1].CreatedAt,
	}
	ch <- r1
	ch <- r2
	close(ch)

	svc, db := createTestService()
	db.On("Create", mock.Anything).
		Return(Metric{}, nil).
		Twice()

	svc.Consume("test_metric_0", ch)
	db.AssertExpectations(t)
}

func createTestService() (*Service, *mockMetricStore) {
	db := mockMetricStore{}
	svc := NewService(&db)
	return svc, &db
}
