package newspaper4k

import (
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/tguidoux/newspaper4k-go/internal/helpers"
	"github.com/tguidoux/newspaper4k-go/internal/ioc"
	"github.com/tguidoux/newspaper4k-go/internal/parsers"
	urlslib "github.com/tguidoux/newspaper4k-go/internal/urls"
	"github.com/tguidoux/newspaper4k-go/pkg/configuration"
	"github.com/tguidoux/newspaper4k-go/pkg/newspaper"
)

// IOCsExtractor extracts indicators of compromise from an article
type IOCsExtractor struct {
	config *configuration.Configuration
}

// NewIOCsExtractor creates a new IOCsExtractor
func NewIOCsExtractor(config *configuration.Configuration) *IOCsExtractor {
	return &IOCsExtractor{config: config}
}

// Parse extracts IOCs from the article's HTML and text and updates the Article
func (xe *IOCsExtractor) Parse(a *newspaper.Article) error {
	// Ensure document is available
	if a.Doc == nil {
		doc, err := parsers.FromString(a.HTML)
		if err == nil {
			a.Doc = doc
		}
	}

	// Aggregate sources to search: article HTML, full HTML, and text
	sources := []string{}
	if a.ArticleHTML != "" {
		sources = append(sources, a.ArticleHTML)
	}
	if a.HTML != "" {
		sources = append(sources, a.HTML)
	}
	if a.Text != "" {
		sources = append(sources, a.Text)
	}

	combined := strings.Join(sources, "\n")
	if combined == "" {
		return nil
	}

	// Extract IOCs using internal/ioc package
	found := ioc.ExtractIOCs(combined, true)

	// Build buckets for each IOC type
	emails := []string{}
	domains := []string{}
	ipv4s := []string{}
	ipv6s := []string{}
	urls := []string{}
	files := []string{}
	bitcoins := []string{}
	md5s := []string{}
	sha1s := []string{}
	sha256s := []string{}
	sha512s := []string{}
	cves := []string{}
	capecs := []string{}
	cwes := []string{}
	cpes := []string{}
	otherurls := []string{}

	for _, fi := range found {
		switch fi.Type {
		case ioc.Email:
			emails = append(emails, fi.IOC)
		case ioc.Domain:
			domains = append(domains, fi.IOC)
		case ioc.IPv4:
			ipv4s = append(ipv4s, fi.IOC)
		case ioc.IPv6:
			ipv6s = append(ipv6s, fi.IOC)
		case ioc.URL:
			urls = append(urls, fi.IOC)
		case ioc.File:
			files = append(files, fi.IOC)
		case ioc.Bitcoin:
			bitcoins = append(bitcoins, fi.IOC)
		case ioc.MD5:
			md5s = append(md5s, fi.IOC)
		case ioc.SHA1:
			sha1s = append(sha1s, fi.IOC)
		case ioc.SHA256:
			sha256s = append(sha256s, fi.IOC)
		case ioc.SHA512:
			sha512s = append(sha512s, fi.IOC)
		case ioc.CVE:
			cves = append(cves, fi.IOC)
		case ioc.CAPEC:
			capecs = append(capecs, fi.IOC)
		case ioc.CWE:
			cwes = append(cwes, fi.IOC)
		case ioc.CPE:
			cpes = append(cpes, fi.IOC)
		default:
			otherurls = append(otherurls, fi.IOC)
		}
	}

	// Normalize/deduplicate preserving order
	a.Emails = helpers.UniqueStrings(emails, helpers.UniqueOptions{CaseSensitive: false, PreserveOrder: true})
	a.Domains = helpers.UniqueStrings(domains, helpers.UniqueOptions{CaseSensitive: false, PreserveOrder: true})
	a.IPv4s = helpers.UniqueStrings(ipv4s, helpers.UniqueOptions{CaseSensitive: false, PreserveOrder: true})
	a.IPv6s = helpers.UniqueStrings(ipv6s, helpers.UniqueOptions{CaseSensitive: false, PreserveOrder: true})
	a.Files = helpers.UniqueStrings(files, helpers.UniqueOptions{CaseSensitive: false, PreserveOrder: true})
	a.Bitcoins = helpers.UniqueStrings(bitcoins, helpers.UniqueOptions{CaseSensitive: false, PreserveOrder: true})
	a.MD5s = helpers.UniqueStrings(md5s, helpers.UniqueOptions{CaseSensitive: false, PreserveOrder: true})
	a.SHA1s = helpers.UniqueStrings(sha1s, helpers.UniqueOptions{CaseSensitive: false, PreserveOrder: true})
	a.SHA256s = helpers.UniqueStrings(sha256s, helpers.UniqueOptions{CaseSensitive: false, PreserveOrder: true})
	a.SHA512s = helpers.UniqueStrings(sha512s, helpers.UniqueOptions{CaseSensitive: false, PreserveOrder: true})
	a.CVEs = helpers.UniqueStrings(cves, helpers.UniqueOptions{CaseSensitive: false, PreserveOrder: true})
	a.CAPECs = helpers.UniqueStrings(capecs, helpers.UniqueOptions{CaseSensitive: false, PreserveOrder: true})
	a.CWEs = helpers.UniqueStrings(cwes, helpers.UniqueOptions{CaseSensitive: false, PreserveOrder: true})
	a.CPEs = helpers.UniqueStrings(cpes, helpers.UniqueOptions{CaseSensitive: false, PreserveOrder: true})

	// If the article has a Document, also look for links in anchor tags
	if a.Doc != nil {
		a.Doc.Find("a").Each(func(i int, s *goquery.Selection) {
			href, _ := s.Attr("href")
			if href != "" {
				otherurls = append(otherurls, href)
			}
		})
	}

	// Parse and normalize other URLs
	preparedSourceURL, err := urlslib.Parse(a.SourceURL)
	if err != nil {
		preparedSourceURL = nil
	}
	for i, u := range otherurls {
		parsed, err := urlslib.Parse(u)
		if err != nil {
			otherurls[i] = u
			continue
		}

		err = parsed.Prepare(preparedSourceURL)
		if err != nil {
			continue
		}

		otherurls[i] = parsed.String()

	}

	// Combine and deduplicate all URLs
	a.OtherURLs = helpers.UniqueStrings(append(urls, otherurls...), helpers.UniqueOptions{CaseSensitive: false, PreserveOrder: true})

	return nil
}
