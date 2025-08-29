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
