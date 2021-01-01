package config

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/mneverov/webapp101/pkg/scrape"
)

var testScrapingInterval = 42 * time.Second

func TestConfigService_GetAll(t *testing.T) {
	t.Run("should propagate error from DB", func(t *testing.T) {
		ts := createTestServices()
		ts.db.On("GetAll").
			Return(nil, assert.AnError).
			Once()

		_, err := ts.cfgService.GetAll()
		require.Error(t, err)
		assert.Equal(t, assert.AnError, err)
		ts.db.AssertExpectations(t)
	})

	t.Run("should return configs from DB", func(t *testing.T) {
		ts := createTestServices()
		ts.db.On("GetAll").
			Return([]Config{testCfg}, nil).
			Once()

		res, err := ts.cfgService.GetAll()
		assert.NoError(t, err)
		assert.Len(t, res.Data, 1)
		assert.Equal(t, testCfg, res.Data[0])
		ts.db.AssertExpectations(t)
	})
}

func TestConfigService_Create(t *testing.T) {
	t.Run("should return error when interval is invalid", func(t *testing.T) {
		ts := createTestServices()
		testCfg := testCfg
		testCfg.ScrapingInterval = "invalid_duration"

		_, err := ts.cfgService.Create(testCfg)

		require.Error(t, err)
		assert.Regexp(t, testCfg.ScrapingInterval, err)
	})

	t.Run("should return error on DB failure", func(t *testing.T) {
		ts := createTestServices()
		cfg := testCfg
		ts.db.On("Create", cfg).
			Return(Config{}, assert.AnError).
			Once()

		_, err := ts.cfgService.Create(cfg)

		require.Error(t, err)
		assert.Equal(t, assert.AnError, err)
		ts.db.AssertExpectations(t)
	})

	t.Run("should return error when scraper manager fails", func(t *testing.T) {
		ts := createTestServices()
		cfg := testCfg

		ts.db.On("Create", cfg).
			Return(cfg, nil).
			Once()

		ts.scraperManager.
			On("Run", cfg.Name, cfg.URL, testScrapingInterval).
			Return(nil, assert.AnError).
			Once()

		_, err := ts.cfgService.Create(cfg)

		require.Error(t, err)
		assert.Equal(t, assert.AnError, err)
		ts.scraperManager.AssertExpectations(t)
		ts.db.AssertExpectations(t)
	})

	t.Run("should return created config", func(t *testing.T) {
		ts := createTestServices()
		cfg := testCfg
		ch := make(<-chan scrape.Result)

		ts.db.On("Create", cfg).
			Return(cfg, nil).
			Once()

		ts.scraperManager.
			On("Run", cfg.Name, cfg.URL, testScrapingInterval).
			Return(ch, nil).
			Once()

		ts.metricService.On("Consume", cfg.Name, ch).Return()

		res, err := ts.cfgService.Create(cfg)

		assert.NoError(t, err)
		assert.Equal(t, cfg, res)
		ts.scraperManager.AssertExpectations(t)
		ts.db.AssertExpectations(t)
		// no assert for metric service since the Consume operation is async:
		// the test might end sooner than the goroutine is run.
		// ts.metricService.AssertExpectations(t)
	})
}

// TestConfigService_Update only tests happy path. The rest of the tests may
// be added by participants.
func TestConfigService_Update(t *testing.T) {
	t.Run("should update existing config", func(t *testing.T) {
		ts := createTestServices()
		cfg := testCfg
		ch := make(<-chan scrape.Result)

		ts.db.On("Update", cfg).
			Return(cfg, nil).
			Once()

		ts.scraperManager.
			On("Update", cfg.Name, cfg.URL, testScrapingInterval).
			Return(ch, nil).
			Once()

		ts.metricService.On("Consume", cfg.Name, ch).Return()

		err := ts.cfgService.Update(cfg)

		assert.NoError(t, err)
		ts.scraperManager.AssertExpectations(t)
		ts.db.AssertExpectations(t)
		// no assert for metric service since the Consume operation is async:
		// the test might end sooner than the goroutine is run.
		// ts.metricService.AssertExpectations(t)
	})
}

type ts struct {
	cfgService     *Service
	metricService  *mockMetricService
	scraperManager *mockScraperManager
	db             *mockConfigStore
}

func createTestServices() *ts {
	db := &mockConfigStore{}
	scraperManager := &mockScraperManager{}
	metricService := &mockMetricService{}
	cfgService := NewService(db, metricService, scraperManager)

	return &ts{
		cfgService:     cfgService,
		metricService:  metricService,
		scraperManager: scraperManager,
		db:             db,
	}
}
