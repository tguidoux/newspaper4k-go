package cleaner

import (
	"regexp"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/tguidoux/newspaper4k-go/internal/parsers"
	"github.com/tguidoux/newspaper4k-go/pkg/configuration"
)

// Cleaner interface defines methods for cleaning HTML documents
type Cleaner interface {
	Clean(doc *goquery.Selection) *goquery.Selection
	CleanWhitespace(text string) string
}

// DocumentCleaner provides methods to clean and manipulate HTML documents
type DocumentCleaner struct {
	config                 *configuration.Configuration
	removeNodesRe          *regexp.Regexp
	removeNodesRelatedRe   *regexp.Regexp
	divToPRe               *regexp.Regexp
	captionRe              *regexp.Regexp
	googleRe               *regexp.Regexp
	entriesRe              *regexp.Regexp
	facebookRe             *regexp.Regexp
	facebookBroadcastingRe *regexp.Regexp
	twitterRe              *regexp.Regexp
	consentRe              *regexp.Regexp
	containsArticle        string
}

// NewDocumentCleaner creates a new DocumentCleaner with the given configuration
func NewDocumentCleaner(config *configuration.Configuration) *DocumentCleaner {
	dc := &DocumentCleaner{
		config: config,
		removeNodesRe: regexp.MustCompile(
			"^side$|combx|retweet|mediaarticlerelated|menucontainer|" +
				"navbar|storytopbar-bucket|utility-bar|inline-share-tools|" +
				"comment|PopularQuestions|contact|foot|footer|Footer|footnote|" +
				"cnn_strycaptiontxt|cnn_html_slideshow|cnn_strylftcntnt|" +
				"links|meta$|shoutbox|sponsor|" +
				"tags|socialnetworking|socialNetworking|cnnStryHghLght|" +
				"cnn_stryspcvbx|^inset$|pagetools|post-attributes|" +
				"welcome_form|contentTools2|the_answers|" +
				"communitypromo|runaroundLeft|subscribe|vcard|articleheadings|" +
				"date|^print$|popup|author-dropdown|tools|socialtools|byline|" +
				"konafilter|KonaFilter|breadcrumbs|^fn$|wp-caption-text|" +
				"legende|ajoutVideo|timestamp|js_replies",
		),
		removeNodesRelatedRe: regexp.MustCompile(
			`related[-\s\_]?(search|topics|media|info|tags|article|content|links)|` +
				`(search|topics|media|info|tags|article|content|links)[-\s\_]?related`,
		),
		divToPRe:               regexp.MustCompile(`<(a|blockquote|dl|div|img|ol|p|pre|table|ul)`),
		captionRe:              regexp.MustCompile("^caption$"),
		googleRe:               regexp.MustCompile(" google "),
		entriesRe:              regexp.MustCompile("^[^entry-]more.*$"),
		facebookRe:             regexp.MustCompile("facebook"),
		facebookBroadcastingRe: regexp.MustCompile("facebook-broadcasting"),
		twitterRe:              regexp.MustCompile("twitter"),
		consentRe: regexp.MustCompile(
			`cookie|cookies|cookieconsent|cookie-consent|cookie_banner|cookie-banner|cookie_notice|cookie-notice|cookiepolicy|cookie_policy|cookiePolicy|consent|consent-banner|consent-popup|gdpr|eu-consent|ccpa|accept-cookies|cookieNotice`,
		),
		containsArticle: `.//article|.//*[@id="article"]|.//*[contains(@itemprop,"articleBody")]`,
	}

	return dc
}

// Clean removes chunks of the DOM as specified
func (dc *DocumentCleaner) Clean(node *goquery.Selection) *goquery.Selection {
	node = dc.cleanBodyClasses(node)
	node = dc.cleanArticleTags(node)
	node = dc.cleanEmTags(node)
	node = dc.removeDropCaps(node)
	node = dc.removeScriptsStyles(node)
	node = dc.cleanBadTags(node)

	// Remove image captions
	node = dc.removeNodesRegex(node, dc.captionRe)
	node = dc.cleanCaptionTags(node)

	node = dc.removeNodesRegex(node, dc.googleRe)
	node = dc.removeNodesRegex(node, dc.entriesRe)

	// Remove social media cards
	node = dc.removeNodesRegex(node, dc.facebookRe)
	node = dc.removeNodesRegex(node, dc.twitterRe)
	node = dc.removeNodesRegex(node, dc.facebookBroadcastingRe)

	// Remove cookie/consent/GDPR banners and popups
	node = dc.removeNodesRegex(node, dc.consentRe)

	// Remove "related" sections
	node = dc.removeNodesRegex(node, dc.removeNodesRelatedRe)

	// Remove spans inside of paragraphs
	node = dc.cleanParaSpans(node)

	node = dc.reduceArticle(node)

	return node
}

// CleanWhitespace removes tabs, whitespace lines from text and adds double newlines to paragraphs
func (dc *DocumentCleaner) CleanWhitespace(text string) string {
	text = strings.ReplaceAll(text, "\t", " ")
	lines := strings.Split(text, "\n")
	var cleaned []string
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed != "" {
			cleaned = append(cleaned, trimmed)
		}
	}
	return strings.Join(cleaned, "\n\n")
}

// cleanBodyClasses removes the `class` attribute from the <body> tag
func (dc *DocumentCleaner) cleanBodyClasses(node *goquery.Selection) *goquery.Selection {
	node.Find("body").RemoveAttr("class")
	return node
}

