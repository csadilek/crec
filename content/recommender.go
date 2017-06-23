package content

import (
	"strings"
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

		localizedContent := index.GetLocalizedContent(params["lang"].(string))
		if disjunction {
			lcMap := make(map[*Content]bool)
			for _, lc := range localizedContent {
				lcMap[lc] = true
			}
			for _, t := range tagSplits {
				for _, tc := range index.GetTaggedContent(strings.ToLower(t)) {
					if lcMap[tc] {
						c = append(c, tc)
					}
				}
			}
		} else {
			// TODO could use GetTaggedContent and build up a hit map
			c = Filter(localizedContent, AllTagFilter(tagSplits))
		}
	}
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

// LocaleBasedRecommender recommends content based on the provide locale string e.g. at-DE
type LocaleBasedRecommender struct {
}

// Recommend content for the given locale
func (r *LocaleBasedRecommender) Recommend(index *Index, params map[string]interface{}) (Recommendations, error) {
	locale := params["locale"].(string)
	if locale != "" {
		return index.GetLocalizedContent(locale), nil
	}

	return []*Content{}, nil
}
