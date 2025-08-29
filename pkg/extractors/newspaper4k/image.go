package newspaper4k

import (
	"sort"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/tguidoux/newspaper4k-go/internal/parsers"
	"github.com/tguidoux/newspaper4k-go/internal/urls"
	"github.com/tguidoux/newspaper4k-go/pkg/configuration"
	"github.com/tguidoux/newspaper4k-go/pkg/constants"
	"github.com/tguidoux/newspaper4k-go/pkg/newspaper"
)

// ImageExtractor extracts images from articles
type ImageExtractor struct {
	config    *configuration.Configuration
	topImage  string
	metaImage string
	images    []string
	favicon   string
}

// NewImageExtractor creates a new ImageExtractor
func NewImageExtractor(config *configuration.Configuration) *ImageExtractor {
	return &ImageExtractor{
		config:    config,
		topImage:  "",
		metaImage: "",
		images:    []string{},
		favicon:   "",
	}
}

// Parse extracts images from the article and updates the article in-place
func (ie *ImageExtractor) Parse(a *newspaper.Article) error {
	ie.topImage = ""
	ie.metaImage = ""
	ie.images = []string{}
	ie.favicon = ""

	if a.Doc == nil {
		doc, err := parsers.FromString(a.HTML)
		if err != nil {
			return err
		}
		a.Doc = doc
	}

	ie.parse(a.Doc, a.TopNode, a.URL)

	a.TopImage = ie.topImage
	a.MetaImg = ie.metaImage
	a.Images = ie.images
	a.MetaFavicon = ie.favicon

	return nil
}

// parse main method to extract images from a document
func (ie *ImageExtractor) parse(doc *goquery.Document, topNode *goquery.Selection, articleURL string) {
	ie.favicon = ie.getFavicon(doc)

	ie.metaImage = ie.getMetaImage(doc)
	if ie.metaImage != "" {
		ie.metaImage = urls.JoinURL(articleURL, ie.metaImage)
	}

	ie.images = ie.getImages(doc, articleURL)
	ie.topImage = ie.getTopImage(doc, topNode, articleURL)
}

// getFavicon extracts the favicon from a website
func (ie *ImageExtractor) getFavicon(doc *goquery.Document) string {
	// Look for favicon links using parser's GetTags method
	faviconElements := parsers.GetElementsByTagslist(doc.Selection, []string{"link"})
	for _, element := range faviconElements {
		rel := parsers.GetAttribute(element, "rel", nil, "")
		if relStr, ok := rel.(string); ok && (relStr == "icon" || relStr == "shortcut icon") {
			href := parsers.GetAttribute(element, "href", nil, "")
			if hrefStr, ok := href.(string); ok && hrefStr != "" && ie.favicon == "" {
				ie.favicon = hrefStr
			}
		}
	}
	return ie.favicon
}

// getMetaImage extracts image from the meta tags of the document
func (ie *ImageExtractor) getMetaImage(doc *goquery.Document) string {
	candidates := []ImageCandidate{}

	for _, elem := range constants.META_IMAGE_TAGS {
		var items []*goquery.Selection

		if strings.Contains(elem.Value, "|") {
			// Handle regex matching using parser's GetTagsRegex
			values := strings.Split(elem.Value, "|")
			for _, value := range values {
				attribs := map[string]string{elem.Attr: value}
				regexElements := parsers.GetTagsRegex(doc.Selection, elem.Tag, attribs)
				items = append(items, regexElements...)
			}
		} else {
			// Use parser's GetTags method
			attribs := map[string]string{elem.Attr: elem.Value}
			exactElements := parsers.GetTags(doc.Selection, elem.Tag, attribs, "exact", false)
			items = append(items, exactElements...)
		}

		for _, item := range items {
			content := parsers.GetAttribute(item, elem.Content, nil, "")
			if contentStr, ok := content.(string); ok && contentStr != "" {
				candidates = append(candidates, ImageCandidate{URL: contentStr, Score: elem.Score})
			}
		}
	}

	// Filter out empty URLs
	validCandidates := []ImageCandidate{}
	for _, candidate := range candidates {
		if candidate.URL != "" {
			validCandidates = append(validCandidates, candidate)
		}
	}

	if len(validCandidates) == 0 {
		return ""
	}

	// Sort by score (highest first)
	sort.Slice(validCandidates, func(i, j int) bool {
		return validCandidates[i].Score > validCandidates[j].Score
	})

	return validCandidates[0].URL
}

