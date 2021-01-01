package scrape

import (
	"fmt"
	"net/http"
	"time"
)

type httpClient interface {
	Do(req *http.Request) (*http.Response, error)
}

// InMemoryManager provides methods to manage scrapers in memory.
type InMemoryManager struct {
	producers map[string]*producer
	client    httpClient
}

// NewInMemoryManager creates a new InMemoryManager.
func NewInMemoryManager(client httpClient) *InMemoryManager {
	return &InMemoryManager{
		producers: make(map[string]*producer),
		client:    client,
	}
}

// Run creates a new scraper and runs the scraping routine.
// NOT THREAD SAFE!
func (m *InMemoryManager) Run(
	name, url string, scrapeInterval time.Duration,
) (<-chan Result, error) {
	_, exists := m.producers[name]
	if exists {
		return nil, fmt.Errorf("scraper %s does already exist", name)
	}

	s := newHTTPScraper(m.client, url)
	p := newProducer(name, s, scrapeInterval)
	m.producers[name] = p

	go p.run()

	return p.resCh, nil
}

// Update updates the scraper associated with the given name.
// NOT THREAD SAFE!
func (m *InMemoryManager) Update(
	name, url string, scrapeInterval time.Duration,
) (<-chan Result, error) {
	p, exists := m.producers[name]
	if !exists {
		return nil, fmt.Errorf("scraper %s does not exist", name)
	}

	// stop and delete the existing scraper
	delete(m.producers, "name")
	p.stopCh <- struct{}{}
	// create a new scraper
	s := newHTTPScraper(m.client, url)
	p = newProducer(name, s, scrapeInterval)
	m.producers[name] = p

	go p.run()

	return p.resCh, nil
}

// Stop stops the scraper associated with the given name and removes it
// from the list of scrapers.
// NOT THREAD SAFE!
func (m *InMemoryManager) Stop(name string) error {
	p, exists := m.producers[name]
	if !exists {
		return fmt.Errorf("scraper %s does not exist", name)
	}

	delete(m.producers, "name")
	p.stopCh <- struct{}{}
	return nil
}
