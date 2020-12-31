package metric

import (
	"github.com/go-pg/pg/v10"
	"github.com/pkg/errors"
)

// Postgres implements store interface, provides interaction with PG DB for
// simple CRUD operations for metrics.
type Postgres struct {
	db *pg.DB
}

// NewPostgresStorage creates a new instance of the Postgres Storage.
func NewPostgresStorage(db *pg.DB) *Postgres {
	return &Postgres{db: db}
}

// Create creates a new metric config.
func (s *Postgres) Create(metric Metric) (Metric, error) {
	_, err := s.db.Model(&metric).
		Returning("*").
		Insert()

	if err != nil {
		return Metric{},
			errors.Wrapf(err, "failed to store metric %s", metric.Name)
	}

	return metric, nil
}

// Get returns metrics that satisfy given filter.
func (s *Postgres) Get(filter Filter) ([]Metric, error) {
	metrics := make([]Metric, 0)

	err := s.db.Model(&metrics).
		Where("created_at >= ?", filter.Since).
		Where("name = ?", filter.Name).
		Select(&metrics)

	if err != nil {
		return nil, errors.Wrapf(err, "failed to get metrics %s", filter)
	}
	return metrics, nil
}
