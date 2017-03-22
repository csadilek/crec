package server

import (
	"net/http"

	"encoding/json"
	"log"

	"html/template"
	"strings"

	"mozilla.org/crec/content"
	"mozilla.org/crec/ingester"
)

// Server to host public API for content consumption
type Server struct {
	Addr    string            // Address to start server e.g. ":8080"
	Path    string            // Path to bind handler function e.g. "/content"
	Indexer *ingester.Indexer // Aavailable content
}

// Start a server to provide an API for content consumption
func (s *Server) Start() {
	http.HandleFunc(s.Path, s.contentHandler)
	http.ListenAndServe(s.Addr, nil)
}

const minPageSize = 5

func (s *Server) contentHandler(w http.ResponseWriter, r *http.Request) {
	topic := r.URL.Query().Get("t")
	format := r.URL.Query().Get("f")
	query := r.URL.Query().Get("q")
	acceptHeader := r.Header.Get("Accept")

	w.Header().Set("Access-Control-Allow-Origin", "*")
	c := s.fetchContent(topic, format, query)
	if strings.Contains(acceptHeader, "html") && !strings.EqualFold(format, "json") {
		s.respondWithHTML(w, c)
	} else if strings.Contains(acceptHeader, "json") || strings.EqualFold(format, "json") {
		s.respondWithJSON(w, c)
	} else {
		log.Println("Invalid format requested.")
	}
}

func (s *Server) fetchContent(topic string, format string, query string) []*content.Content {
	var c []*content.Content
	if topic != "" {
		splits := strings.Split(topic, ",")
		splitMap := make(map[string]bool)
		for _, s := range splits {
			splitMap[strings.ToLower(s)] = true
		}

		c = content.Filter(s.Indexer.Content, func(c *content.Content) bool {
			for _, t := range c.Tags {
				if _, ok := splitMap[strings.ToLower(t)]; ok {
					return true
				}
			}
			return false
		})

		if len(c) < minPageSize {
			for _, topic := range splits {
				c = append(c, s.queryIndexForContent(topic)...)
			}
		}
	} else if query != "" {
		c = s.queryIndexForContent(query)
	} else {
		c = s.Indexer.Content
	}
	return c
}

func (s *Server) queryIndexForContent(q string) []*content.Content {
	c, err := s.Indexer.Query(q)
	if err != nil {
		log.Fatal("Failed to query index:", err)
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
