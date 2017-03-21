package content

import (
	"fmt"

	"github.com/mmcdole/gofeed"
)

// Content is just a simple wrapper of a gofeed Item for now. More to come...
type Content struct {
	Source  string
	Title   string
	Link    string
	Image   string
	Summary string
	Item    *gofeed.Item `json:"-"`
}

// Tags for filtering content items
func (c *Content) Tags() []string {
	return c.Item.Categories
}

// Text of this content item aka the meat.
func (c *Content) Text() string {
	return c.Item.Content
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
