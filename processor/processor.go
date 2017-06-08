package processor

import (
	"reflect"
	"strings"

	"log"
)

// Registry manages a global mapping of process names to types
type Registry struct {
	processors map[string]reflect.Type
}

// GetRegistry create a new processor registry
func GetRegistry() *Registry {
	r := &Registry{}
	r.processors = make(map[string]reflect.Type)
	r.processors["ExternalLinkRemover"] = reflect.TypeOf(ExternalLinkRemover{})
	r.processors["ImageExtractor"] = reflect.TypeOf(ImageExtractor{})
	r.processors["BoldElementRemover"] = reflect.TypeOf(BoldElementRemover{})
	return r
}

// GetNewProcessor returns the content processor with the given name
func (r *Registry) GetNewProcessor(name string) Processor {
	processor, ok := r.processors[name]
	if !ok {
		log.Fatal("Couldn't find content processor with name " + name)
	}

	return reflect.New(processor).Elem().Interface().(Processor)
}

// Context provides a processing context passed between processors
type Context struct {
	Content interface{}
	HTML    bool
	JSON    bool
	Result  map[string]string
}

// NewHTMLContext creates a new HTML specific processing context
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
	val, err := findAttributeValueOfFirstMatch(context, "img", "src")
	if err != nil {
		return nil, err
	}
	if !strings.HasPrefix(val, "http:") {
		val = "http:" + val
	}
	context.Result["image"] = val
	return context, nil
}
