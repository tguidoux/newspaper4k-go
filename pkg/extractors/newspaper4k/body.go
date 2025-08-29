package newspaper4k

import (
	"math"
	"sort"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/tguidoux/newspaper4k-go/internal/nlp"
	"github.com/tguidoux/newspaper4k-go/internal/parsers"
	"github.com/tguidoux/newspaper4k-go/pkg/configuration"
	"github.com/tguidoux/newspaper4k-go/pkg/constants"
	"github.com/tguidoux/newspaper4k-go/pkg/newspaper"
)

// scoreWeights mirrors the weights used in the original Python implementation
var scoreWeights = map[string]float64{
	"bottom_negativescore_nodes": 0.25,
	"boost_score":                30,
	"parent_node":                1.0,
	"parent_parent_node":         0.4,
	"node_count_threshold":       15,
	"negative_score_threshold":   40,
	"negative_score_boost":       5.0,
	"boost_max_steps_from_node":  3,
	"boost_min_stopword_count":   5,
}

// BodyExtractor extracts the article body/top node from a document
type BodyExtractor struct {
	config              *configuration.Configuration
	topNode             *goquery.Selection
	topNodeComplemented *goquery.Selection
	stopwords           *nlp.StopWords
}

// NewBodyExtractor creates a new BodyExtractor
func NewBodyExtractor(config *configuration.Configuration) *BodyExtractor {
	return &BodyExtractor{config: config}
}

// Parse computes the top node and augmented top node and updates the Article
func (be *BodyExtractor) Parse(a *newspaper.Article) error {
	if a.Doc == nil {
		doc, err := parsers.FromString(a.HTML)
		if err != nil {
			return err
		}
		a.Doc = doc
	}
	// initialize stopwords
	lang := a.GetLanguage().String()
	sw, _ := nlp.NewStopWords(lang)
	be.stopwords = sw

	be.topNode = be.calculateBestNode(a.Doc)
	be.topNodeComplemented = be.complementWithSiblings(a.Doc, be.topNode)

	// Update article
	a.TopNode = be.topNodeComplemented
	if be.topNodeComplemented != nil {
		a.ArticleHTML = parsers.OuterHTML(be.topNodeComplemented)
		a.Text = parsers.GetText(be.topNodeComplemented)
	}

	return nil
}

// calculateBestNode finds the best node representing the article body
func (be *BodyExtractor) calculateBestNode(doc *goquery.Document) *goquery.Selection {
	var topNode *goquery.Selection

	be.boostHighlyLikelyNodes(doc)

	nodesWithText := be.computeFeatures(doc)

	// sort nodes by level (deepest first)
	sort.Slice(nodesWithText, func(i, j int) bool {
		return parsers.GetLevel(nodesWithText[i]) > parsers.GetLevel(nodesWithText[j])
	})

	parentNodes := be.computeGravityScores(nodesWithText)

	if len(parentNodes) > 0 {
		sort.Slice(parentNodes, func(i, j int) bool {
			return parsers.GetNodeGravityScore(parentNodes[i]) > parsers.GetNodeGravityScore(parentNodes[j])
		})
		topNode = parentNodes[0]
	}

	// Fallback: if we couldn't find a suitable top node, use the document body
	if topNode == nil && doc != nil {
		body := doc.Find("body").First()
		if body != nil && body.Length() > 0 {
			topNode = body
		}
	}

	return topNode
}

