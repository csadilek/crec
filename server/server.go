package server

import (
	"fmt"
	"net/http"
	"net/url"

	"encoding/json"
	"log"

	"html/template"
	"strings"

	"unsafe"

	"sync/atomic"

	"io/ioutil"

	"mozilla.org/crec/config"
	"mozilla.org/crec/content"
	"mozilla.org/crec/ingester"
	"mozilla.org/crec/provider"
)

// Server to host public API for content consumption
type Server struct {
	indexer      unsafe.Pointer        // Indexer providing access to content
	recommenders []content.Recommender // Array of configured content recommenders
	config       *config.Config        // Reference to system config
	providers    provider.Providers    // All configured content providers
}

// Start a server which provides an API for content consumption
func (s *Server) Start(config *config.Config, indexer *ingester.Indexer, providers provider.Providers) {
	s.config = config
	s.SetIndexer(indexer)
	s.providers = providers
	s.configureRecommenders()

	http.HandleFunc(config.GetImportPath(), s.importContentHandler)
	http.HandleFunc(config.GetContentPath(), s.contentHandler)
	fmt.Printf("Server listening at %s\n", config.GetAddr())
	err := http.ListenAndServe(config.GetAddr(), nil)
	if err != nil {
		log.Fatal("Server failed to start: ", err)
	}
}

func (s *Server) importContentHandler(w http.ResponseWriter, r *http.Request) {
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Failed to read body of request.\n"))
		return
	}

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

	err = ingester.Queue(s.config, body, provider)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Failed to enqueue content for indexing.\n"))
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (s *Server) contentHandler(w http.ResponseWriter, r *http.Request) {
	if match := r.Header.Get("If-None-Match"); match != "" {
		if match == s.getIndexer().GetID() {
			w.WriteHeader(http.StatusNotModified)
			return
		}
	}

	w.Header().Set("Etag", s.getIndexer().GetID())
	w.Header().Set("Cache-Control", "max-age="+s.config.GetClientCacheMaxAge()+", must-revalidate")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	c := s.produceRecommendations(r.URL.Query())

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

func (s *Server) produceRecommendations(values url.Values) []*content.Content {
	params := make(map[string]string)
	params["tags"] = values.Get("t")
	params["query"] = values.Get("q")

	c := make([]*content.Content, 0)
	for _, rec := range s.recommenders {
		crec, err := rec.Recommend(s.getIndexer().GetContent(), params)
		if err != nil {
			log.Println("Recommender failed: ", err)
			continue
		}
		c = append(c, crec...)
	}
	return c
}

func (s *Server) respondWithHTML(w http.ResponseWriter, c []*content.Content) {
	t, err := template.ParseFiles("template/item.html")
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

//SetIndexer atomically updates the server's indexer to reflect updated content
func (s *Server) SetIndexer(indexer *ingester.Indexer) {
	atomic.StorePointer(&s.indexer, unsafe.Pointer(indexer))
}

func (s *Server) getIndexer() *ingester.Indexer {
	return (*ingester.Indexer)(atomic.LoadPointer(&s.indexer))
}

func (s *Server) configureRecommenders() {
	tagBasedRecommender := &content.TagBasedRecommender{}
	queryBasedRecommender := &content.QueryBasedRecommender{
		Search: func(q string) ([]*content.Content, error) { return s.getIndexer().Query(q) }}

	s.recommenders = []content.Recommender{tagBasedRecommender, queryBasedRecommender}
}
