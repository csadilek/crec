package processor

import (
	"reflect"
	"strings"

	"github.com/andybalholm/cascadia"

	"golang.org/x/net/html"
)

// Registry manages a global mapping of process names to types
type Registry struct {
	processors map[string]reflect.Type
}

// CreateRegistry create a new processor registry
func CreateRegistry() *Registry {
	r := &Registry{}
	r.processors = make(map[string]reflect.Type)
	r.processors["ExternalLinkRemover"] = reflect.TypeOf(ExternalLinkRemover{})
	r.processors["ImageExtractor"] = reflect.TypeOf(ImageExtractor{})
	r.processors["BoldElementRemover"] = reflect.TypeOf(BoldElementRemover{})
	return r
}

// GetNewProcessor returns the content processor with the given name
func (r *Registry) GetNewProcessor(name string) Processor {
	return reflect.New(r.processors[name]).Elem().Interface().(Processor)
}

// Context provides a processing context passed between processors
type Context struct {
	Content interface{}
	HTML    bool
	JSON    bool
	Result  map[string]string
}

// NewHTMLContext create a new HTML specific processing context
func NewHTMLContext(content interface{}) *Context {
	return &Context{Content: content, HTML: true, Result: make(map[string]string)}
}

// Processor applies content manipulation steps before ingestion
type Processor interface {
	Process(*Context) (*Context, error)
}

// ExternalLinkRemover removes links to external content (i.e. related articles)
type ExternalLinkRemover struct{}

// Process the provided content
func (p ExternalLinkRemover) Process(context *Context) (*Context, error) {
	return removeNodes(context, []string{"a"})
}

// BoldElementRemover removes bold elements
type BoldElementRemover struct{}

// Process the provided content
func (p BoldElementRemover) Process(context *Context) (*Context, error) {
	return removeNodes(context, []string{"b"})
}

// ImageExtractor processes content to find an image URI
type ImageExtractor struct{}

// Process the provided content
func (p ImageExtractor) Process(context *Context) (*Context, error) {
	if !context.HTML {
		return context, nil
	}
	node := context.Content.(*html.Node)
	img := cascadia.MustCompile("img").MatchFirst(node)
	if img != nil {
		for _, a := range img.Attr {
			if a.Key == "src" {
				if a.Val != "" {
					var src string
					if !strings.HasPrefix(a.Val, "http:") {
						src = "http:" + a.Val
					} else {
						src = a.Val
					}
					context.Result["image"] = src
					break
				}
			}
		}
	}
	return context, nil
}