// computeGravityScores propagates scores up to parents and grandparents
func (be *BodyExtractor) computeGravityScores(nodesWithText []*goquery.Selection) []*goquery.Selection {
	parentNodesMap := make(map[string]*goquery.Selection)

	nodesCount := len(nodesWithText)
	negativeScoring := 0.0
	bottomNegNodes := float64(nodesCount) * scoreWeights["bottom_negativescore_nodes"]
	boostDiscount := 1.0

	for i, node := range nodesWithText {
		boostScore := 0.0
		if be.isBoostable(node) {
			boostScore = scoreWeights["boost_score"] / boostDiscount
			boostDiscount += 1.0
		}

		if nodesCount > int(scoreWeights["node_count_threshold"]) {
			distFromEnd := float64(nodesCount - i)
			if distFromEnd <= bottomNegNodes {
				booster := bottomNegNodes - distFromEnd
				boostScore = -(booster * booster)
				negscore := boostScore
				if negscore < 0 {
					negscore = -negscore
				}
				negscore = negscore + negativeScoring
				if negscore > scoreWeights["negative_score_threshold"] {
					boostScore = scoreWeights["negative_score_boost"]
				}
			}
		}

		stopWordCountIface := parsers.GetAttribute(node, "stop_words", 0, 0)
		stopWordCount := 0
		if val, ok := stopWordCountIface.(int); ok {
			stopWordCount = val
		}

		upscore := float64(stopWordCount) + boostScore

		parent := node.Parent()
		be.updateScore(parent, upscore)
		be.updateNodeCount(parent, 1)
		if parent != nil && parent.Length() > 0 {
			parentNodesMap[parsers.OuterHTML(parent)] = parent
		}

		parentParent := parent.Parent()
		be.updateNodeCount(parentParent, 1)
		be.updateScore(parentParent, upscore*scoreWeights["parent_parent_node"])
		if parentParent != nil && parentParent.Length() > 0 {
			parentNodesMap[parsers.OuterHTML(parentParent)] = parentParent
		}
	}

	var parents []*goquery.Selection
	for _, v := range parentNodesMap {
		parents = append(parents, v)
	}
	return parents
}

// computeFeatures finds candidate nodes with reasonable text
func (be *BodyExtractor) computeFeatures(doc *goquery.Document) []*goquery.Selection {
	candidates := []*goquery.Selection{}

	nodes := be.nodesToCheck(doc)
	sort.Slice(nodes, func(i, j int) bool { return parsers.GetLevel(nodes[i]) > parsers.GetLevel(nodes[j]) })

	for _, node := range nodes {
		text := parsers.GetText(node)
		if strings.TrimSpace(text) == "" {
			continue
		}

		// count stopwords and words
		stopCount, wordCount := be.getStopwordStats(text)
		highLink := parsers.IsHighlinkDensity(node, be.config.Language())

		// compute children stats
		children := node.Find("*[stop_words]")
		childrenStop := 0
		childrenWords := 0
		children.Each(func(i int, s *goquery.Selection) {
			stopWordsIface := parsers.GetAttribute(s, "stop_words", 0, 0)
			if v, ok := stopWordsIface.(int); ok && v > 0 {
				childrenStop += v
			}
			wordCountIface := parsers.GetAttribute(s, "word_count", 0, 0)
			if vv, ok := wordCountIface.(int); ok {
				childrenWords += vv
			}
		})

		parsers.SetAttribute(node, "stop_words", stopCount-childrenStop)
		parsers.SetAttribute(node, "word_count", wordCount-childrenWords)
		parsers.SetAttribute(node, "is_highlink_density", 0)
		if highLink {
			parsers.SetAttribute(node, "is_highlink_density", 1)
		}
		parsers.SetAttribute(node, "node_level", parsers.GetLevel(node))

		if stopCount > 2 && !highLink {
			candidates = append(candidates, node)
		}
	}

	return candidates
}

// nodesToCheck returns candidate nodes to inspect
func (be *BodyExtractor) nodesToCheck(doc *goquery.Document) []*goquery.Selection {
	var nodes []*goquery.Selection

	tags := []string{"p", "pre", "td", "article", "div"}
	for _, tag := range tags {
		if tag == "div" {
			items := []*goquery.Selection{}
			for _, attr := range []string{"articlebody", "article", "story"} {
				items = append(items, parsers.GetTags(doc.Selection, tag, map[string]string{"id": attr}, "word", true)...)    // id
				items = append(items, parsers.GetTags(doc.Selection, tag, map[string]string{"class": attr}, "word", true)...) // class
			}
			items = append(items, parsers.GetTagsRegex(doc.Selection, tag, map[string]string{"class": "paragraph"})...)
			if len(items) == 0 && len(nodes) < 5 {
				items = parsers.GetTags(doc.Selection, tag, nil, "exact", false)
			}
			// deduplicate
			seen := map[string]bool{}
			for _, it := range items {
				key := parsers.OuterHTML(it)
				if !seen[key] {
					nodes = append(nodes, it)
					seen[key] = true
				}
			}
		} else {
			items := parsers.GetTags(doc.Selection, tag, nil, "exact", false)
			nodes = append(nodes, items...)
		}
	}

	// Do not miss some Article Bodies or Article Sections
	for _, itemprop := range []string{"articleBody", "articlebody", "articleText", "articleSection"} {
		items := parsers.GetTags(doc.Selection, "", map[string]string{"itemprop": itemprop}, "word", false)
		for _, item := range items {
			// Check if not already in nodes
			found := false
			for _, n := range nodes {
				if parsers.OuterHTML(n) == parsers.OuterHTML(item) {
					found = true
					break
				}
			}
			if !found {
				nodes = append(nodes, item)
			}
		}
	}

	return nodes
}

