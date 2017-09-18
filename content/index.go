package content

import (
	"log"
	"path/filepath"
	"sync"
	"time"

	"golang.org/x/text/language"

	"strings"

	"github.com/blevesearch/bleve"
	"github.com/nu7hatch/gouuid"
)

// Index responsible for indexing content
type Index struct {
	id                   string
	allContent           []*Content
	content              map[string]*Content
	localizedContent     map[string][]*Content
	providers            map[string][]*Content
	providersLastUpdated map[string]time.Time
	languages            map[string][]*Content
	regions              map[string][]*Content
	scripts              map[string][]*Content
	tags                 map[string][]*Content
	fullText             bleve.Index
	mux                  sync.Mutex
}

// CreateIndex creates an index instance, using the provided file name and root directory
func CreateIndex(c Config) *Index {
	u, err := uuid.NewV4()
	if err != nil {
		log.Fatal("Failed to create index directory:", err)
	}

	var fullTextIndex bleve.Index
	if c.FullTextIndexActive() {
		indexPath := filepath.FromSlash(c.GetFullTextIndexDir() + "/" + u.String() + "/" + c.GetFullTextIndexFile())
		fullTextIndex, err = bleve.Open(indexPath)
		if err != nil {
			mapping := bleve.NewIndexMapping()
			fullTextIndex, err = bleve.New(indexPath, mapping)
			if err != nil {
				log.Fatal("Failed to create index: ", err)
			}
		}
	}

	return &Index{
		id:                   u.String(),
		allContent:           make([]*Content, 0),
		content:              make(map[string]*Content),
		localizedContent:     make(map[string][]*Content),
		providers:            make(map[string][]*Content),
		providersLastUpdated: make(map[string]time.Time),
		languages:            make(map[string][]*Content),
		regions:              make(map[string][]*Content),
		scripts:              make(map[string][]*Content),
		tags:                 make(map[string][]*Content),
		fullText:             fullTextIndex}
}

// createIndexWithID creates and empty index with the provided ID
func createIndexWithID(id string) *Index {
	return &Index{
		id:                   id,
		allContent:           make([]*Content, 0),
		content:              make(map[string]*Content),
		localizedContent:     make(map[string][]*Content),
		providers:            make(map[string][]*Content),
		providersLastUpdated: make(map[string]time.Time),
		languages:            make(map[string][]*Content),
		regions:              make(map[string][]*Content),
		scripts:              make(map[string][]*Content),
		tags:                 make(map[string][]*Content),
		fullText:             nil}
}

// Add adds the provided content items to this index
func (i *Index) Add(c []*Content) error {
	i.mux.Lock()
	defer i.mux.Unlock()

	for _, content := range c {
		err := i.AddItem(content)
		if err != nil {
			return err
		}
	}
	return nil
}

// AddItem adds a content item to this index. This method is not thread-safe.
func (i *Index) AddItem(c *Content) error {
	i.allContent = append(i.allContent, c)
	i.content[c.ID] = c

	// Index provider
	i.providers[c.Source] = append(i.providers[c.Source], c)

	// Index tags
	for _, tag := range c.Tags {
		lTag := strings.ToLower(tag)
		i.tags[lTag] = append(i.tags[lTag], c)
	}

	// Index lang/region/script
	if len(c.Regions) == 0 {
		i.regions["any"] = append(i.regions["any"], c)
	} else {
		for _, region := range c.Regions {
			indexLocaleValue(region, c, i.regions)
		}
	}
	indexLocaleValue(c.Language, c, i.languages)
	indexLocaleValue(c.Script, c, i.scripts)

	// Add to full-text index
	if i.fullText != nil {
		return i.fullText.Index(c.ID, c.Title+" "+c.Excerpt)
	}

	return nil
}

// Query index for content
func (i *Index) Query(q string) ([]*Content, error) {
	c := make([]*Content, 0)
	if i.fullText == nil {
		return c, nil
	}

	query := bleve.NewQueryStringQuery(q)
	searchRequest := bleve.NewSearchRequest(query)
	searchResult, err := i.fullText.Search(searchRequest)
	if searchResult != nil {
		for _, hit := range searchResult.Hits {
			hitc := i.content[hit.ID]
			if hitc != nil {
				c = append(c, hitc)
			}
		}
	}
	return c, err
}

// GetID returns the unique ID of this index
func (i *Index) GetID() string {
	return i.id
}

// GetContent returns all indexed content
func (i *Index) GetContent() []*Content {
	return i.allContent
}

// GetLocalizedContent returns indexed content matching the provided language, script and regions
func (i *Index) GetLocalizedContent(acceptLang string) []*Content {
	if i.localizedContent[acceptLang] != nil {
		return i.localizedContent[acceptLang]
	}

	tags, _, _ := language.ParseAcceptLanguage(acceptLang)
	if len(tags) == 0 {
		return i.allContent
	}

	c := make([]*Content, 0)

	langHits := i.languages["any"]
	regionHits := i.regions["any"]
	scriptHits := i.scripts["any"]
	for _, tag := range tags {
		b, _ := tag.Base()
		r, _ := tag.Region()
		s, _ := tag.Script()
		langHits = append(langHits, i.languages[strings.ToLower(b.String())]...)
		regionHits = append(regionHits, i.regions[strings.ToLower(r.String())]...)
		scriptHits = append(scriptHits, i.scripts[strings.ToLower(s.String())]...)
	}

	langHitMap := make(map[*Content]bool)
	regionHitMap := make(map[*Content]bool)
	for _, langHit := range langHits {
		langHitMap[langHit] = true
	}
	for _, regionHit := range regionHits {
		regionHitMap[regionHit] = true
	}
	for _, scriptHit := range scriptHits {
		if langHitMap[scriptHit] && regionHitMap[scriptHit] {
			c = append(c, scriptHit)
		}
	}

	return c
}

// PreLoadLocales builds up an index of localized content for the provided lang strings
func (i *Index) PreLoadLocales(acceptLang string) {
	for _, lang := range strings.Split(acceptLang, ",") {
		i.localizedContent[lang] = i.GetLocalizedContent(lang)
	}
}

// GetProviderLastUpdated returns the last updated time of the given provider
func (i *Index) GetProviderLastUpdated(provider string) time.Time {
	return i.providersLastUpdated[provider]
}

// SetProviderLastUpdated sets the last updated time of the given provider
func (i *Index) SetProviderLastUpdated(provider string) {
	i.providersLastUpdated[provider] = time.Now()
}

// GetProviderContent returns all indexed content from the given provider
func (i *Index) GetProviderContent(provider string) []*Content {
	return i.providers[provider]
}

// GetTaggedContent returns content containing the provided tag
func (i *Index) GetTaggedContent(tag string) []*Content {
	return i.tags[tag]
}

func indexLocaleValue(key string, val *Content, m map[string][]*Content) {
	k := strings.ToLower(key)
	if k == "" {
		m["any"] = append(m["any"], val)
	} else {
		m[k] = append(m[k], val)
	}
}
