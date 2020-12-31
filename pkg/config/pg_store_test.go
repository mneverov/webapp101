package config

import (
	"testing"

	_ "github.com/lib/pq"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/mneverov/webapp101/pkg/testutil"
)

func TestConfigDB_GetAll(t *testing.T) {
	conn := testutil.TestDB(t, dbOpts, "config")
	db := NewPostgresStorage(conn)
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

func TestConfigDB_Create(t *testing.T) {
	conn := testutil.TestDB(t, dbOpts, "config")
	db := NewPostgresStorage(conn)
	t.Run(
		"should return error when failed to create config (duplicated name)",
		func(t *testing.T) {
			_, err := db.Create(exampleCfg)
			require.Error(t, err)
			assert.Regexp(t, exampleCfg.Name, err)
		},
	)

	t.Run("should return created config", func(t *testing.T) {
		res, err := db.Create(testCfg)
		assert.NoError(t, err)
		assert.Equal(t, testCfg, res)
	})
}

func TestConfigDB_Get(t *testing.T) {
	conn := testutil.TestDB(t, dbOpts, "config")
	db := NewPostgresStorage(conn)
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

func TestConfigDB_Update(t *testing.T) {
	conn := testutil.TestDB(t, dbOpts, "config")
	db := NewPostgresStorage(conn)
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

func TestConfigDB_Delete(t *testing.T) {
	conn := testutil.TestDB(t, dbOpts, "config")
	db := NewPostgresStorage(conn)
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
