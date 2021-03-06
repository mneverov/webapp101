package config

import (
	"github.com/go-pg/pg/v10"
	"github.com/pkg/errors"
)

// Postgres provides interaction with Postgresql DB for simple CRUD operations
// for metrics configs.
type Postgres struct {
	db *pg.DB
}

// NewPostgresStorage creates a new instance of the Postgres Storage.
func NewPostgresStorage(db *pg.DB) *Postgres {
	return &Postgres{db: db}
}

// GetAll returns all existing configs.
func (s *Postgres) GetAll() ([]Config, error) {
	configs := make([]Config, 0)
	err := s.db.Model(&configs).Select()
	if err != nil {
		return nil, errors.Wrapf(err, "failed to get all configs")
	}
	return configs, err
}

/*
GetAllPlainDriver provides an example how to work with plain PG driver.
Note, that in this example *sql.DB is used instead of *pg.DB

func (s *Postgres) GetAllPlainDriver() ([]Config, error) {
	rows, err := s.db.Conn().Query("select name, url, interval from configs")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	configs := make([]Config, 0)
	for rows.Next() {
		var cfg Config
		err := rows.Scan(&cfg.Name, &cfg.URL, &cfg.ScrapingInterval)
		if err != nil {
			return nil, err
		}
		configs = append(configs, cfg)
	}
	if err = rows.Close(); err != nil {
		return nil, err
	}
	return configs, nil
}
*/

// Create creates a new config.
func (s *Postgres) Create(cfg Config) (Config, error) {
	_, err := s.db.Model(&cfg).
		Returning("*").
		Insert()
	if err != nil {
		return Config{},
			errors.Wrapf(err, "failed to create config %s", cfg.Name)
	}

	return cfg, nil
}

// Get returns a config with the given name.
func (s *Postgres) Get(name string) (Config, error) {
	cfg := Config{Name: name}
	err := s.db.Model(&cfg).WherePK().Select()
	if err != nil {
		return Config{}, errors.Wrapf(err, "failed to get config %s", name)
	}
	return cfg, nil
}

// Update updates a config with the given name.
func (s *Postgres) Update(cfg Config) (Config, error) {
	_, err := s.db.Model(&cfg).WherePK().
		Returning("*").
		Update()
	if err != nil {
		return Config{},
			errors.Wrapf(err, "failed to update config %s", cfg.Name)
	}
	return cfg, nil
}

// Delete deletes a config with the given name.
func (s *Postgres) Delete(name string) (Config, error) {
	cfg := Config{Name: name}
	_, err := s.db.Model(&cfg).WherePK().
		Returning("*").
		Delete()
	if err != nil {
		return Config{},
			errors.Wrapf(err, "failed to delete config %s", cfg.Name)
	}

	return cfg, nil
}
