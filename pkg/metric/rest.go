package metric

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"time"
)

// Handler represents a metric handler.
type Handler struct {
	service metricService
}

// NewHandler creates a new metric handler.
func NewHandler(service metricService) *Handler {
	return &Handler{service: service}
}

// Get returns a list of metrics filtered by given query parameters.
// GET /metrics?name=metricName&since=scrapeInterval.
func (h *Handler) Get(w http.ResponseWriter, r *http.Request) {
	name := r.URL.Query().Get("name")
	since := r.URL.Query().Get("since")

	err := validateQuery(name, since, r.URL.Query())
	if err != nil {
		log.Printf("%+v\n", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	timestamp, _ := time.Parse(time.RFC3339, since)
	metrics, err := h.service.Get(Filter{Name: name, Since: timestamp})
	if err != nil {
		log.Printf("%+v\n", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	output, err := json.Marshal(metrics)
	if err != nil {
		log.Printf("failed to marshal response %+v\n", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("content-type", "application/json")
	w.WriteHeader(http.StatusOK)
	_, err = w.Write(output)
	if err != nil {
		log.Printf("failed to write response: %+v", err)
	}
}

func validateQuery(name, timestamp string, q url.Values) error {
	if len(name) == 0 || len(timestamp) == 0 {
		return fmt.Errorf(
			"both name and timestamp must be present, was [%v]",
			q.Encode(),
		)
	}

	_, err := time.Parse(time.RFC3339, timestamp)
	if err != nil {
		return fmt.Errorf("invalid timestamp format: %s", timestamp)
	}
	return nil
}
