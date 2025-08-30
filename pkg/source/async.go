package source

import (
	"fmt"
	"sync"

	"github.com/tguidoux/newspaper4k-go/internal/helpers"
	"github.com/tguidoux/newspaper4k-go/internal/urls"
	"github.com/tguidoux/newspaper4k-go/pkg/newspaper"
)

// AsyncSource is like DefaultSource but performs feed checks and feed parsing concurrently
type AsyncSource struct {
	*DefaultSource
	// Prepare for future additional params
}

// NewAsyncSource creates a new AsyncSource by reusing NewDefaultSource
func NewAsyncSource(request SourceRequest) (*AsyncSource, error) {
	ds, err := NewDefaultSource(request)
	if err != nil {
		return nil, fmt.Errorf("failed to create default source: %v", err)
	}
	return &AsyncSource{DefaultSource: ds}, nil
}

// GetFeeds concurrently checks common feed URLs and feeds discovered in categories
func (s *AsyncSource) GetFeeds() {

	commonFeedURLs := s.getCommonFeeds()

	in := make(chan string, helpers.Min(len(commonFeedURLs), s.Config.MaxWorkers))
	out := make(chan newspaper.Feed, helpers.Min(len(commonFeedURLs), s.Config.MaxWorkers))
	var wg sync.WaitGroup

	for i := 0; i < s.Config.MaxWorkers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for feedURL := range in {
				url := urls.PrepareURL(feedURL, feedURL)
				rss, valid, err := s.checkFeed(url)
				if valid && err != nil {
					out <- newspaper.Feed{URL: url, RSS: rss}
				}
			}
		}()
	}

	// feeder
	go func() {
		for _, u := range commonFeedURLs {
			in <- u
		}
		close(in)
	}()

	// collector
	var collectorWg sync.WaitGroup
	feedsCollected := []newspaper.Feed{}
	collectorWg.Add(1)
	go func() {
		defer collectorWg.Done()
		for f := range out {
			feedsCollected = append(feedsCollected, f)
		}
	}()

	// wait for workers to finish then close out
	go func() {
		wg.Wait()
		close(out)
	}()
	collectorWg.Wait()

	// Extract feed URLs from categories (s.extractFeedURLs is promoted from DefaultSource)
	feedURLs := s.extractFeedURLs(s.Categories)
	for _, fu := range feedURLs {
		feedsCollected = append(feedsCollected, newspaper.Feed{URL: urls.PrepareURL(fu, fu)})
	}

	validFeeds := helpers.UniqueStructByKey(
		feedsCollected,
		func(f newspaper.Feed) string { return f.URL },
		helpers.UniqueOptions{CaseSensitive: true, PreserveOrder: true},
	)

	s.Feeds = validFeeds
}
