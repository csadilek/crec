package processor

import (
	"strings"
	"testing"

	"github.com/andybalholm/cascadia"

	"reflect"

	"golang.org/x/net/html"
)

func TestExternalLinkRemover(t *testing.T) {
	ctx, err := getContext("<html><body><div><a href=\"url\"></a></div><a href=\"url\"></a></body></html>")
	if err != nil {
		t.Error(err)
	}

	p := &ExternalLinkRemover{}
	ctx, err = p.Process(ctx)
	if err != nil {
		t.Error(err)
	}

	res := cascadia.MustCompile("a").MatchFirst(ctx.Content.(*html.Node))
	if res != nil {
		t.Error("Should not have found an anchor tag")
	}
}

func TestBoldElementRemover(t *testing.T) {
	ctx, err := getContext("<html><body><div><b>b</b></div><b>b</b></body></html>")
	if err != nil {
		t.Error(err)
	}

	p := &BoldElementRemover{}
	ctx, err = p.Process(ctx)
	if err != nil {
		t.Error(err)
	}

	res := cascadia.MustCompile("b").MatchFirst(ctx.Content.(*html.Node))
	if res != nil {
		t.Error("Should not have found a <b> tag")
	}
}

func TestImageExtractor(t *testing.T) {
	ctx, err := getContext("<html><body><div><img src=\"http://image-link\" /></div></body></html>")
	if err != nil {
		t.Error(err)
	}

	p := &ImageExtractor{}
	ctx, err = p.Process(ctx)
	if err != nil {
		t.Error(err)
	}

	imageURL := ctx.Result["image"]
	if imageURL == "" {
		t.Errorf("Expected to find image URL.")
	}
	if imageURL != "http://image-link" {
		t.Errorf("Expected to find image link, but found %v", imageURL)
	}
}

func TestGetRegistry(t *testing.T) {
	reg := GetRegistry()
	for k, v := range reg.processors {
		processor := reg.GetNewProcessor(k)
		if processor == nil {
			t.Errorf("Failed to instantiate content processor with name: %v", processor)
		}
		got := reflect.TypeOf(processor)
		if got != v {
			t.Errorf("Processor has wrong type, expected %v, but got %v", v, got.Name())
		}
	}
}

func getContext(content string) (*Context, error) {
	r := strings.NewReader(content)
	doc, err := html.Parse(r)
	if err != nil {
		return nil, err
	}

	return NewHTMLContext(doc), nil
}
