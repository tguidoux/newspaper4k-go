package source

import (
	"github.com/tguidoux/newspaper4k-go/pkg/configuration"
	"github.com/tguidoux/newspaper4k-go/pkg/newspaper"
)

// Source interface defines the methods for a news source
type Source interface {
	Build(inputHTML string, onlyHomepage bool, onlyInPath bool)
	Download()
	Parse()
	SearchCategories()
	GetFeeds()
	getCommonFeeds() []string
	checkFeed(feedURL string) (string, bool, error)
	extractDescription()
	DownloadCategories()
	BuildCategories()
	feedsToArticles() []newspaper.Article
	categoriesToArticles() []newspaper.Article
	GetArticles(limit int, onlyInPath bool)
	Size() int
	FeedURLs() []string
	CategoryURLs() []string
	ArticleURLs() []string
	PrintSummary()
	String() string
}

type SourceRequest struct {
	URL    string
	Config configuration.Configuration
}

type BuildParams struct {
	InputHTML       string
	OnlyHomepage    bool
	OnlyInPath      bool
	LimitCategories int
	LimitArticles   int
}