// getImages gets all image sources from img tags
func (ie *ImageExtractor) getImages(doc *goquery.Document, articleURL string) []string {
	images := []string{}

	doc.Find("img").Each(func(i int, s *goquery.Selection) {
		src := ie.getImageSrc(s)
		if src != "" && !strings.HasPrefix(src, "data:") {
			fullURL := urls.JoinURL(articleURL, src)
			if fullURL != "" {
				images = append(images, fullURL)
			}
		}
	})

	return images
}

// getImageSrc gets the src attribute from an img tag, checking multiple possible attributes
func (ie *ImageExtractor) getImageSrc(img *goquery.Selection) string {
	// Check for various src attributes in order of preference
	srcAttrs := []string{"src", "data-src", "data-original", "data-lazy-src"}

	for _, attr := range srcAttrs {
		if src, exists := img.Attr(attr); exists && src != "" {
			return src
		}
	}

	return ""
}

// getTopImage gets the top image for the article
func (ie *ImageExtractor) getTopImage(doc *goquery.Document, topNode *goquery.Selection, articleURL string) string {
	// If we have a meta image and don't need to fetch images, use it
	if ie.metaImage != "" && !ie.config.FetchImages {
		return ie.metaImage
	}

	imgCandidates := []ImageCandidate{}

	doc.Find("img").Each(func(i int, s *goquery.Selection) {
		src, exists := s.Attr("src")
		if !exists || src == "" || strings.HasPrefix(src, "data:") {
			return
		}

		distance := 0
		if topNode != nil && topNode.Length() > 0 {
			distance = ie.nodeDistance(topNode, s)
		}

		imgCandidates = append(imgCandidates, ImageCandidate{
			URL:     src,
			Score:   distance,
			Element: s,
		})
	})

	if len(imgCandidates) == 0 {
		return ""
	}

	// Sort by distance (closest to top node first)
	sort.Slice(imgCandidates, func(i, j int) bool {
		return imgCandidates[i].Score < imgCandidates[j].Score
	})

	// Return the first valid image
	for _, candidate := range imgCandidates {
		fullURL := urls.JoinURL(articleURL, candidate.URL)
		if fullURL != "" {
			return fullURL
		}
	}

	return ""
}

// nodeDistance calculates the distance between two nodes in the DOM tree
func (ie *ImageExtractor) nodeDistance(node1, node2 *goquery.Selection) int {
	if node1 == nil || node2 == nil {
		return 999
	}

	path1 := ie.getNodePath(node1)
	path2 := ie.getNodePath(node2)

	minLen := len(path1)
	if len(path2) < minLen {
		minLen = len(path2)
	}

	for i := 0; i < minLen; i++ {
		if path1[i] != path2[i] {
			return len(path1[i:]) + len(path2[i:])
		}
	}

	return abs(len(path1) - len(path2))
}

// getNodePath gets the path from root to node
func (ie *ImageExtractor) getNodePath(node *goquery.Selection) []string {
	path := []string{}
	current := node

	for current.Length() > 0 {
		tagName := current.Get(0).Data
		path = append([]string{tagName}, path...)
		current = current.Parent()
	}

	return path
}

// ImageCandidate represents an image with its score
type ImageCandidate struct {
	URL     string
	Score   int
	Element *goquery.Selection
}

// abs returns the absolute value of an integer
func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}
