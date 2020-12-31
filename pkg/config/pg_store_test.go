package config

import (
	"database/sql"
	"fmt"
	"strings"
	"testing"

	"github.com/go-pg/migrations/v8"
	"github.com/go-pg/pg/v10"
	"github.com/go-testfixtures/testfixtures/v3"
	_ "github.com/lib/pq"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var (
	testCfg = Config{
		Name:             "test_cfg",
		URL:              "test_url",
		ScrapingInterval: "42s",
	}
	exampleCfg = Config{
		Name:             "example",
		URL:              "http://example.com/",
		ScrapingInterval: "5s",
	}
)

func TestGetAll(t *testing.T) {
	db := testDB(t)
	t.Run("should return found configs", func(t *testing.T) {
		res, err := db.GetAll()
		assert.NoError(t, err)
		assert.Len(t, res, 2)
		assert.Contains(t, res, exampleCfg)
	})

	t.Run("should return empty slice when no configs found", func(t *testing.T) {
		_, err := db.Delete("github_jobs")
		require.NoError(t, err)
		_, err = db.Delete("example")
		require.NoError(t, err)

		res, err := db.GetAll()
		assert.NoError(t, err)
		assert.NotNil(t, res)
		assert.Len(t, res, 0)
	})
}

func TestCreate(t *testing.T) {
	db := testDB(t)
	t.Run("should return error when failed to create config (duplicated name)", func(t *testing.T) {
		_, err := db.Create(exampleCfg)
		require.Error(t, err)
		assert.Regexp(t, exampleCfg.Name, err)
	})

	t.Run("should return created config", func(t *testing.T) {
		res, err := db.Create(testCfg)
		assert.NoError(t, err)
		assert.Equal(t, testCfg, res)
	})
}

func TestGet(t *testing.T) {
	db := testDB(t)
	t.Run("should return error when no config found", func(t *testing.T) {
		_, err := db.Get("non_existing_config")
		require.Error(t, err)
		assert.Regexp(t, "non_existing_config", err)
	})

	t.Run("should return config with given name", func(t *testing.T) {
		res, err := db.Get(exampleCfg.Name)
		assert.NoError(t, err)
		assert.Equal(t, exampleCfg, res)
	})
}

func TestUpdate(t *testing.T) {
	db := testDB(t)
	t.Run("should return error when no config found", func(t *testing.T) {
		_, err := db.Update(testCfg)
		require.Error(t, err)
		assert.Regexp(t, testCfg.Name, err)
	})

	t.Run("should return updated config", func(t *testing.T) {
		expectedCfg := exampleCfg
		expectedCfg.ScrapingInterval = "100500s"
		res, err := db.Update(expectedCfg)
		assert.NoError(t, err)
		assert.Equal(t, expectedCfg, res)
	})
}

func TestDelete(t *testing.T) {
	db := testDB(t)
	t.Run("should return no error when no config found", func(t *testing.T) {
		_, err := db.Delete("non_existing_config")

		require.Error(t, err)
		assert.Regexp(t, "no rows", err)
	})

	t.Run("should return deleted config", func(t *testing.T) {
		res, err := db.Delete("example")
		assert.NoError(t, err)
		assert.Equal(t, "example", res.Name)

		_, err = db.Get("example")
		require.Error(t, err)
		assert.Regexp(t, "no rows", err)
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
		t.Fatalf("config: failed to connect %+v\n", err)
	}
	err = loadFixtures(fixturesDB)
	if err != nil {
		t.Fatalf("config: failed to load fixtures %+v\n", err)
	}

	return db
}

func migrate(t *testing.T, opts *pg.Options) {
	col := migrations.NewCollection()
	err := col.DiscoverSQLMigrations("../../resources/migrations")
	if err != nil {
		t.Fatalf("config: failed to discover migrations %+v\n", err)
	}
	// need to have *pg.DB, not our defined PostgresStorage
	migrationsDB := pg.Connect(opts)
	defer migrationsDB.Close()

	// it is mandatory to run init before migrations:
	// see https://github.com/go-pg/migrations/issues/48
	_, _, err = col.Run(migrationsDB, "init")
	if err != nil {
		t.Fatalf("config: failed to init migrations %+v\n", err)
	}

	oldVersion, newVersion, err := col.Run(migrationsDB, "reset")
	if err != nil {
		t.Fatalf("config: failed to reset migrations %+v\n", err)
	}
	t.Logf("config: reset migrations from %d to %d \n", oldVersion, newVersion)

	// even if Run accepts vararg of commands, all should run separately
	oldVersion, newVersion, err = col.Run(migrationsDB, "up")
	if err != nil {
		t.Fatalf("config: failed to apply migrations %+v\n", err)
	}
	t.Logf("config: apply migrations from %d to %d \n", oldVersion, newVersion)
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
