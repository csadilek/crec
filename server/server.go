package server

import (
	"net/http"

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
	Addr    string         // Address to start server e.g. ":8080"
	Path    string         // Path to bind handler function e.g. "/content"
	indexer unsafe.Pointer // Indexer providing access to content
}

// Start a server to provide an API for content consumption
func (s *Server) Start(indexer *ingester.Indexer) {
	s.SetIndexer(indexer)
	http.HandleFunc(s.Path, s.contentHandler)
	http.ListenAndServe(s.Addr, nil)
}

const minPageSize = 5

func (s *Server) contentHandler(w http.ResponseWriter, r *http.Request) {
	if match := r.Header.Get("If-None-Match"); match != "" {
		if match == s.getIndexer().GetID() {
			w.WriteHeader(http.StatusNotModified)
			return
		}
	}
	w.Header().Set("Etag", s.getIndexer().GetID())
	w.Header().Set("Cache-Control", "max-age=120")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	tags := r.URL.Query().Get("t")
	format := r.URL.Query().Get("f")
	query := r.URL.Query().Get("q")
	acceptHeader := r.Header.Get("Accept")

	c := s.fetchContent(tags, format, query)
	if strings.Contains(acceptHeader, "html") && !strings.EqualFold(format, "json") {
		s.respondWithHTML(w, c)
	} else if strings.Contains(acceptHeader, "json") || strings.EqualFold(format, "json") {
		s.respondWithJSON(w, c)
	} else {
		w.WriteHeader(http.StatusNotAcceptable)
		w.Write([]byte("Media type" + acceptHeader + " not supported."))
	}
}

func (s *Server) fetchContent(tags string, format string, query string) []*content.Content {
	var c []*content.Content
	if tags != "" {
		var tagSplits []string
		var disjunction = true
		tagSplits = strings.Split(tags, ",")

		// TODO use a smarter query parser
		if !strings.Contains(tags, ",") && strings.Contains(tags, " ") {
			tagSplits = strings.Split(tags, " ")
			disjunction = false
		}

		tagMap := make(map[string]bool)
		for _, s := range tagSplits {
			tagMap[strings.TrimSpace(strings.ToLower(s))] = true
		}

		if disjunction {
			c = content.Filter(s.getIndexer().GetContent(), content.AnyTagFilter(tagMap))
		} else {
			c = content.Filter(s.getIndexer().GetContent(), content.AllTagFilter(tagMap))
		}

		// Not enough content based on tag matches -> find more using a full-text search
		if len(c) < minPageSize {
			for _, tag := range tagSplits {
				c = append(c, s.queryIndexForContent(tag)...)
			}
		}
	} else if query != "" {
		c = s.queryIndexForContent(query)
	}
	return content.Transform(c, func(item *content.Content) *content.Content {
		// TODO i18n and provide exact matches (this should all be based on a recommendation ext)
		if tags != "" {
			item.Explanation = "Selected for users interested in " + tags
		}
		return item
	})
}

func (s *Server) queryIndexForContent(q string) []*content.Content {
	c, err := s.getIndexer().Query(q)
	if err != nil {
		log.Fatal("Failed to query index:", err)
	}
	return c
}

//SetIndexer atomically updates the server's indexer to reflect updated content
func (s *Server) SetIndexer(indexer *ingester.Indexer) {
	atomic.StorePointer(&s.indexer, unsafe.Pointer(indexer))
}

func (s *Server) getIndexer() *ingester.Indexer {
	return (*ingester.Indexer)(atomic.LoadPointer(&s.indexer))
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
