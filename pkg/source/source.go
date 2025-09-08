package source

import (
	"github.com/tguidoux/newspaper4k-go/pkg/configuration"
)

// Source interface defines the methods for a news source
type Source interface {
	Build() error
	Download() error
	Parse() error
	SearchCategories() error
	GetFeeds()
	DownloadCategories()
	BuildCategories()
	GetArticles()
	Size() int
}

type SourceRequest struct {
	URL    string
	Config configuration.Configuration
}

type BuildParams struct {
	InputHTML       string
	OnlyHomepage    bool
	OnlySameDomain  bool
	AllowSubDomain  bool
	LimitCategories int
	LimitArticles   int
	Shuffle         bool
}

func DefaultBuildParams() BuildParams {
	return BuildParams{
		InputHTML:       "",
		OnlyHomepage:    false,
		OnlySameDomain:  false,
		AllowSubDomain:  true,
		LimitCategories: 100,
		LimitArticles:   1000,
		Shuffle:         false,
	}
}
