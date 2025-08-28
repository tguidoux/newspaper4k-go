# newspaper4k-go release notes

## 2025-08-28 - v1.0.0

### Added/Refactoring

- **Initial release** of newspaper4k-go - A Go implementation of newspaper4k
- Complete article extraction functionality from URLs and HTML strings
- Comprehensive metadata extraction including:
  - Article title, authors, and publication date
  - Full article text content
  - Images, videos, and media extraction
  - Keywords and automatic summary generation
  - Language detection and site metadata
  - Favicon, description, and other site information

### Features

- `NewArticleFromURL()` - Extract and parse articles directly from web URLs
- `NewArticleFromHTML()` - Parse articles from HTML string content
- Automatic content cleaning and article structure detection
- Multi-language support with automatic language detection
- Image and video extraction with metadata
- Keyword analysis and content summarization
- Robust error handling for network requests and malformed content

### Others

- Full project setup with Go module structure
- Comprehensive test suite
- Complete documentation and usage examples
- Ready for production use
