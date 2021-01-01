package scrape

import (
	"log"
	"time"
)

// producer is a wrapper over scraper which produce an infinite stream of
// scraping results.
type producer struct {
	name             string
	scraper          scraper
	scrapingInterval time.Duration
	stopCh           chan struct{}
	resCh            chan Result
}

// newProducer constructs a new producer.
// Result and stop channels will be instantiated.
func newProducer(name string, scraper scraper, scrapingInterval time.Duration) *producer {
	return &producer{
		name:             name,
		scraper:          scraper,
		scrapingInterval: scrapingInterval,
		stopCh:           make(chan struct{}),
		resCh:            make(chan Result),
	}
}

// run runs the producer routine: after every scraping interval a web page
// will be scraped and the result will be gathered and published to resCh.
// The routine is terminated by the producer stop channel.
func (p *producer) run() {
	t := time.NewTicker(p.scrapingInterval)
	defer t.Stop()

	for {
		select {
		case <-p.stopCh:
			log.Printf("shutdown producer %s\n", p.name)
			close(p.resCh)
			return
		case <-t.C:
			res, err := p.scraper.scrape()
			if err != nil {
				log.Printf("scrape failed: %+v\n", err)
				continue
			}
			p.resCh <- res
		}
	}
}
