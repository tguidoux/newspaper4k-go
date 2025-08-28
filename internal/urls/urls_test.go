package urls

import (
	"testing"
)

func TestParse(t *testing.T) {
	tests := []struct {
		name          string
		urlStr        string
		wantScheme    string
		wantDomain    string
		wantSubdomain string
		wantTLD       string
		wantPort      string
		wantPath      string
		wantRawQuery  string
		wantFragment  string
		wantFileType  string
		wantErr       bool
	}{
		{
			name:          "CNN article URL",
			urlStr:        "https://edition.cnn.com/2023/10/01/world/sample-article/index.html?utm_source=newsletter",
			wantScheme:    "https",
			wantDomain:    "cnn",
			wantSubdomain: "edition",
			wantTLD:       "com",
			wantPort:      "",
			wantPath:      "/2023/10/01/world/sample-article/index.html",
			wantRawQuery:  "utm_source=newsletter",
			wantFragment:  "",
			wantFileType:  "html",
			wantErr:       false,
		},
		{
			name:          "BBC news URL",
			urlStr:        "https://www.bbc.com/news/world-12345678",
			wantScheme:    "https",
			wantDomain:    "bbc",
			wantSubdomain: "www",
			wantTLD:       "com",
			wantPort:      "",
			wantPath:      "/news/world-12345678",
			wantRawQuery:  "",
			wantFragment:  "",
			wantFileType:  "",
			wantErr:       false,
		},
		{
			name:          "BBC news URL with fragment",
			urlStr:        "https://www.bbc.com/news/world-12345678#fasef",
			wantScheme:    "https",
			wantDomain:    "bbc",
			wantSubdomain: "www",
			wantTLD:       "com",
			wantPort:      "",
			wantPath:      "/news/world-12345678",
			wantRawQuery:  "",
			wantFragment:  "fasef",
			wantFileType:  "",
			wantErr:       false,
		},
		{
			name:          "Simple domain",
			urlStr:        "https://example.com",
			wantScheme:    "https",
			wantDomain:    "example",
			wantSubdomain: "",
			wantTLD:       "com",
			wantPort:      "",
			wantPath:      "",
			wantRawQuery:  "",
			wantFragment:  "",
			wantFileType:  "",
			wantErr:       false,
		},
		{
			name:          "Subdomain example",
			urlStr:        "https://sub.sub2.example.com",
			wantScheme:    "https",
			wantDomain:    "example",
			wantSubdomain: "sub.sub2",
			wantTLD:       "com",
			wantPort:      "",
			wantPath:      "",
			wantRawQuery:  "",
			wantFragment:  "",
			wantFileType:  "",
			wantErr:       false,
		},
		{
			name:          "FTP URL",
			urlStr:        "ftp://invalid-protocol.com/resource",
			wantScheme:    "ftp",
			wantDomain:    "invalid-protocol",
			wantSubdomain: "",
			wantTLD:       "com",
			wantPort:      "",
			wantPath:      "/resource",
			wantRawQuery:  "",
			wantFragment:  "",
			wantFileType:  "",
			wantErr:       false,
		},
		{
			name:          "Invalid URL",
			urlStr:        "not-a-valid-url",
			wantScheme:    "",
			wantDomain:    "",
			wantSubdomain: "",
			wantTLD:       "",
			wantPort:      "",
			wantPath:      "not-a-valid-url",
			wantRawQuery:  "",
			wantFragment:  "",
			wantFileType:  "",
			wantErr:       false,
		},
		{
			name:          "URL with port",
			urlStr:        "http://subdomain.example.co.uk:8080/path/to/page?query=123#section",
			wantScheme:    "http",
			wantDomain:    "example",
			wantSubdomain: "subdomain",
			wantTLD:       "co.uk",
			wantPort:      "8080",
			wantPath:      "/path/to/page",
			wantRawQuery:  "query=123",
			wantFragment:  "section",
			wantFileType:  "",
			wantErr:       false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Parse(tt.urlStr)
			if (err != nil) != tt.wantErr {
				t.Errorf("Parse() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if err != nil {
				return
			}

			if got.Scheme != tt.wantScheme {
				t.Errorf("Parse() Scheme = %v, want %v", got.Scheme, tt.wantScheme)
			}
			if got.Domain != tt.wantDomain {
				t.Errorf("Parse() Domain = %v, want %v", got.Domain, tt.wantDomain)
			}
			if got.Subdomain != tt.wantSubdomain {
				t.Errorf("Parse() Subdomain = %v, want %v", got.Subdomain, tt.wantSubdomain)
			}
			if got.TLD != tt.wantTLD {
				t.Errorf("Parse() TLD = %v, want %v", got.TLD, tt.wantTLD)
			}
			if got.Port != tt.wantPort {
				t.Errorf("Parse() Port = %v, want %v", got.Port, tt.wantPort)
			}
			if got.Path != tt.wantPath {
				t.Errorf("Parse() Path = %v, want %v", got.Path, tt.wantPath)
			}
			if got.RawQuery != tt.wantRawQuery {
				t.Errorf("Parse() RawQuery = %v, want %v", got.RawQuery, tt.wantRawQuery)
			}
			if got.Fragment != tt.wantFragment {
				t.Errorf("Parse() Fragment = %v, want %v", got.Fragment, tt.wantFragment)
			}
			if got.FileType != tt.wantFileType {
				t.Errorf("Parse() FileType = %v, want %v", got.FileType, tt.wantFileType)
			}
		})
	}
}

