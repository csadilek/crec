package server

import (
	"net/http"
	"net/url"

	"encoding/json"
	"log"

	"html/template"
	"strings"

	"unsafe"

	"sync/atomic"

	"mozilla.org/crec/content"
	"mozilla.org/crec/ingester"
)

// Server to host public API for content consumption
type Server struct {
	Addr         string                // Address to start server e.g. ":8080"
	Path         string                // Path to bind handler function e.g. "/content"
	indexer      unsafe.Pointer        // Indexer providing access to content
	recommenders []content.Recommender // Array of available recommenders
}

// Start a server which provides an API for content consumption
func (s *Server) Start(indexer *ingester.Indexer) {
	s.SetIndexer(indexer)
	s.setRecommenders()
	http.HandleFunc(s.Path, s.contentHandler)
	http.ListenAndServe(s.Addr, nil)
}

func (s *Server) contentHandler(w http.ResponseWriter, r *http.Request) {
	if match := r.Header.Get("If-None-Match"); match != "" {
		if match == s.getIndexer().GetID() {
			w.WriteHeader(http.StatusNotModified)
			return
		}
	}

	w.Header().Set("Etag", s.getIndexer().GetID())
	w.Header().Set("Cache-Control", "max-age=120, must-revalidate")
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
			log.Println("Recommender problem: ", err)
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

func (s *Server) setRecommenders() {
	tagBasedRecommender := &content.TagBasedRecommender{}
	queryBasedRecommender := &content.QueryBasedRecommender{
		Search: func(q string) ([]*content.Content, error) { return s.getIndexer().Query(q) }}

	s.recommenders = []content.Recommender{tagBasedRecommender, queryBasedRecommender}
}
