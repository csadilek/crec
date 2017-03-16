package content

import (
	"fmt"

	"github.com/mmcdole/gofeed"
)

// Content is just a simple wrapper of a gofeed Item for now. More to come...
type Content struct {
	Source string
	Item   *gofeed.Item
}

// RenderedContent is a content representation renderable as HTML
type RenderedContent struct {
	Title   string
	Link    string
	Summary string
}

// Title of this content item
func (c *Content) Title() string {
	return c.Item.Title
}

// Summary of this content item
func (c *Content) Summary() string {
	return c.Item.Description
}

// Tags for filtering content items
func (c *Content) Tags() []string {
	return c.Item.Categories
}

// Text of this content item aka the meat.
func (c *Content) Text() string {
	return c.Item.Content
}

// Image title and URL
func (c *Content) Image() (string, string) {
	return c.Item.Image.Title, c.Item.Image.URL
}

// Link to the full content item
func (c *Content) Link() string {
	return c.Item.Link
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
