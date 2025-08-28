package newspaper

// MOTLEY_REPLACEMENT is used for cleaning title text
var MOTLEY_REPLACEMENT = []string{"&#65533;", ""}

// TITLE_REPLACEMENTS is used for cleaning split titles
var TITLE_REPLACEMENTS = []string{"&raquo;", "»"}

// A_REL_TAG_SELECTOR XPath selector for anchor tags with rel='tag'
const A_REL_TAG_SELECTOR = "//a[@rel='tag']"

// A_HREF_TAG_SELECTOR XPath selector for anchor tags with href containing tag-related paths
const A_HREF_TAG_SELECTOR = "//a[contains(@href, '/tag/')] | //a[contains(@href, '/tags/')] | //a[contains(@href, '/topic/')] | //a[contains(@href, '?keyword=')]"

// RE_LANG regex pattern for language codes
const RE_LANG = "^[A-Za-z]{2}$"

// AUTHOR_ATTRS attributes to look for author information
var AUTHOR_ATTRS = []string{"name", "rel", "itemprop", "class", "id", "property"}

// AUTHOR_VALS values to look for in author attributes
var AUTHOR_VALS = []string{
	"author",
	"byline",
	"dc.creator",
	"byl",
	"article:author",
	"article:author_name",
	"story-byline",
	"article-author",
	"parsely-author",
	"sailthru.author",
	"citation_author",
}

// AUTHOR_STOP_WORDS words to ignore when extracting authors
var AUTHOR_STOP_WORDS = []string{
	"By",
	"Reuters",
	"IANS",
	"AP",
	"AFP",
	"PTI",
	"IANS",
	"ANI",
	"DPA",
	"Senior Reporter",
	"Reporter",
	"Writer",
	"Opinion Writer",
}

// TITLE_META_INFO meta tag names for title information
var TITLE_META_INFO = []string{
	"dc.title",
	"og:title",
	"headline",
	"articletitle",
	"article-title",
	"parsely-title",
	"title",
}

// PUBLISH_DATE_META_INFO meta tag names for publish date information
var PUBLISH_DATE_META_INFO = []string{
	"published_date",
	"published_time",
	"cXenseParse:publishtime",
	"pubdate",
	"publish_date",
	"PublishDate",
	"dcterms.created",
	"rnews:datePublished",
	"article:published_time",
	"prism.publicationDate",
	"displaydate",
	"OriginalPublicationDate",
	"og:published_time",
	"datePublished",
	"article_date_original",
	"article.published",
	"published_time_telegram",
	"sailthru.date",
	"date",
	"Date",
	"original-publish-date",
	"DC.date.issued",
	"dc.date",
	"DC.Date",
	"parsely-pub-date",
	"publishtime",
	"og:regDate",
	"publication_date",
	"uploadDate",
	"coverageEndTime",
	"publishdate",
	"publish-date",
	"publishedAtDate",
	"dcterms.date",
	"publishedDate",
	"creationDateTime",
	"pub_date",
	"updated_time",
	"og:updated_time",
	"datemodified",
	"last-modified",
	"Last-Modified",
	"DC.date.modified",
	"article:modified_time",
	"modified_time",
	"modifiedDateTime",
	"dc.dcterms.modified",
	"lastmod",
	"eomportal-lastUpdate",
}

// PublishDateTag represents a tag configuration for extracting publish dates
type PublishDateTag struct {
	Attribute string `json:"attribute"`
	Value     string `json:"value"`
	Content   string `json:"content"`
}

// PUBLISH_DATE_TAGS configurations for publish date extraction
var PUBLISH_DATE_TAGS = []PublishDateTag{
	{Attribute: "itemprop", Value: "datePublished", Content: "datetime"},
	{Attribute: "pubdate", Value: "pubdate", Content: "datetime"},
	{Attribute: "class", Value: "entry-date", Content: "datetime"},
	{Attribute: "class", Value: "article-date", Content: "content"},
	{Attribute: "class", Value: "postInfoDate", Content: "content"},
	{Attribute: "class", Value: "time", Content: "content"},
	{Attribute: "id", Value: "date", Content: "content"},
}

// ArticleBodyTag represents configuration for article body extraction
type ArticleBodyTag struct {
	Tag        string `json:"tag,omitempty"`
	Class      string `json:"class,omitempty"`
	Itemprop   string `json:"itemprop,omitempty"`
	Itemtype   string `json:"itemtype,omitempty"`
	Role       string `json:"role,omitempty"`
	ScoreBoost int    `json:"score_boost"`
}

