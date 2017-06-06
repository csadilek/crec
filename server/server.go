package server

import (
	"net/http"
	"path/filepath"

	"golang.org/x/text/language"

	"encoding/json"
	"log"

	"html/template"
	"strings"

	"unsafe"

	"sync/atomic"

	"io/ioutil"

	"reflect"

	"mozilla.org/crec/config"
	"mozilla.org/crec/content"
	"mozilla.org/crec/ingester"
	"mozilla.org/crec/provider"
)

// Server to host public API for content consumption
type Server struct {
	index        unsafe.Pointer        // Index providing access to content
	recommenders []content.Recommender // Array of configured content recommenders
	config       *config.Config        // Reference to system config
	providers    provider.Providers    // All configured content providers
}

// Create a new server instance
func Create(config *config.Config, providers provider.Providers, index *ingester.Index) *Server {
	s := Server{}
	s.config = config
	s.SetIndex(index)
	s.providers = providers
	s.configureRecommenders()

	http.HandleFunc(config.GetImportPath(), s.handleImport)
	http.HandleFunc(config.GetContentPath(), s.handleContent)
	return &s
}

// Start a server which provides an API for content consumption
func (s *Server) Start() error {
	log.Printf("Server listening at %s\n", s.config.GetAddr())
	return http.ListenAndServe(s.config.GetAddr(), nil)
}

func (s *Server) handleImport(w http.ResponseWriter, r *http.Request) {
	apikey := strings.TrimSpace(strings.TrimLeft(r.Header.Get("Authorization"), "APIKEY"))

	provider, err := GetProviderForAPIKey(apikey, s.config)
	if err != nil {
		w.WriteHeader(http.StatusForbidden)
		return
	}

	_, ok := s.providers[provider]
	if !ok {
		w.WriteHeader(http.StatusForbidden)
		return
	}

	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Failed to read request body.\n"))
		return
	}

	err = ingester.Queue(s.config, body, provider)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Failed to enqueue content for indexing.\n"))
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (s *Server) handleContent(w http.ResponseWriter, r *http.Request) {
	if match := r.Header.Get("If-None-Match"); match != "" {
		if match == s.getIndex().GetID() {
			w.WriteHeader(http.StatusNotModified)
			return
		}
	}

	w.Header().Set("Access-Control-Allow-Origin", "*")

	c, hadErrors := s.produceRecommendations(r)
	if !hadErrors {
		w.Header().Set("Etag", s.getIndex().GetID())
		w.Header().Set("Cache-Control", "max-age="+s.config.GetClientCacheMaxAge()+", must-revalidate")
	}

	format := r.URL.Query().Get("f")
	acceptHeader := r.Header.Get("Accept")
	if strings.Contains(acceptHeader, "html") && !strings.EqualFold(format, "json") {
		s.respondWithHTML(w, c)
	} else if strings.Contains(acceptHeader, "json") ||
		strings.HasSuffix(acceptHeader, "*") ||
		strings.EqualFold(format, "json") {
		s.respondWithJSON(w, c)
	} else {
		w.WriteHeader(http.StatusNotAcceptable)
		w.Write([]byte("Media type " + acceptHeader + " not supported.\n"))
	}
}

func (s *Server) produceRecommendations(r *http.Request) ([]*content.Content, bool) {
	tags, _, _ := language.ParseAcceptLanguage(r.Header.Get("Accept-Language"))

	params := make(map[string]string)
	params["tags"] = r.URL.Query().Get("t")
	params["query"] = r.URL.Query().Get("q")
	params["provider"] = r.URL.Query().Get("p")

	c := make([]*content.Content, 0)
	cDedupe := make(map[string]bool)
	hadErrors := false
	for _, rec := range s.recommenders {
		crec, err := rec.Recommend(s.getIndex().GetLocalizedContent(tags), params)
		if err != nil {
			log.Printf("%v failed: %v\n", reflect.TypeOf(rec).Elem().Name(), err)
			hadErrors = true
			continue
		}
		for _, rec := range crec {
			if _, ok := cDedupe[rec.ID]; !ok {
				cDedupe[rec.ID] = true
				c = append(c, rec)
			}
		}

	}
	return c, hadErrors
}

func (s *Server) respondWithHTML(w http.ResponseWriter, c []*content.Content) {
	t, err := template.ParseFiles(filepath.FromSlash(s.config.GetTemplateDir() + "/item.html"))
	if err != nil {
		log.Fatal("Failed to parse template: ", err)
	}

	w.Header().Set("Content-Type", "text/html;charset=UTF-8")
	for _, r := range c {
		t.Execute(w, &r)
	}
}

func (s *Server) respondWithJSON(w http.ResponseWriter, c []*content.Content) {
	bytes, err := json.Marshal(c)
	if err != nil {
		log.Fatal("Failed to marshal content to JSON: ", err)
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(bytes)
}

//SetIndex atomically updates the server's index to reflect updated content
func (s *Server) SetIndex(index *ingester.Index) {
	atomic.StorePointer(&s.index, unsafe.Pointer(index))
}

func (s *Server) getIndex() *ingester.Index {
	return (*ingester.Index)(atomic.LoadPointer(&s.index))
}

func (s *Server) configureRecommenders() {
	tagBasedRecommender := &content.TagBasedRecommender{}

	queryBasedRecommender := &content.QueryBasedRecommender{
		Search: func(q string) ([]*content.Content, error) {
			return s.getIndex().Query(q)
		}}

	providerBasedRecommender := &content.ProviderBasedRecommender{
		Search: func(provider string) []*content.Content {
			return s.getIndex().GetProviderContent(provider)
		}}

	s.recommenders = []content.Recommender{
		tagBasedRecommender,
		queryBasedRecommender,
		providerBasedRecommender}
}