// cleanArticleTags removes specified attributes from <article> tags
func (dc *DocumentCleaner) cleanArticleTags(node *goquery.Selection) *goquery.Selection {
	node.Find("article").Each(func(i int, s *goquery.Selection) {
		s.RemoveAttr("id")
		s.RemoveAttr("name")
		s.RemoveAttr("class")
	})
	return node
}

// cleanEmTags removes <em> tags that don't contain any <img> tags
func (dc *DocumentCleaner) cleanEmTags(node *goquery.Selection) *goquery.Selection {
	node.Find("em").Each(func(i int, s *goquery.Selection) {
		if s.Find("img").Length() == 0 {
			parsers.DropTags(s)
		}
	})
	return node
}

// cleanCaptionTags removes image caption tags from the document
func (dc *DocumentCleaner) cleanCaptionTags(node *goquery.Selection) *goquery.Selection {
	// Remove figure tags but keep img tags
	node.Find("figure").Each(func(i int, s *goquery.Selection) {
		imgs := s.Find("img")
		if imgs.Length() > 0 && s.Parent().Length() > 0 {
			s.Parent().AppendSelection(imgs)
		}
		s.Remove()
	})

	// Remove figcaption tags but keep img tags
	node.Find("figcaption").Each(func(i int, s *goquery.Selection) {
		imgs := s.Find("img")
		if imgs.Length() > 0 && s.Parent().Length() > 0 {
			s.Parent().AppendSelection(imgs)
		}
		s.Remove()
	})

	// Remove elements with itemprop="caption"
	node.Find("[itemprop='caption']").Remove()

	// Remove instagram media
	node.Find("[class='instagram-media']").Remove()

	// Remove image-caption classes
	node.Find("[class='image-caption']").Remove()

	// Remove div/span with class containing "caption"
	node.Find("div[class*='caption'], span[class*='caption']").Remove()

	return node
}

// removeDropCaps removes spans with class dropcap or drop_cap
func (dc *DocumentCleaner) removeDropCaps(node *goquery.Selection) *goquery.Selection {
	node.Find("span[class~='dropcap'], span[class~='drop_cap']").Remove()
	return node
}

// removeScriptsStyles removes scripts, styles, and comments from the document
func (dc *DocumentCleaner) removeScriptsStyles(node *goquery.Selection) *goquery.Selection {
	node.Find("script").Remove()
	node.Find("style").Remove()
	// Note: goquery automatically handles comment removal during parsing
	return node
}

// cleanBadTags cleans some known bad tags from the document
func (dc *DocumentCleaner) cleanBadTags(node *goquery.Selection) *goquery.Selection {
	// Remove elements with bad IDs
	node.Find("*").Each(func(i int, s *goquery.Selection) {
		id, exists := s.Attr("id")
		if exists && dc.removeNodesRe.MatchString(id) {
			// Check if it contains article content
			if s.Find("article, [id='article'], [itemprop*='articleBody']").Length() == 0 {
				s.Remove()
			}
		}
	})

	// Remove elements with bad classes
	node.Find("*").Each(func(i int, s *goquery.Selection) {
		class, exists := s.Attr("class")
		if exists && dc.removeNodesRe.MatchString(class) {
			// Check if it contains article content
			if s.Find("article, [id='article'], [itemprop*='articleBody']").Length() == 0 {
				s.Remove()
			}
		}
	})

	// Remove elements with bad names
	node.Find("*").Each(func(i int, s *goquery.Selection) {
		name, exists := s.Attr("name")
		if exists && dc.removeNodesRe.MatchString(name) {
			s.Remove()
		}
	})

	// Remove navigation, menus, headers, footers, etc.
	badTags := []string{"aside", "nav", "noscript", "menu"}
	for _, tag := range badTags {
		node.Find(tag).Remove()
	}

	return node
}

// removeNodesRegex removes HTML nodes that match the specified regex pattern
func (dc *DocumentCleaner) removeNodesRegex(node *goquery.Selection, pattern *regexp.Regexp) *goquery.Selection {
	node.Find("*").Each(func(i int, s *goquery.Selection) {
		id, hasID := s.Attr("id")
		class, hasClass := s.Attr("class")

		if (hasID && pattern.MatchString(id)) || (hasClass && pattern.MatchString(class)) {
			s.Remove()
		}
	})
	return node
}

// cleanParaSpans removes span tags within paragraph tags
func (dc *DocumentCleaner) cleanParaSpans(node *goquery.Selection) *goquery.Selection {
	node.Find("p span").Remove()
	return node
}

// reduceArticle reduces the article by removing unnecessary tags
func (dc *DocumentCleaner) reduceArticle(node *goquery.Selection) *goquery.Selection {
	body := node.Find("body")
	if body.Length() == 0 {
		return node
	}

	keepTags := []string{"p", "br", "img", "h1", "h2", "h3", "h4", "h5", "h6", "ul", "body", "article", "section"}

	body.Find("*").Each(func(i int, s *goquery.Selection) {
		tagName := s.Get(0).Data

		shouldKeep := false
		for _, keepTag := range keepTags {
			if tagName == keepTag {
				shouldKeep = true
				break
			}
		}

		if !shouldKeep {
			text := s.Text()
			// In goquery, we don't have direct access to tail text like in lxml
			// But we can check if the element has any content
			if text == "" && s.Children().Length() == 0 {
				parsers.DropTags(s)
			}
		}
	})

	return node
}
