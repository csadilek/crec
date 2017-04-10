package content

import (
	"fmt"
	"strings"
)

// Content represents our unified data model abstracting from the various formats used by providers.
type Content struct {
	ID          string   `json:"id,omitempty"`          // Globally unique content identifier
	Source      string   `json:"source,omitempty"`      // Identifier of the content provider
	Title       string   `json:"title,omitempty"`       // Title provided for this content
	Link        string   `json:"link,omitempty"`        // Link to a web site showing a detailed view of the content
	Image       string   `json:"image,omitempty"`       // Image URI to a preview image
	Summary     string   `json:"summary,omitempty"`     // Summary of the content
	HTML        string   `json:"html,omitempty"`        // HTML view of the content
	Explanation string   `json:"explanation,omitempty"` // Explanation as to why the content was recommended to a specific client
	Author      string   `json:"author,omitempty"`      // Author of the content
	Published   string   `json:"published,omitempty"`   // Publication date
	Tags        []string `json:"tags,omitempty"`        // Tags and categories applied to this content
	Language    string   `json:"language,omitempty"`    // Language the content is written in
	Regions     []string `json:"regions,omitempty"`     // Regions the content is applicable to
	Script      string   `json:"script,omitempty"`      // Script the content is written in
}

func (c *Content) String() string {
	return fmt.Sprintf("%s: %s", c.Source, c.Title)
}

// filter the content using the provided predicate function
func filter(c []*Content, f func(*Content) bool) []*Content {
	vsf := make([]*Content, 0)
	for _, v := range c {
		if f(v) {
			vsf = append(vsf, v)
		}
	}
	return vsf
}

// anyTagFilter returns a filter function which retains the content if any tag is present
func anyTagFilter(tags map[string]bool) func(*Content) bool {
	return func(c *Content) bool {
		for _, t := range c.Tags {
			if _, ok := tags[strings.ToLower(t)]; ok {
				return true
			}
		}
		return false
	}
}

// allTagFilter returns a filter functions which retains the content if all tags are present
func allTagFilter(tags map[string]bool) func(*Content) bool {
	return func(c *Content) bool {
		tagMap := make(map[string]bool)
		for _, tag := range c.Tags {
			tagMap[strings.TrimSpace(tag)] = true
		}
		for k := range tags {
			if _, ok := tagMap[k]; !ok {
				return false
			}
		}
		return true
	}
}

// transform applies the provided function to a copy of every element in the provided array
func transform(c []*Content, f func(Content) *Content) []*Content {
	vsf := make([]*Content, 0)
	for _, v := range c {
		vsf = append(vsf, f(*v))
	}
	return vsf
}
