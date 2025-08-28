package parsers

import (
	"strings"
	"testing"

	"github.com/PuerkitoBio/goquery"
)

func TestDropTags(t *testing.T) {
	html := `<div><p>text</p><span>span</span></div>`
	doc, _ := FromString(html)
	p := doc.Find("p")
	span := doc.Find("span")

	DropTags(p, span)

	if doc.Find("p").Length() != 0 || doc.Find("span").Length() != 0 {
		t.Error("Tags were not dropped")
	}

	text := GetText(doc.Find("div"))
	if text != "" {
		t.Errorf("Expected '', got '%s'", text)
	}
}

func TestGetUnicodeHTML(t *testing.T) {
	input := "Hello, 世界"
	output := GetUnicodeHTML(input)
	if output != input {
		t.Errorf("Expected %s, got %s", input, output)
	}
}

func TestFromString(t *testing.T) {
	html := `<html><body><h1>Title</h1></body></html>`
	doc, err := FromString(html)
	if err != nil {
		t.Fatalf("Error parsing HTML: %v", err)
	}

	if doc.Find("h1").Text() != "Title" {
		t.Error("Failed to parse HTML correctly")
	}
}

func TestNodeToString(t *testing.T) {
	html := `<div><p>text</p></div>`
	doc, _ := FromString(html)
	node := doc.Find("p")
	result := NodeToString(node)
	if result != "text" {
		t.Errorf("Expected 'text', got '%s'", result)
	}
}

func TestGetTagsRegex(t *testing.T) {
	html := `<div><a href="http://example.com">link1</a><a href="http://test.com">link2</a></div>`
	doc, _ := FromString(html)
	attribs := map[string]string{"href": "example"}
	results := GetTagsRegex(doc.Selection, "a", attribs)
	if len(results) != 1 {
		t.Errorf("Expected 1 result, got %d", len(results))
	}
}

func TestGetTags(t *testing.T) {
	html := `<div><a href="http://example.com">link</a></div>`
	doc, _ := FromString(html)
	attribs := map[string]string{"href": "http://example.com"}
	results := GetTags(doc.Selection, "a", attribs, "exact", false)
	if len(results) != 1 {
		t.Errorf("Expected 1 result, got %d", len(results))
	}
}

func TestGetElementsByAttribs(t *testing.T) {
	html := `<div><a href="http://example.com">link</a></div>`
	doc, _ := FromString(html)
	attribs := map[string]string{"href": "http://example.com"}
	results := GetElementsByAttribs(doc.Selection, attribs, "exact")
	if len(results) != 1 {
		t.Errorf("Expected 1 result, got %d", len(results))
	}
}

func TestGetMetatags(t *testing.T) {
	html := `<html><head><meta name="description" content="test"></head></html>`
	doc, _ := FromString(html)
	results := GetMetatags(doc.Selection, "description")
	if len(results) != 1 {
		t.Errorf("Expected 1 result, got %d", len(results))
	}
}

func TestGetElementsByTagslist(t *testing.T) {
	html := `<div><p>para</p><span>span</span></div>`
	doc, _ := FromString(html)
	results := GetElementsByTagslist(doc.Selection, []string{"p", "span"})
	if len(results) != 2 {
		t.Errorf("Expected 2 results, got %d", len(results))
	}
}

func TestCreateElement(t *testing.T) {
	element := CreateElement("p", "text", "")
	if element == nil {
		t.Error("Failed to create element")
	}
	if element.Text() != "text" {
		t.Errorf("Expected 'text', got '%s'", element.Text())
	}
}

func TestRemove(t *testing.T) {
	html := `<div><p>text</p><span>keep</span></div>`
	doc, _ := FromString(html)
	p := doc.Find("p")
	Remove([]*goquery.Selection{p}, []string{"span"})
	if doc.Find("p").Length() != 0 {
		t.Error("Element was not removed")
	}
	if doc.Find("span").Length() == 0 {
		t.Error("Kept element was removed")
	}
}

