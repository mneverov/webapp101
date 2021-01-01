package metric

//go:generate mockery --inpackage --all --case=underscore

import (
	"log"
	"time"

	"github.com/mneverov/webapp101/pkg/scrape"
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
	Data []Metric `json:"data"`
}

// Filter contains a set of parameters to filter metrics.
type Filter struct {
	Name  string
	Since time.Time
}

type metricStore interface {
	Create(metric Metric) (Metric, error)
	Get(filter Filter) ([]Metric, error)
}

type metricService interface {
	Get(f Filter) (Metrics, error)
	Consume(name string, resCh <-chan scrape.Result)
}

// Service provides methods to work with Metrics.
type Service struct {
	store metricStore
}

// NewService creates a new metric service.
func NewService(store metricStore) *Service {
	return &Service{store: store}
}

// Get returns metrics that satisfy given filter, or empty Metrics if no
// metrics found.
func (s *Service) Get(f Filter) (Metrics, error) {
	metrics, err := s.store.Get(f)
	if err != nil {
		return Metrics{}, err
	}
	return Metrics{Data: metrics}, nil
}

// Consume runs infinite loop to consume all the results from the given channel.
// Consume exits on result channel close.
func (s *Service) Consume(name string, resCh <-chan scrape.Result) {
	// iterate through the resCh.
	for r := range resCh {
		// on each result: assemble Metric
		m := Metric{
			Name:              name,
			StatusCode:        r.StatusCode,
			ResponseSizeBytes: r.ResponseSizeBytes,
			ResponseTimeMs:    r.ResponseTimeMs,
			CreatedAt:         r.CreatedAt,
		}
		// store it in DB
		_, err := s.store.Create(m)
		if err != nil {
			log.Printf("failed to store metric %#v %+v", m, err)
		}
	}
}
