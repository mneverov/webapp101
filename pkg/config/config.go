package config

//go:generate mockery --inpackage --all --case=underscore

import (
	"log"
	"time"

	"github.com/pkg/errors"

	"github.com/mneverov/webapp101/pkg/scrape"
)

// Config represents a metric config.
type Config struct {
	Name             string    `json:"name"              pg:"name,pk"`
	URL              string    `json:"url"               pg:"url,use_zero"`
	ScrapingInterval string    `json:"scraping_interval" pg:"scraping_interval,use_zero"`
	DeletedAt        time.Time `json:"-"                 pg:"deleted_at,soft_delete"`
}

// Configs contains a collection of configs.
type Configs struct {
	Data []Config `json:"data"`
}

type configStore interface {
	GetAll() ([]Config, error)
	Create(cfg Config) (Config, error)
	Get(name string) (Config, error)
	Update(cfg Config) (Config, error)
	Delete(name string) (Config, error)
}

// nolint
type configService interface {
	GetAll() (Configs, error)
	Create(cfg Config) (Config, error)
	Get(name string) (Config, error)
	Update(cfg Config) error
	Delete(name string) error
}

type metricService interface {
	Consume(name string, resCh <-chan scrape.Result)
}

type scraperManager interface {
	Run(name, url string, scrapePeriod time.Duration) (<-chan scrape.Result, error)
	Update(name, url string, scrapePeriod time.Duration) (<-chan scrape.Result, error)
	Stop(name string) error
}

// Service provides methods to work with Configs.
type Service struct {
	store          configStore
	metricService  metricService
	scraperManager scraperManager
}

// NewService creates a new config service.
func NewService(
	store configStore, metricService metricService, scraperManager scraperManager,
) *Service {
	return &Service{
		store:          store,
		metricService:  metricService,
		scraperManager: scraperManager,
	}
}

// GetAll returns all existing configs.
func (s *Service) GetAll() (Configs, error) {
	configs, err := s.store.GetAll()
	if err != nil {
		return Configs{}, err
	}
	return Configs{Data: configs}, err
}

// Create creates a new config.
func (s *Service) Create(cfg Config) (Config, error) {
	duration, err := time.ParseDuration(cfg.ScrapingInterval)
	if err != nil {
		return Config{},
			errors.Wrapf(
				err, "failed to parse scraping interval %q",
				cfg.ScrapingInterval,
			)
	}

	cfg, err = s.store.Create(cfg)
	if err != nil {
		return cfg, err
	}

	resCh, err := s.scraperManager.Run(cfg.Name, cfg.URL, duration)
	if err != nil {
		return Config{}, err
	}
	go s.metricService.Consume(cfg.Name, resCh)
	return cfg, nil
}

// Get returns a config with the given name.
func (s *Service) Get(name string) (Config, error) {
	cfg, err := s.store.Get(name)
	if err != nil {
		return Config{}, err
	}
	return cfg, nil
}

// Update updates a config with the given name.
func (s *Service) Update(cfg Config) error {
	duration, err := time.ParseDuration(cfg.ScrapingInterval)
	if err != nil {
		return errors.Wrapf(
			err, "failed to parse scraping interval %q", cfg.ScrapingInterval,
		)
	}

	cfg, err = s.store.Update(cfg)
	if err != nil {
		return err
	}

	resCh, err := s.scraperManager.Update(cfg.Name, cfg.URL, duration)
	if err != nil {
		return err
	}
	go s.metricService.Consume(cfg.Name, resCh)
	return nil
}

// Delete deletes a config with the given name.
func (s *Service) Delete(name string) error {
	_, err := s.store.Delete(name)
	if err != nil {
		return err
	}

	err = s.scraperManager.Stop(name)
	if err != nil {
		// the scraper has already been deleted - log and proceed.
		log.Printf("%+v\n", err)
	}

	return nil
}