// isBoostable checks whether node should be boosted
func (be *BodyExtractor) isBoostable(node *goquery.Selection) bool {
	maxSteps := int(scoreWeights["boost_max_steps_from_node"])
	siblings := be.walkSiblings(node)
	for i, s := range siblings {
		if i >= maxSteps {
			break
		}
		if s == nil || s.Length() == 0 {
			continue
		}
		if s.Get(0).Data != node.Get(0).Data {
			continue
		}
		sw := parsers.GetAttribute(s, "stop_words", 0, 0)
		if v, ok := sw.(int); ok {
			if v > int(scoreWeights["boost_min_stopword_count"]) {
				return true
			}
		}
	}
	return false
}

// boostHighlyLikelyNodes biases nodes that look like article containers
func (be *BodyExtractor) boostHighlyLikelyNodes(doc *goquery.Document) {
	candidates := []*goquery.Selection{}
	for _, tag := range []string{"p", "pre", "td", "article", "div"} {
		candidates = append(candidates, parsers.GetTags(doc.Selection, tag, nil, "exact", false)...)
	}

	for _, e := range candidates {
		boost := be.isHighlyLikely(e)
		if boost > 0 {
			be.updateScore(e, boost*scoreWeights["parent_node"])
		}
	}
}

// isHighlyLikely checks tag patterns against ARTICLE_BODY_TAGS
func (be *BodyExtractor) isHighlyLikely(node *goquery.Selection) float64 {
	// helper to match tag dict
	match := func(node *goquery.Selection, tag constants.ArticleBodyTag) bool {
		if node.Length() == 0 {
			return false
		}
		ntag := node.Get(0).Data
		if tag.Tag != "" && ntag != tag.Tag {
			return false
		}
		// check attributes
		for k, v := range map[string]string{"class": tag.Class, "itemprop": tag.Itemprop, "itemtype": tag.Itemtype, "role": tag.Role} {
			if v == "" {
				continue
			}
			val, _ := node.Attr(k)
			if strings.HasPrefix(v, "re:") {
				pattern := v[3:]
				if !strings.Contains(strings.ToLower(val), strings.ToLower(pattern)) {
					return false
				}
			} else {
				if !strings.EqualFold(val, v) {
					return false
				}
			}
		}
		return true
	}

	best := 0
	for _, t := range constants.ARTICLE_BODY_TAGS {
		if match(node, t) {
			if t.ScoreBoost > best {
				best = t.ScoreBoost
			}
		}
	}
	return float64(best)
}

// updateScore adds to gravityScore attribute
func (be *BodyExtractor) updateScore(node *goquery.Selection, add float64) {
	if node == nil || node.Length() == 0 {
		return
	}
	cur := parsers.GetNodeGravityScore(node)
	parsers.SetAttribute(node, "gravityScore", cur+add)
}

// updateNodeCount increases gravityNodes counter
func (be *BodyExtractor) updateNodeCount(node *goquery.Selection, add int) {
	if node == nil || node.Length() == 0 {
		return
	}
	curIface := parsers.GetAttribute(node, "gravityNodes", 0, 0)
	cur := 0
	if v, ok := curIface.(int); ok {
		cur = v
	}
	parsers.SetAttribute(node, "gravityNodes", cur+add)
}

// walkSiblings returns preceding siblings (nearest first)
func (be *BodyExtractor) walkSiblings(node *goquery.Selection) []*goquery.Selection {
	var res []*goquery.Selection
	if node == nil || node.Length() == 0 {
		return res
	}
	for s := node.Prev(); s.Length() > 0; s = s.Prev() {
		res = append(res, s)
	}
	return res
}

