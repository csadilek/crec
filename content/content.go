package content

import (
	"fmt"
	"strings"
	"time"
)

// Type allows labelling content for client-side display (e.g. ordering) purposes
type Type string

const (
	// RECOMMENDED is the default content type
	RECOMMENDED Type = "recommended"
	// PROMOTED indicates content which is boosted based on popularity
	PROMOTED Type = "promoted"
	// SPONSORED indicates partner content
	SPONSORED Type = "sponsored"
)

// Content represents our unified data model abstracting from the various formats used by providers.
type Content struct {
	// Globally unique content identifier
	ID string `json:"id,omitempty"`

	// Identifier of the content provider
	Source string `json:"source,omitempty"`

	// Title provided for this content
	Title string `json:"title,omitempty"`

	// URL to a web site showing a detailed view of the content
	URL string `json:"url,omitempty"`

	// Image URI to a preview image
	Image string `json:"image_src,omitempty"`

	// Excerpt of the content
	Excerpt string `json:"excerpt,omitempty"`

	// HTML view of the content
	HTML string `json:"-"`

	// Explanation as to why the content was recommended to a specific client
	Explanation string `json:"explanation,omitempty"`

	// Author of the content
	Author string `json:"author,omitempty"`

	// Publication date
	Published string `json:"published_timestamp,omitempty"`

	// Tags and categories applied to this content
	Tags []string `json:"tags,omitempty"`

	// Language the content is written in
	Language string `json:"-"`

	// Regions the content is applicable to
	Regions []string `json:"-"`

	// Script the content is written in
	Script string `json:"-"`

	// Specifies the domain similarities of this content. The domain
	// name is used as key, the weight as value. This can be used
	// to map content to specific user interests i.e. based on
	// their browsing history.
	Domains map[string]float32 `json:"domain_affinities,omitempty"`

	// Specifies the content type
	CType Type `json:"type,omitempty"`
}

func (c *Content) String() string {
	return fmt.Sprintf("Source: %s: Title: %s", c.Source, c.Title)
}

// Config contract for objects holding all content-related settings
type Config interface {
	GetFullTextIndexDir() string
	GetFullTextIndexFile() string
	GetImportQueueDir() string
	GetIndexRefreshInterval() time.Duration
	GetLocales() string
	GetProviderRegistryDir() string
	FullTextIndexActive() bool
}

// Filter content using the provided predicate function
func Filter(c []*Content, f func(*Content) bool) []*Content {
	vsf := make([]*Content, 0)
	for _, v := range c {
		if f(v) {
			vsf = append(vsf, v)
		}
	}
	return vsf
}

// AnyTagFilter returns a filter function which retains the content if any
// of the provided tags is present
func AnyTagFilter(tags map[string]bool) func(*Content) bool {
	return func(c *Content) bool {
		for _, t := range c.Tags {
			if _, ok := tags[strings.ToLower(t)]; ok {
				return true
			}
		}
		return false
	}
}

// AllTagFilter returns a filter functions which retains the content if all
// of the provided tags are present
func AllTagFilter(tags []string) func(*Content) bool {
	return func(c *Content) bool {
		tagMap := make(map[string]bool)
		for _, tag := range c.Tags {
			tagMap[strings.TrimSpace(tag)] = true
		}
		for _, t := range tags {
			if _, ok := tagMap[t]; !ok {
				return false
			}
		}
		return true
	}
}

// Transform applies the provided function to a copy of every element in the provided array
func Transform(c []*Content, f func(Content) *Content) []*Content {
	vsf := make([]*Content, 0)
	for _, v := range c {
		vsf = append(vsf, f(*v))
	}
	return vsf
}
