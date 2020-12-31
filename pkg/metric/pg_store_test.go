package metric

import (
	"testing"
	"time"

	_ "github.com/lib/pq"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/mneverov/webapp101/pkg/testutil"
)

func TestMetricDB_Get(t *testing.T) {
	metricStartTime, err := time.Parse(time.RFC3339, "2020-12-21T23:00:00Z")
	require.NoError(t, err)

	conn := testutil.TestDB(t, dbOpts, "metric")
	db := NewPostgresStorage(conn)

	t.Run("should return empty slice when no metrics found", func(t *testing.T) {
		f := Filter{
			Name:  "unknown_metric",
			Since: metricStartTime,
		}
		res, err := db.Get(f)
		require.NoError(t, err)
		assert.NotNil(t, res)
		assert.Empty(t, res)
	})

	t.Run("should return found metrics filtered by timestamp", func(t *testing.T) {
		expectedOldestTimestamp := metricStartTime.Add(2 * time.Second)
		f := Filter{
			Name:  "github_jobs",
			Since: expectedOldestTimestamp,
		}

		res, err := db.Get(f)
		assert.NoError(t, err)
		assert.Len(t, res, 3)
		for _, m := range res {
			assert.Equal(t, f.Name, m.Name)
			assert.True(t, !m.CreatedAt.Before(expectedOldestTimestamp))
		}
	})
}

func TestMetricDB_Create(t *testing.T) {
	conn := testutil.TestDB(t, dbOpts, "metric")
	db := NewPostgresStorage(conn)

	t.Run("should return error when metric is invalid", func(t *testing.T) {
		unknownMetric := Metric{Name: "unknown_metric"}
		_, err := db.Create(unknownMetric)
		require.Error(t, err)
		assert.Regexp(t, unknownMetric.Name, err)
	})

	t.Run("should return successfully created metric", func(t *testing.T) {
		m := Metric{
			Name:              "example",
			StatusCode:        201,
			ResponseSizeBytes: 5,
			ResponseTimeMs:    20,
			CreatedAt:         time.Now().Truncate(time.Millisecond),
		}
		res, err := db.Create(m)
		assert.NoError(t, err)
		m.ID = res.ID
		m.CreatedAt = res.CreatedAt
		assert.Equal(t, m, res)
	})
}
