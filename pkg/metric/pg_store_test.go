package metric

import (
	"database/sql"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/go-pg/migrations/v8"
	"github.com/go-pg/pg/v10"
	"github.com/go-testfixtures/testfixtures/v3"
	_ "github.com/lib/pq"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMetricDB_Get(t *testing.T) {
	metricStartTime, err := time.Parse(time.RFC3339, "2020-12-21T23:00:00Z")
	require.NoError(t, err)

	db := testDB(t)

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
	db := testDB(t)

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

func testDB(t *testing.T) *Postgres {
	opts := pg.Options{
		Addr:     "127.0.0.1:5544",
		User:     "webapp101",
		Password: "webapp101",
		Database: "webapp101_test",
	}
	db, err := NewPostgresStorage(&opts)
	require.NoError(t, err)

	migrate(t, &opts)

	fixturesDB, err := connectDB(&opts)
	if err != nil {
		t.Fatalf("metric: failed to connect %+v\n", err)
	}
	err = loadFixtures(fixturesDB)
	if err != nil {
		t.Fatalf("metric: failed to load fixtures %+v\n", err)
	}

	return db
}

func migrate(t *testing.T, opts *pg.Options) {
	col := migrations.NewCollection()
	err := col.DiscoverSQLMigrations("../../resources/migrations")
	if err != nil {
		t.Fatalf("metric: failed to discover migrations %+v\n", err)
	}
	// need to have *pg.DB, not our defined PostgresStorage
	migrationsDB := pg.Connect(opts)
	defer migrationsDB.Close()

	// it is mandatory to run init before migrations:
	// see https://github.com/go-pg/migrations/issues/48
	_, _, err = col.Run(migrationsDB, "init")
	if err != nil {
		t.Fatalf("metric: failed to init migrations %+v\n", err)
	}

	oldVersion, newVersion, err := col.Run(migrationsDB, "reset")
	if err != nil {
		t.Fatalf("metric: failed to reset migrations %+v\n", err)
	}
	t.Logf("config: reset migrations from %d to %d \n", oldVersion, newVersion)

	// even if Run accepts vararg of commands, all should run separately
	oldVersion, newVersion, err = col.Run(migrationsDB, "up")
	if err != nil {
		t.Fatalf("metric: failed to apply migrations %+v\n", err)
	}
	t.Logf("metric: apply migrations from %d to %d \n", oldVersion, newVersion)
}

func loadFixtures(conn *sql.DB) error {
	fixtures, err := testfixtures.New(
		testfixtures.Database(conn),
		testfixtures.Dialect("postgres"),
		testfixtures.Directory("../../resources/fixtures"),
	)
	if err != nil {
		return err
	}

	return fixtures.Load()
}

func connectDB(opts *pg.Options) (*sql.DB, error) {
	source := fmt.Sprintf("user=%s password='%s' host=%s port=%s dbname=%s sslmode=disable",
		opts.User,
		opts.Password,
		strings.Split(opts.Addr, ":")[0],
		strings.Split(opts.Addr, ":")[1],
		opts.Database,
	)
	conn, err := sql.Open("postgres", source)
	if err != nil {
		return nil, err
	}

	return conn, nil
}
