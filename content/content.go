package content

import (
	"fmt"
	"strings"

	"github.com/mmcdole/gofeed"
)

// Content is just a simple wrapper of a gofeed Item for now. More to come...
type Content struct {
	ID          string       `json:"id,omitempty"`
	Source      string       `json:"source,omitempty"`
	Title       string       `json:"title,omitempty"`
	Link        string       `json:"link,omitempty"`
	Image       string       `json:"image,omitempty"`
	Summary     string       `json:"summary,omitempty"`
	HTML        string       `json:"html,omitempty"`
	Explanation string       `json:"explanation,omitempty"`
	Author      string       `json:"author,omitempty"`
	Published   string       `json:"published,omitempty"`
	Tags        []string     `json:"tags,omitempty"`
	Language    string       `json:"language,omitempty"`
	Regions     []string     `json:"regions,omitempty"`
	Script      string       `json:"script,omitempty"`
	Item        *gofeed.Item `json:"-"`
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

// transform applies the provided function to every element in the provided array
func transform(c []*Content, f func(*Content) *Content) []*Content {
	vsf := make([]*Content, 0)
	for _, v := range c {
		vsf = append(vsf, f(v))
	}
	return vsf
}
