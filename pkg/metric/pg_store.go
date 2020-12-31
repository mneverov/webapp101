package metric

import (
	"context"
	"log"

	"github.com/go-pg/pg/v10"
	"github.com/go-pg/pg/v10/orm"
	"github.com/pkg/errors"
)

// Postgres implements store interface, provides interaction with PG DB for
// simple CRUD operations for metrics.
type Postgres struct {
	db *pg.DB
}

// NewPostgresStorage creates a new instance of the Postgres Storage.
func NewPostgresStorage(opts *pg.Options) (*Postgres, error) {
	db := pg.Connect(opts)

	if err := db.Ping(context.Background()); err != nil {
		_ = db.Close()
		return nil, errors.Wrapf(err, "failed to ping metric DB")
	}

	log.Printf("metric: connected to \"postgres://%s:***@%s/%s\" sslmode enabled=%t\n",
		opts.User, opts.Addr, opts.Database, opts.TLSConfig != nil)

	return &Postgres{db: db}, nil
}

// Create creates a new metric config.
func (s *Postgres) Create(metric Metric) (Metric, error) {
	_, err := s.db.Model(&metric).
		Returning("*").
		Insert()
	if err != nil {
		return Metric{},
			errors.Wrapf(err, "failed to create metric %s", metric.Name)
	}

	return metric, nil
}

// Get returns metrics that satisfy given filter.
func (s *Postgres) Get(filter Filter) (Metrics, error) {
	metrics := Metrics{}

	err := s.db.Model(&metrics).
		Relation("Metrics", func(q *orm.Query) (*orm.Query, error) {
			q.Where("created_at >= ?", filter.Since)
			return q, nil
		}).
		Where("cfg.name = ?", filter.Name).
		Select(&metrics)
	if err != nil {
		return Metrics{}, errors.Wrapf(err, "failed to get metrics %s", filter)
	}

	return metrics, nil
}