func TestURL_IsValidNewsArticleURL(t *testing.T) {
	tests := []struct {
		name   string
		urlStr string
		want   bool
	}{
		{
			name:   "CNN article URL",
			urlStr: "https://edition.cnn.com/2023/10/01/world/sample-article/index.html?utm_source=newsletter",
			want:   true,
		},
		{
			name:   "BBC news URL",
			urlStr: "https://www.bbc.com/news/world-12345678",
			want:   true,
		},
		{
			name:   "BBC news URL with fragment",
			urlStr: "https://www.bbc.com/news/world-12345678#fasef",
			want:   true,
		},
		{
			name:   "Simple domain",
			urlStr: "https://example.com",
			want:   false,
		},
		{
			name:   "Subdomain example",
			urlStr: "https://sub.sub2.example.com",
			want:   false,
		},
		{
			name:   "FTP URL",
			urlStr: "ftp://invalid-protocol.com/resource",
			want:   false,
		},
		{
			name:   "Invalid URL",
			urlStr: "not-a-valid-url",
			want:   false,
		},
		{
			name:   "URL with port",
			urlStr: "http://subdomain.example.co.uk:8080/path/to/page?query=123#section",
			want:   false,
		},
		{
			name:   "Mailto URL",
			urlStr: "mailto:test@example.com",
			want:   false,
		},
		{
			name:   "Bad domain - Google",
			urlStr: "https://www.google.com/search?q=test",
			want:   false,
		},
		{
			name:   "Bad path - careers",
			urlStr: "https://example.com/careers/job-opening",
			want:   false,
		},
		{
			name:   "Good path - story",
			urlStr: "https://example.com/story/breaking-news",
			want:   true,
		},
		{
			name:   "Numeric ID pattern",
			urlStr: "https://example.com/article/123456",
			want:   true,
		},
		{
			name:   "Date pattern",
			urlStr: "https://example.com/2023/10/01/article-title",
			want:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			u, err := Parse(tt.urlStr)
			if err != nil {
				t.Fatalf("Parse() error = %v", err)
			}
			if got := u.IsValidNewsArticleURL(); got != tt.want {
				t.Errorf("URL.IsValidNewsArticleURL() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestURL_IsAbsolute(t *testing.T) {
	tests := []struct {
		name   string
		urlStr string
		want   bool
	}{
		{
			name:   "Absolute URL",
			urlStr: "https://example.com/path",
			want:   true,
		},
		{
			name:   "Relative URL",
			urlStr: "/path",
			want:   false,
		},
		{
			name:   "Invalid URL",
			urlStr: "not-a-url",
			want:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			u, err := Parse(tt.urlStr)
			if err != nil {
				t.Fatalf("Parse() error = %v", err)
			}
			if got := u.IsAbsolute(); got != tt.want {
				t.Errorf("URL.IsAbsolute() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestURL_GetPathChunks(t *testing.T) {
	tests := []struct {
		name   string
		urlStr string
		want   []string
	}{
		{
			name:   "CNN article path",
			urlStr: "https://edition.cnn.com/2023/10/01/world/sample-article/index.html",
			want:   []string{"2023", "10", "01", "world", "sample-article", "index.html"},
		},
		{
			name:   "Simple path",
			urlStr: "https://example.com/path/to/resource",
			want:   []string{"path", "to", "resource"},
		},
		{
			name:   "Root path",
			urlStr: "https://example.com/",
			want:   []string{},
		},
		{
			name:   "No path",
			urlStr: "https://example.com",
			want:   []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			u, err := Parse(tt.urlStr)
			if err != nil {
				t.Fatalf("Parse() error = %v", err)
			}
			got := u.GetPathChunks()
			if len(got) != len(tt.want) {
				t.Errorf("URL.GetPathChunks() = %v, want %v", got, tt.want)
				return
			}
			for i, chunk := range got {
				if chunk != tt.want[i] {
					t.Errorf("URL.GetPathChunks()[%d] = %v, want %v", i, chunk, tt.want[i])
				}
			}
		})
	}
}

func TestPrepareURL(t *testing.T) {
	tests := []struct {
		name      string
		urlStr    string
		sourceURL string
		want      string
	}{
		{
			name:      "Relative URL with source",
			urlStr:    "/2023/10/01/article",
			sourceURL: "https://edition.cnn.com",
			want:      "https://edition.cnn.com/2023/10/01/article",
		},
		{
			name:      "Absolute URL with source",
			urlStr:    "https://example.com/article",
			sourceURL: "https://cnn.com",
			want:      "https://example.com/article",
		},
		{
			name:      "URL without source",
			urlStr:    "https://example.com/article",
			sourceURL: "",
			want:      "https://example.com/article",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := PrepareURL(tt.urlStr, tt.sourceURL)
			if got != tt.want {
				t.Errorf("PrepareURL() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestRedirectBack(t *testing.T) {
	tests := []struct {
		name         string
		urlStr       string
		sourceDomain string
		want         string
	}{
		{
			name:         "URL with redirect parameter",
			urlStr:       "https://pinterest.com/pin?url=https://realnews.com/article",
			sourceDomain: "pinterest.com",
			want:         "https://realnews.com/article",
		},
		{
			name:         "URL without redirect parameter",
			urlStr:       "https://example.com/article",
			sourceDomain: "example.com",
			want:         "https://example.com/article",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := RedirectBack(tt.urlStr, tt.sourceDomain)
			if got != tt.want {
				t.Errorf("RedirectBack() = %v, want %v", got, tt.want)
			}
		})
	}
}
