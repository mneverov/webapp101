package config

import "time"

// Config represents a metric config.
type Config struct {
	Name             string    `json:"name"              pg:"name,pk"`
	URL              string    `json:"url"               pg:"url"`
	ScrapingInterval string    `json:"scraping_interval" pg:"scraping_interval"`
	DeletedAt        time.Time `json:"-"                 pg:"deleted_at,soft_delete"`
}

// nolint
type configStore interface {
	GetAll() ([]Config, error)
	Create(cfg Config) (Config, error)
	Get(name string) (Config, error)
	Update(cfg Config) (Config, error)
	Delete(name string) (Config, error)
}
