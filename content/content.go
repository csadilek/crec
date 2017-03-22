package content

import (
	"fmt"

	"github.com/mmcdole/gofeed"
)

// Content is just a simple wrapper of a gofeed Item for now. More to come...
type Content struct {
	ID        string       `json:"id,omitempty"`
	Source    string       `json:"source,omitempty"`
	Title     string       `json:"title,omitempty"`
	Link      string       `json:"link,omitempty"`
	Image     string       `json:"image,omitempty"`
	Summary   string       `json:"summary,omitempty"`
	HTML      string       `json:"html,omitempty"`
	Author    string       `json:"author,omitempty"`
	Published string       `json:"published,omitempty"`
	Tags      []string     `json:"tags,omitempty"`
	Item      *gofeed.Item `json:"-"`
}

func (c *Content) String() string {
	return fmt.Sprintf("%s: %s", c.Source, c.Item.Title)
}

// Filter the content using the provided predicate function
func Filter(c []*Content, f func(*Content) bool) []*Content {
	vsf := make([]*Content, 0)
	for _, v := range c {
		if f(v) {
			vsf = append(vsf, v)
		}
	}
	return vsf
}