// complementWithSiblings builds an off-tree node composed of candidates at same level
func (be *BodyExtractor) complementWithSiblings(doc *goquery.Document, node *goquery.Selection) *goquery.Selection {
	if node == nil || node.Length() == 0 {
		return node
	}
	// If the chosen node is already the body or an article element,
	// return it directly to preserve its text and avoid cloning issues.
	if node.Get(0) != nil {
		tag := node.Get(0).Data
		if tag == "body" || tag == "article" {
			return node
		}
	}
	level := parsers.GetLevel(node)
	candidates := parsers.GetNodesAtLevel(doc.Selection, level)

	// create a new document body
	newDoc, _ := parsers.FromString("<html><body></body></html>")
	body := newDoc.Find("body").First()

	// Prefer a normalized baseline score computed from paragraph children
	baseScore := be.getNormalizedScore(node)
	if math.IsInf(baseScore, 1) {
		// fallback to node gravity score
		baseScore = parsers.GetNodeGravityScore(node)
	}

	for _, n := range candidates {
		if parsers.OuterHTML(n) == parsers.OuterHTML(node) {
			body.AppendSelection(n.Clone())
			continue
		}
		// only merge nodes of the same tag type (eg. div with div, article with article)
		if n.Get(0) == nil || node.Get(0) == nil {
			continue
		}
		if n.Get(0).Data != node.Get(0).Data {
			continue
		}

		// accept nodes with comparable gravity score and low link density
		score := parsers.GetNodeGravityScore(n)
		if score > baseScore*0.3 && !parsers.IsHighlinkDensity(n, be.config.Language()) {
			body.AppendSelection(n.Clone())
			continue
		}

		// try to salvage plausible paragraphs from this sibling node
		ps := be.getPlausibleContent(n, baseScore)
		for _, p := range ps {
			if p != nil {
				body.AppendSelection(p)
			}
		}
	}

	return body
}

// getNormalizedScore returns the average positive gravity score of paragraph children
func (be *BodyExtractor) getNormalizedScore(top *goquery.Selection) float64 {
	if top == nil || top.Length() == 0 {
		return math.Inf(1)
	}
	var scores []float64
	top.Find("p").Each(func(i int, s *goquery.Selection) {
		sc := parsers.GetNodeGravityScore(s)
		if sc > 0 {
			scores = append(scores, sc)
		}
	})
	if len(scores) == 0 {
		return math.Inf(1)
	}
	sum := 0.0
	for _, v := range scores {
		sum += v
	}
	return sum / float64(len(scores))
}

// getPlausibleContent extracts paragraphs from a candidate sibling node that
// look like they belong to the article based on stopword counts and link density.
func (be *BodyExtractor) getPlausibleContent(node *goquery.Selection, baseline float64) []*goquery.Selection {
	var res []*goquery.Selection
	if node == nil || node.Length() == 0 {
		return res
	}

	// if this node itself is a paragraph, consider it directly
	if node.Get(0) != nil && node.Get(0).Data == "p" {
		txt := parsers.GetText(node)
		if strings.TrimSpace(txt) != "" && !parsers.IsHighlinkDensity(node, be.config.Language()) {
			// clone paragraph for safety
			res = append(res, node.Clone())
		}
		return res
	}

	// otherwise inspect paragraph children
	node.Find("p").Each(func(i int, s *goquery.Selection) {
		stopWordsIface := parsers.GetAttribute(s, "stop_words", 0, 0)
		stopCount := 0
		if v, ok := stopWordsIface.(int); ok {
			stopCount = v
		}
		if stopCount <= 0 {
			return
		}
		if parsers.IsHighlinkDensity(s, be.config.Language()) {
			return
		}
		// baseline can be Inf when not available -> use node gravity score
		base := baseline
		if math.IsInf(base, 1) {
			base = parsers.GetNodeGravityScore(node)
		}
		weight := 0.3
		if float64(stopCount) > base*weight {
			// create a new paragraph element containing only the text
			el := parsers.CreateElement("p", parsers.GetText(s), "")
			if el != nil {
				res = append(res, el)
			}
		}
	})

	return res
}

// getStopwordStats returns stop word count and word count for text
func (be *BodyExtractor) getStopwordStats(text string) (int, int) {
	if be.stopwords == nil {
		return 0, 0
	}
	tokens := be.stopwords.Tokenize(text)
	stopCount := 0
	for _, t := range tokens {
		if be.stopwords.StopWords[strings.ToLower(t)] {
			stopCount++
		}
	}
	return stopCount, len(tokens)
}
