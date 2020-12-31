package metric

import (
	"time"

	"github.com/mneverov/webapp101/pkg/config"
)

// Metric represents a single web page metric, gathered with a scraper.
type Metric struct {
	ID                int       `json:"-"                    pg:"id,pk"`
	Name              string    `json:"-"                    pg:"name,use_zero"`
	StatusCode        int       `json:"status_code"          pg:"status_code,use_zero"`
	ResponseSizeBytes int64     `json:"response_size_bytes"  pg:"response_size,use_zero"`
	ResponseTimeMs    int       `json:"response_time_ms"     pg:"response_time,use_zero"`
	CreatedAt         time.Time `json:"created_at"           pg:"created_at"`
}

// Metrics represents a collection of metrics for a web page defined in the
// config for some period of time.
type Metrics struct {
	tableName struct{} `pg:"configs,alias:cfg"` // nolint // tableName is used in go-pg internally
	config.Config
	Metrics []Metric `json:"data" pg:"fk:name,rel:has-many"`
}

// Filter contains a set of parameters to filter metrics.
type Filter struct {
	Name  string
	Since time.Time
}

// nolint
type metricStore interface {
	Create(metric Metric) (Metric, error)
	Get(filter Filter) (Metrics, error)
}
