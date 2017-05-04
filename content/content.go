package content

import (
	"fmt"
	"strings"
)

// Content represents our unified data model abstracting from the various formats used by providers.
type Content struct {
	// Globally unique content identifier
	ID string `json:"id,omitempty"`

	// Identifier of the content provider
	Source string `json:"source,omitempty"`

	// Title provided for this content
	Title string `json:"title,omitempty"`

	// Link to a web site showing a detailed view of the content
	Link string `json:"link,omitempty"`

	// Image URI to a preview image
	Image string `json:"image,omitempty"`

	// Summary of the content
	Summary string `json:"summary,omitempty"`

	// HTML view of the content
	HTML string `json:"html,omitempty"`

	// Explanation as to why the content was recommended to a specific client
	Explanation string `json:"explanation,omitempty"`

	// Author of the content
	Author string `json:"author,omitempty"`

	// Publication date
	Published string `json:"published,omitempty"`

	// Tags and categories applied to this content
	Tags []string `json:"tags,omitempty"`

	// Language the content is written in
	Language string `json:"language,omitempty"`

	// Regions the content is applicable to
	Regions []string `json:"regions,omitempty"`

	// Script the content is written in
	Script string `json:"script,omitempty"`
}

func (c *Content) String() string {
	return fmt.Sprintf("Source: %s: Title: %s", c.Source, c.Title)
}

// filter content using the provided predicate function
func filter(c []*Content, f func(*Content) bool) []*Content {
	vsf := make([]*Content, 0)
	for _, v := range c {
		if f(v) {
			vsf = append(vsf, v)
		}
	}
	return vsf
}

// anyTagFilter returns a filter function which retains the content if any
// of the provided tags is present
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

// allTagFilter returns a filter functions which retains the content if all
// of the provided tags are present
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
