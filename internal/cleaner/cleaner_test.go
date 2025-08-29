package cleaner

import (
	"strings"
	"testing"

	"github.com/PuerkitoBio/goquery"
)

func TestNewDocumentCleaner(t *testing.T) {
	dc := NewDocumentCleaner()
	if dc == nil {
		t.Fatal("NewDocumentCleaner returned nil")
	}
}

func TestCleanWhitespace(t *testing.T) {
	dc := NewDocumentCleaner()
	input := "Hello\tworld\n\nThis is a test\n"
	expected := "Hello world\n\nThis is a test"
	result := dc.CleanWhitespace(input)
	if result != expected {
		t.Errorf("Expected %q, got %q", expected, result)
	}
}

func TestClean(t *testing.T) {
	dc := NewDocumentCleaner()
	html := `<html><body class="test-class"><script>alert('test')</script><p>Hello <span>world</span></p><em>italic</em><em><img src="test.jpg"/></em></body></html>`
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(html))
	if err != nil {
		t.Fatal(err)
	}
	result := dc.Clean(doc.Selection)

	// Check that body class is removed
	if result.Find("body").AttrOr("class", "") != "" {
		t.Error("Body class not removed")
	}

	// Check that script is removed
	if result.Find("script").Length() > 0 {
		t.Error("Script not removed")
	}

	// Check that span inside p is removed
	if result.Find("p span").Length() > 0 {
		t.Error("Span inside p not removed")
	}

	// Check that empty em is removed
	if result.Find("em").Length() != 1 {
		t.Error("Empty em not removed correctly")
	}
}

func TestCleanBodyClasses(t *testing.T) {
	dc := NewDocumentCleaner()
	html := `<html><body class="test-class"><p>Content</p></body></html>`
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(html))
	if err != nil {
		t.Fatal(err)
	}
	result := dc.cleanBodyClasses(doc.Selection)
	if result.Find("body").AttrOr("class", "") != "" {
		t.Error("Body class not removed")
	}
}

func TestCleanArticleTags(t *testing.T) {
	dc := NewDocumentCleaner()
	html := `<html><body><article id="art" class="art-class" name="art-name"><p>Content</p></article></body></html>`
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(html))
	if err != nil {
		t.Fatal(err)
	}
	result := dc.cleanArticleTags(doc.Selection)
	article := result.Find("article")
	if article.AttrOr("id", "") != "" || article.AttrOr("class", "") != "" || article.AttrOr("name", "") != "" {
		t.Error("Article attributes not removed")
	}
}

func TestCleanEmTags(t *testing.T) {
	dc := NewDocumentCleaner()
	html := `<html><body><em>italic</em><em><img src="test.jpg"/></em></body></html>`
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(html))
	if err != nil {
		t.Fatal(err)
	}
	result := dc.cleanEmTags(doc.Selection)
	if result.Find("em").Length() != 1 {
		t.Error("Empty em not removed correctly")
	}
}

func TestRemoveScriptsStyles(t *testing.T) {
	dc := NewDocumentCleaner()
	html := `<html><body><script>js</script><style>css</style><p>Content</p></body></html>`
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(html))
	if err != nil {
		t.Fatal(err)
	}
	result := dc.removeScriptsStyles(doc.Selection)
	if result.Find("script").Length() > 0 || result.Find("style").Length() > 0 {
		t.Error("Scripts or styles not removed")
	}
}

func TestCleanBadTags(t *testing.T) {
	dc := NewDocumentCleaner()
	html := `<html><body><aside>aside</aside><nav>nav</nav><p>Content</p></body></html>`
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(html))
	if err != nil {
		t.Fatal(err)
	}
	result := dc.cleanBadTags(doc.Selection)
	if result.Find("aside").Length() > 0 || result.Find("nav").Length() > 0 {
		t.Error("Bad tags not removed")
	}
}

func TestRemoveNodesRegex(t *testing.T) {
	dc := NewDocumentCleaner()
	html := `<html><body><div id="facebook">fb</div><p>Content</p></body></html>`
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(html))
	if err != nil {
		t.Fatal(err)
	}
	result := dc.removeNodesRegex(doc.Selection, dc.facebookRe)
	if result.Find("#facebook").Length() > 0 {
		t.Error("Facebook node not removed")
	}
}

func TestCleanParaSpans(t *testing.T) {
	dc := NewDocumentCleaner()
	html := `<html><body><p>Hello <span>world</span></p></body></html>`
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(html))
	if err != nil {
		t.Fatal(err)
	}
	result := dc.cleanParaSpans(doc.Selection)
	if result.Find("p span").Length() > 0 {
		t.Error("Span inside p not removed")
	}
}
