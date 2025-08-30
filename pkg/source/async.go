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

func (s *AsyncSource) Build() error {
	return s.BuildAsync()
}

func (s *AsyncSource) BuildAsync() error {
	return s.BuildWithParamsAsync(DefaultBuildParams())
}

// Build encapsulates download and basic parsing
func (s *AsyncSource) BuildWithParams(params BuildParams) error {
	return s.BuildWithParamsAsync(params)
}
func (s *AsyncSource) BuildWithParamsAsync(params BuildParams) error {

	// Step 1: Download and parse homepage
	// if InputHTML is provided, use it instead of downloading
	if params.InputHTML != "" {
		s.HTML = params.InputHTML
	} else {
		err := s.Download()
		if err != nil {
			return fmt.Errorf("failed to download source: %v", err)
		}
	}
	err := s.Parse()
	if err != nil {
		return fmt.Errorf("failed to parse source: %v", err)
	}

	// Step 2: Set categories and feeds, download and parse them
	// if onlyHomepage is true, skip categories and feeds
	if params.OnlyHomepage {
		s.Categories = []newspaper.Category{{URL: s.URL, HTML: s.HTML, Doc: s.Doc}}
	} else {
		err := s.SearchCategories()
		if err != nil {
			return fmt.Errorf("failed to set categories: %v", err)
		}
		s.DownloadCategories()
	}
	s.BuildCategories()

	if len(s.Categories) > params.LimitCategories {
		s.Categories = s.Categories[:params.LimitCategories]
	}

	// Step 3: Download and parse feed
	// we skip feeds if onlyHomepage is true
	if !params.OnlyHomepage {
		s.GetFeedsWithParams(params)
	}

	return nil
}

func (s *AsyncSource) GetFeedsWithParams(params BuildParams) {
	s.GetFeedsWithParamsAsync(params)
}

// GetFeeds concurrently checks common feed URLs and feeds discovered in categories
func (s *AsyncSource) GetFeedsWithParamsAsync(params BuildParams) {
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
				if valid && err == nil {
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
		helpers.UniqueOptions{CaseSensitive: true, PreserveOrder: false},
	)

	s.Feeds = validFeeds
}

// DownloadCategories downloads HTML for all categories
func (s *AsyncSource) DownloadCategories() {
	s.DownloadCategoriesAsync()
}

func (s *AsyncSource) DownloadCategoriesAsync() {

	in := make(chan newspaper.Category, helpers.Min(len(s.Categories), s.Config.MaxWorkers))
	out := make(chan newspaper.Category, helpers.Min(len(s.Categories), s.Config.MaxWorkers))
	var wg sync.WaitGroup

	for i := 0; i < s.Config.MaxWorkers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for category := range in {
				err := s.downloadCategory(&category)
				if err == nil {
					out <- category
				}
			}
		}()
	}

	// feeder
	go func() {
		for _, u := range s.Categories {
			in <- u
		}
		close(in)
	}()

	// collector
	var collectorWg sync.WaitGroup
	categoriesCollected := []newspaper.Category{}
	collectorWg.Add(1)
	go func() {
		defer collectorWg.Done()
		for f := range out {
			categoriesCollected = append(categoriesCollected, f)
		}
	}()

	// wait for workers to finish then close out
	go func() {
		wg.Wait()
		close(out)
	}()
	collectorWg.Wait()

	validCategories := helpers.UniqueStructByKey(
		categoriesCollected,
		func(f newspaper.Category) string { return f.URL },
		helpers.UniqueOptions{CaseSensitive: true, PreserveOrder: false},
	)

	s.Categories = validCategories

}
