package urls

import "testing"

func TestRedirectBack(t *testing.T) {
	t.Run("redirect param url", func(t *testing.T) {
		in := "https://redirector.example.com/?url=https://site.example.com/article/1"
		got := RedirectBack(in, "redirector.example.com")
		want := "https://site.example.com/article/1"
		if got != want {
			t.Fatalf("RedirectBack returned %q, want %q", got, want)
		}
	})

	t.Run("same source domain no change", func(t *testing.T) {
		in := "https://news.example.com/article/1"
		got := RedirectBack(in, "news.example.com")
		if got != in {
			t.Fatalf("RedirectBack modified url for same domain: got %q", got)
		}
	})
}

func TestPrepareURLAndJoin(t *testing.T) {
	t.Run("join relative with base", func(t *testing.T) {
		base := "https://example.com/base/path/"
		rel := "page.html"
		got := PrepareURL(rel, base)
		want := "https://example.com/base/path/page.html"
		if got != want {
			t.Fatalf("PrepareURL returned %q, want %q", got, want)
		}
	})

	t.Run("empty source returns input", func(t *testing.T) {
		in := "https://example.com/foo"
		if got := PrepareURL(in, ""); got != in {
			t.Fatalf("PrepareURL(empty source) = %q, want %q", got, in)
		}
	})
}

func TestValidURL(t *testing.T) {
	cases := []struct {
		url  string
		want bool
	}{
		{"https://example.com/2020/05/01/article-title", true},
		{"https://example.com/news/12345", true},
		{"https://twitter.com/some/story", false},
		{"https://example.com/about", false},
		{"https://example.com/section/page.pdf", false},
		{"https://example.com/section/page.html", true},
	}

	for _, c := range cases {
		got := ValidURL(c.url, false)
		if got != c.want {
			t.Fatalf("ValidURL(%q) = %v, want %v", c.url, got, c.want)
		}
	}
}

func TestURLHelpers(t *testing.T) {
	if got := URLToFileType("https://a.example/page.html"); got != "html" {
		t.Fatalf("URLToFileType returned %q, want %q", got, "html")
	}

	if got := URLToFileType("https://a.example/nopath"); got != "" {
		t.Fatalf("URLToFileType returned %q, want empty string", got)
	}

	if got := GetDomain("https://host.example.com:8080/path"); got != "host.example.com:8080" {
		t.Fatalf("GetDomain returned %q", got)
	}

	if got := GetScheme("ftp://example.com/file"); got != "ftp" {
		t.Fatalf("GetScheme returned %q", got)
	}

	if got := GetPath("https://example.com/a/b"); got != "/a/b" {
		t.Fatalf("GetPath returned %q", got)
	}

	if !IsAbsURL("https://example.com/a") {
		t.Fatalf("IsAbsURL should return true for absolute URL")
	}

	if IsAbsURL("/relative/path") {
		t.Fatalf("IsAbsURL should return false for relative path")
	}

	// URLJoinIfValid should join base + relative
	joined := URLJoinIfValid("https://example.com/dir/", "sub/page")
	if joined != "https://example.com/dir/sub/page" {
		t.Fatalf("URLJoinIfValid returned %q", joined)
	}
}
