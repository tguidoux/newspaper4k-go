package configuration

import (
	"errors"
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
	Extractors           []any
	DownloadOptions      DownloadOptions
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
		RequestsParams:       RequestsParams{Timeout: 7, Proxies: map[string]string{}, Headers: map[string]string{"User-Agent": "newspaper/0.0.1"}},
		NumberThreads:        10,
		Verbose:              false,
		ThreadTimeoutSeconds: 10,
		AllowBinaryContent:   false,
		IgnoredContentTypes:  map[string]string{},
		UseCachedCategories:  true,
		Extractors:           []any{},
		DownloadOptions:      DownloadOptions{InputHTML: ""},
	}
}

// Getters and setters for key properties
func (c *Configuration) BrowserUserAgent() string {
	return c.RequestsParams.Headers["User-Agent"]
}

func (c *Configuration) SetBrowserUserAgent(val string) {
	c.RequestsParams.Headers["User-Agent"] = val
}

func (c *Configuration) Headers() map[string]string {
	return c.RequestsParams.Headers
}

func (c *Configuration) SetHeaders(val map[string]string) {
	c.RequestsParams.Headers = val
}

func (c *Configuration) RequestTimeout() int {
	return c.RequestsParams.Timeout
}

func (c *Configuration) SetRequestTimeout(val int) {
	c.RequestsParams.Timeout = val
}

func (c *Configuration) Proxies() map[string]string {
	return c.RequestsParams.Proxies
}

func (c *Configuration) SetProxies(val map[string]string) {
	c.RequestsParams.Proxies = val
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
