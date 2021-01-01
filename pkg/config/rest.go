package config

import (
	"encoding/json"
	"log"
	"net/http"
)

// Handler represents a config handler.
type Handler struct {
	service configService
}

// NewHandler creates a new config handler.
func NewHandler(service configService) *Handler {
	return &Handler{service: service}
}

// GetAll returns a list of all configs.
// GET /configs.
func (h *Handler) GetAll(w http.ResponseWriter, _ *http.Request) {
	configs, err := h.service.GetAll()
	if err != nil {
		log.Printf("%+v\n", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	output, err := json.Marshal(configs)
	if err != nil {
		log.Printf("failed to marshal response %+v\n", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("content-type", "application/json")
	w.WriteHeader(http.StatusOK)
	_, err = w.Write(output)
	if err != nil {
		log.Printf("failed to write response: %+v\n", err)
	}
}

// Create creates a config from request. Corresponding scrapper will also be
// created and started.
// POST /configs.
func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
	cfg := Config{}
	err := json.NewDecoder(r.Body).Decode(&cfg)
	if err != nil {
		log.Printf("failed to decode config %+v\n", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	_, err = h.service.Create(cfg)
	if err != nil {
		log.Printf("%+v\n", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("content-type", "application/json")
	w.WriteHeader(http.StatusOK)
}
