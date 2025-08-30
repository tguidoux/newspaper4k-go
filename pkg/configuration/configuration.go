package configuration

import (
	"errors"
	"fmt"

	newspaper4kgo "github.com/tguidoux/newspaper4k-go"
)

// DownloadOptions contains options for downloading an article
type DownloadOptions struct {
	InputHTML string // If provided, use this HTML instead of downloading
}

// Configuration holds settings for Article/Source objects.
type Configuration struct {
	MinWordCount         int
	MinSentCount         int
	MaxTitle             int
	MaxText              int
	MaxKeywords          int
	MaxAuthors           int
	MaxSummary           int
	MaxSummarySent       int
	MaxFileMemo          int
	MaxWorkers           int
	TopImageSettings     TopImageSettings
	MemorizeArticles     bool
	DisableCategoryCache bool
	FetchImages          bool
	FollowMetaRefresh    bool
	UseMetaLanguage      bool
	CleanArticleHTML     bool
	HTTPSuccessOnly      bool
	language             string
	RequestsParams       RequestsParams
	NumberThreads        int
	Verbose              bool
	ThreadTimeoutSeconds int
	AllowBinaryContent   bool
	IgnoredContentTypes  map[string]string
	UseCachedCategories  bool
	DownloadOptions      DownloadOptions
	MaxFeeds             int
}

// TopImageSettings holds settings for finding top image.
type TopImageSettings struct {
	MinWidth   int
	MinHeight  int
	MinArea    int
	MaxRetries int
}

// RequestsParams holds HTTP request parameters.
type RequestsParams struct {
	Timeout int
	Proxies map[string]string
	Headers map[string]string
}

// NewConfiguration returns a Configuration with default values.
func NewConfiguration() *Configuration {
	return &Configuration{
		MinWordCount:         300,
		MinSentCount:         7,
		MaxTitle:             200,
		MaxText:              100000,
		MaxKeywords:          35,
		MaxWorkers:           20,
		MaxFeeds:             100,
		MaxAuthors:           10,
		MaxSummary:           5000,
		MaxSummarySent:       5,
		MaxFileMemo:          20000,
		TopImageSettings:     TopImageSettings{MinWidth: 300, MinHeight: 200, MinArea: 10000, MaxRetries: 2},
		MemorizeArticles:     true,
		DisableCategoryCache: false,
		FetchImages:          true,
		FollowMetaRefresh:    false,
		UseMetaLanguage:      true,
		CleanArticleHTML:     true,
		HTTPSuccessOnly:      true,
		language:             "",
		RequestsParams:       RequestsParams{Timeout: 30, Proxies: map[string]string{}, Headers: map[string]string{"User-Agent": fmt.Sprintf("newspaper4k-go/%s", newspaper4kgo.Version)}},
		NumberThreads:        10,
		Verbose:              false,
		ThreadTimeoutSeconds: 10,
		AllowBinaryContent:   false,
		IgnoredContentTypes:  map[string]string{},
		UseCachedCategories:  true,
		DownloadOptions:      DownloadOptions{InputHTML: ""},
	}
}

func (c *Configuration) Language() string {
	return c.language
}

func (c *Configuration) SetLanguage(val string) error {
	if val == "" || len(val) != 2 {
		return errors.New("language must be a 2 char code, e.g. 'en', 'de'")
	}
	c.language = val
	c.UseMetaLanguage = false
	return nil
}