// ARTICLE_BODY_TAGS configurations for article body extraction
var ARTICLE_BODY_TAGS = []ArticleBodyTag{
	{Tag: "article", Role: "article", ScoreBoost: 25},
	{Itemprop: "articleBody", ScoreBoost: 100},
	{Itemprop: "articleText", ScoreBoost: 40},
	{Itemtype: "https://schema.org/Article", ScoreBoost: 30},
	{Itemtype: "https://schema.org/NewsArticle", ScoreBoost: 30},
	{Itemtype: "https://schema.org/BlogPosting", ScoreBoost: 20},
	{Itemtype: "https://schema.org/ScholarlyArticle", ScoreBoost: 20},
	{Itemtype: "https://schema.org/SocialMediaPosting", ScoreBoost: 20},
	{Itemtype: "https://schema.org/TechArticle", ScoreBoost: 20},
	{Class: "re:paragraph|entry-content|article-text|article-body", ScoreBoost: 15},
}

// MetaImageDict represents configuration for meta image extraction
type MetaImageDict struct {
	Tag     string `json:"tag"`
	Attr    string `json:"attr"`
	Value   string `json:"value"`
	Content string `json:"content"`
	Score   int    `json:"score"`
}

// META_IMAGE_TAGS configurations for meta image extraction
var META_IMAGE_TAGS = []MetaImageDict{
	{
		Tag:     "meta",
		Attr:    "property",
		Value:   "og:image",
		Content: "content",
		Score:   10,
	},
	{
		Tag:     "link",
		Attr:    "rel",
		Value:   "image_src|img_src",
		Content: "href",
		Score:   8,
	},
	{
		Tag:     "meta",
		Attr:    "name",
		Value:   "og:image",
		Content: "content",
		Score:   8,
	},
	{
		Tag:     "link",
		Attr:    "rel",
		Value:   "icon",
		Content: "href",
		Score:   5,
	},
}

// META_LANGUAGE_TAGS configurations for meta language extraction
var META_LANGUAGE_TAGS = []map[string]string{
	{"tag": "meta", "attr": "property", "value": "og:locale"},
	{"tag": "meta", "attr": "http-equiv", "value": "content-language"},
	{"tag": "meta", "attr": "name", "value": "lang"},
}

// URL_STOPWORDS words to ignore in URLs
var URL_STOPWORDS = []string{
	"about",
	"academy",
	"account",
	"admin",
	"advert",
	"advertise",
	"adverts",
	"archive",
	"archives",
	"author",
	"authors",
	"bebo",
	"board",
	"boards",
	"browse",
	"career",
	"careers",
	"charts",
	"comment",
	"comments",
	"contact",
	"contacts",
	"coupons",
	"developer",
	"developers",
	"donate",
	"download",
	"downloads",
	"edit",
	"event",
	"events",
	"facebook",
	"faq",
	"faqs",
	"feed",
	"feeds",
	"feedback",
	"flickr",
	"forum",
	"forums",
	"friendster",
	"help",
	"how-to",
	"howto",
	"imgur",
	"info",
	"itunes",
	"jobs",
	"legal",
	"linkedin",
	"login",
	"mail",
	"map",
	"maps",
	"mobile",
	"myspace",
	"newsletter",
	"newsletters",
	"password",
	"plus",
	"preference",
	"preferences",
	"privacy",
	"privacy-policy",
	"product",
	"products",
	"profile",
	"profiles",
	"proxy",
	"purchase",
	"register",
	"search",
	"services",
	"shop",
	"shopping",
	"signup",
	"site-map",
	"siteindex",
	"sitemap",
	"sitemaps",
	"stop",
	"store",
	"stumbleupon",
	"subscribe",
	"subscription",
	"tag",
	"terms",
	"tickets",
	"toc",
	"twitter",
	"upload",
	"uploads",
	"vimeo",
	"wp-admin",
	"wp-includes",
	"wp-content",
	"wp-json",
	"wp-login",
	"xmlrpc",
	"xmlrpc.php",
	"youtube",
}

// VIDEO_TAGS tags to search for video elements
var VIDEO_TAGS = []string{"iframe", "embed", "object", "video"}

// VIDEO_PROVIDERS supported video providers
var VIDEO_PROVIDERS = []string{"youtube", "youtu.be", "vimeo", "dailymotion", "kewego", "twitch"}

var CATEGORY_URL_PREFIXES = []string{
	"category",
	"categories",
	"topic",
	"topics",
	"subject",
	"subjects",
	"section",
	"sections",
	"cat",
	"cats",
}

var COMMON_FEED_SUFFIXES = []string{
	"/atom.xml",
	"/blog?format=rss",
	"/blog/feed",
	"/blog/rss",
	"/blog/rss.xml",
	"/blogs/news.atom",
	"/blogs/news.rss",
	"/comments/feed/",
	"/feed",
	"/feed/",
	"/feed.xml",
	"/feed/@username",
	"/feed/publication-name",
	"/feeds/latest.xml",
	"/feeds/posts/default",
	"/feeds/posts/default?alt=rss",
	"/new/.rss",
	"/news/feed",
	"/news/feed.xml",
	"/news/rss",
	"/news/rss.xml",
	"/rss",
	"/rss/",
	"/rss.xml",
	"/rss.xml?format=xml",
}
