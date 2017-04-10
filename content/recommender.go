package content

import (
	"strings"
)

// Recommender is an extension point for content recommenders. It suggests
// matching content given a reference to all available content and a map
// of parameters provided by the client.
type Recommender interface {
	Recommend(allContent []*Content, params map[string]string) ([]*Content, error)
}

// TagBasedRecommender recommends content based on tags (matching categories)
type TagBasedRecommender struct{}

// Recommend content based on the provided tags (matching categories)
func (r *TagBasedRecommender) Recommend(allContent []*Content, params map[string]string) ([]*Content, error) {
	var c []*Content
	tags := params["tags"]
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
			c = filter(allContent, anyTagFilter(tagMap))
		} else {
			c = filter(allContent, allTagFilter(tagMap))
		}
	}

	c = transform(c, func(item Content) *Content {
		item.Explanation = "Selected for users interested in " + tags
		return &item
	})
	return c, nil
}

// QueryBasedRecommender recommends content based on a full-text query
type QueryBasedRecommender struct {
	Search func(query string) ([]*Content, error)
}

// Recommend content matching the provided full-text query
func (r *QueryBasedRecommender) Recommend(all []*Content, params map[string]string) ([]*Content, error) {
	query := params["query"]
	if query != "" {
		return r.Search(query)
	}

	return []*Content{}, nil
}
