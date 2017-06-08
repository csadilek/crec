package content

import (
	"strings"

	"golang.org/x/text/language"
)

// Recommendations computed by recommenders
type Recommendations []*Content

// Recommender is an extension point for content recommenders. It computes
// content recommendations given a reference to the system's index and a
// map of parameters provided by the client.
type Recommender interface {
	Recommend(index *Index, params map[string]interface{}) (Recommendations, error)
}

// TagBasedRecommender recommends content based on tags (matching categories)
type TagBasedRecommender struct{}

// Recommend content based on the provided tags (matching categories)
func (r *TagBasedRecommender) Recommend(index *Index, params map[string]interface{}) (Recommendations, error) {
	allContent := index.GetLocalizedContent(params["lang-tags"].([]language.Tag))
	var c []*Content
	tags := params["tags"].(string)
	if tags != "" {
		var tagSplits []string
		var disjunction = true
		tagSplits = strings.Split(tags, ",")

		if !strings.Contains(tags, ",") && strings.Contains(tags, " ") {
			tagSplits = strings.Split(tags, " ")
			disjunction = false
		}

		tagMap := make(map[string]bool)
		for _, s := range tagSplits {
			tagMap[strings.TrimSpace(strings.ToLower(s))] = true
		}

		if disjunction {
			c = Filter(allContent, AnyTagFilter(tagMap))
		} else {
			c = Filter(allContent, AllTagFilter(tagMap))
		}
	}

	c = Transform(c, func(item Content) *Content {
		item.Explanation = "Selected for users interested in " + tags
		return &item
	})
	return c, nil
}

// QueryBasedRecommender recommends content based on a full-text query
type QueryBasedRecommender struct {
}

// Recommend content matching the provided full-text query
func (r *QueryBasedRecommender) Recommend(index *Index, params map[string]interface{}) (Recommendations, error) {
	query := params["query"].(string)
	if query != "" {
		return index.Query(query)
	}

	return []*Content{}, nil
}

// ProviderBasedRecommender recommends content based on a full-text query
type ProviderBasedRecommender struct {
}

// Recommend content from the given provider
func (r *ProviderBasedRecommender) Recommend(index *Index, params map[string]interface{}) (Recommendations, error) {
	provider := params["provider"].(string)
	if provider != "" {
		return index.GetProviderContent(provider), nil
	}

	return []*Content{}, nil
}