func TestGetText(t *testing.T) {
	html := `<div><p>text</p><script>script</script></div>`
	doc, _ := FromString(html)
	text := GetText(doc.Find("div"))
	if text != "text" {
		t.Errorf("Expected 'text', got '%s'", text)
	}
}

func TestInnerTrim(t *testing.T) {
	input := "  hello   world  "
	expected := "hello world"
	result := InnerTrim(input)
	if result != expected {
		t.Errorf("Expected '%s', got '%s'", expected, result)
	}
}

func TestGetAttribute(t *testing.T) {
	html := `<a href="http://example.com">link</a>`
	doc, _ := FromString(html)
	a := doc.Find("a")
	href := GetAttribute(a, "href", nil, "")
	if href != "http://example.com" {
		t.Errorf("Expected 'http://example.com', got '%s'", href)
	}
}

func TestSetAttribute(t *testing.T) {
	html := `<a>link</a>`
	doc, _ := FromString(html)
	a := doc.Find("a")
	SetAttribute(a, "href", "http://example.com")
	href, _ := a.Attr("href")
	if href != "http://example.com" {
		t.Errorf("Expected 'http://example.com', got '%s'", href)
	}
}

func TestOuterHTML(t *testing.T) {
	html := `<p>text</p>`
	doc, _ := FromString(html)
	p := doc.Find("p")
	result := OuterHTML(p)
	if !strings.Contains(result, "<p>") || !strings.Contains(result, "</p>") {
		t.Errorf("Expected outer HTML, got '%s'", result)
	}
}

func TestGetLdJsonObject(t *testing.T) {
	html := `<script type="application/ld+json">{"name": "test"}</script>`
	doc, _ := FromString(html)
	results := GetLdJsonObject(doc.Selection)
	if len(results) != 1 {
		t.Errorf("Expected 1 result, got %d", len(results))
	}
	if results[0]["name"] != "test" {
		t.Errorf("Expected 'test', got '%v'", results[0]["name"])
	}
}

func TestGetNodeDepth(t *testing.T) {
	html := `<div><p><span>text</span></p></div>`
	doc, _ := FromString(html)
	div := doc.Find("div")
	depth := GetNodeDepth(div)
	if depth != 2 {
		t.Errorf("Expected depth 2, got %d", depth)
	}
}

func TestGetLevel(t *testing.T) {
	html := `<html><body><div><p>text</p></div></body></html>`
	doc, _ := FromString(html)
	p := doc.Find("p")
	level := GetLevel(p)
	if level != 3 {
		t.Errorf("Expected level 3, got %d", level)
	}
}

func TestGetNodesAtLevel(t *testing.T) {
	html := `<div><p><span>text</span></p></div>`
	doc, _ := FromString(html)
	div := doc.Find("div")
	nodes := GetNodesAtLevel(div, 2)
	if len(nodes) != 1 {
		t.Errorf("Expected 1 node at level 2, got %d", len(nodes))
	}
}

func TestIsHighlinkDensity(t *testing.T) {
	html := `<div><a>link1</a><a>link2</a><a>link3</a> some text</div>`
	doc, _ := FromString(html)
	div := doc.Find("div")
	if !IsHighlinkDensity(div, "en") {
		t.Error("Expected high link density")
	}
}

func TestGetNodeGravityScore(t *testing.T) {
	html := `<div gravityScore="1.5">text</div>`
	doc, _ := FromString(html)
	div := doc.Find("div")
	score := GetNodeGravityScore(div)
	if score != 0.0 {
		t.Errorf("Expected 0.0, got %f", score)
	}
}

func TestGetWordCount(t *testing.T) {
	text := "hello world test"
	count := getWordCount(text, "")
	if count != 3 {
		t.Errorf("Expected 3, got %d", count)
	}
}

func TestMin(t *testing.T) {
	if min(1, 2) != 1 {
		t.Error("min(1, 2) should be 1")
	}
	if min(2, 1) != 1 {
		t.Error("min(2, 1) should be 1")
	}
}
