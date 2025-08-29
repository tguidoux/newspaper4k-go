# newspaper4k-go

[![Go Report Card](https://goreportcard.com/badge/github.com/tguidoux/newspaper4k-go)](https://goreportcard.com/report/github.com/tguidoux/newspaper4k-go)
[![PkgGoDev](https://pkg.go.dev/badge/github.com/tguidoux/newspaper4k-go)](https://pkg.go.dev/github.com/tguidoux/newspaper4k-go)
[![License](https://img.shields.io/badge/license-MIT-blue.svg)](https://opensource.org/licenses/MIT)
[![Actions Status](https://github.com/tguidoux/newspaper4k-go/workflows/ci/badge.svg)](https://github.com/tguidoux/newspaper4k-go/actions)
[![codecov](https://codecov.io/gh/tguidoux/newspaper4k-go/branch/main/graph/badge.svg)](https://codecov.io/gh/tguidoux/newspaper4k-go)
[![Release](https://img.shields.io/github/release/tguidoux/newspaper4k-go.svg?style=flat-square)](RELEASE-NOTES.md)

A Go library for extracting and parsing articles from web pages. This is a Go implementation of [newspaper4k](https://github.com/AndyTheFactory/newspaper4k), inspired by the [newspaper](https://github.com/codelucas/newspaper) Python library. It provides functionality to scrape article content, extract metadata such as title, authors, publish date, and more from HTML pages or URLs.

**Note:** This library does not support all features of newspaper4k yet, but it can successfully parse articles! Pull requests are welcome to add more features.

## Installation

To use this module in your Go project, you can use the `go get` command:

```bash
$ go get -u github.com/tguidoux/newspaper4k-go
```

## Usage

```go
package main

import (
	"fmt"

	"github.com/tguidoux/newspaper4k-go/pkg/configuration"
	"github.com/tguidoux/newspaper4k-go/pkg/newspaper4k"
)

func main() {
	configuration := configuration.NewConfiguration()
	extractors := newspaper4k.DefaultExtractors(configuration)
	article, err := newspaper4k.NewArticleFromURL("https://www.lemonde.fr/m-styles/article/2025/08/29/marie-chioca-photographe-et-autrice-culinaire-je-me-suis-rendu-compte-que-je-ne-trouvais-pas-d-ouvrages-alliant-cuisine-gourmande-et-cuisine-saine-j-ai-donc-decide-de-les-ecrire_6637363_4497319.html")
	if err != nil {
		fmt.Printf("Error fetching article: %v\n", err)
		return
	}
	err = article.Build(extractors)
	if err != nil {
		fmt.Printf("Error building article: %v\n", err)
		return
	}
	fmt.Println(article.IsParsed)

	// Display results
	jsonArticle, err := article.ToJSON()
	if err != nil {
		fmt.Printf("Error converting article to JSON: %v\n", err)
		return
	}
	fmt.Println("Article fetched from URL and parsed:")
	fmt.Println(jsonArticle)
}

```

More to see in the [examples](./examples) directory.

## Configuration

This library does not require any configuration. Simply import and use the functions as shown in the Usage section.

## Contributing

We welcome contributions! Please feel free to submit a Pull Request. For major changes, please open an issue first to discuss what you would like to change.

Please make sure to update tests as appropriate.

## License

This project is licensed under the MIT License - see the [LICENSE.md](LICENSE.md) file for details.

## Acknowledgments

- This is a Go port of [newspaper4k](https://github.com/AndyTheFactory/newspaper4k) by [AndyTheFactory](https://github.com/AndyTheFactory), thanks to him for continuing the work of the original Python library.
- Thanks to [codelucas](https://github.com/codelucas) for the original [newspaper](https://github.com/codelucas/newspaper) Python library
- Thanks to all contributors and the Go community

## Contact

Théo Guidoux - [GitHub](https://github.com/tguidoux)

## Release History

See the [RELEASE-NOTES.md](RELEASE-NOTES.md) file for details.

## Support

If you encounter any issues or have questions, feel free to [create an issue](https://github.com/tguidoux/newspaper4k-go/issues).
