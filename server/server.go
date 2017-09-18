package server

import (
	"net/http"
	"path/filepath"

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
)

// Server to host public API for importing and consuming content
type Server struct {
	// Index providing access to content, using an unsafe.Pointer to allow for atomic reference swaps
	index unsafe.Pointer
	// Array of configured content recommenders
	recommenders []content.Recommender
	// Reference to system config
	config *config.AppConfig
	// All configured content providers
	providers content.Providers
}

// JSONResponse wraps content recommendations as a JSON object
type JSONResponse struct {
	Recs content.Recommendations `json:"recommendations"`
}

// Create a new server instance
func Create(config *config.AppConfig, providers content.Providers, index *content.Index) *Server {
	recommenders := []content.Recommender{
		&content.TagBasedRecommender{},
		&content.QueryBasedRecommender{},
		&content.ProviderBasedRecommender{},
		&content.LocaleBasedRecommender{}}

	s := Server{index: unsafe.Pointer(index),
		recommenders: recommenders,
		config:       config,
		providers:    providers}

	http.HandleFunc(config.GetImportPath(), s.handleImport)
	http.HandleFunc(config.GetContentPath(), s.handleContent)
	return &s
}

// Start a server to provide an API for importing and consuming content
func (s *Server) Start() error {
	log.Printf("Server listening at %s\n", s.config.GetAddr())
	return http.ListenAndServe(s.config.GetAddr(), nil)
}

func (s *Server) handleImport(w http.ResponseWriter, r *http.Request) {
	apikey := strings.TrimSpace(strings.TrimLeft(r.Header.Get("Authorization"), "APIKEY"))

	provider, err := VerifyKey(apikey, s.config)
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

	err = content.Enqueue(s.config, body, provider)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Failed to enqueue content for indexing.\n"))
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (s *Server) handleContent(w http.ResponseWriter, req *http.Request) {
	index := s.getIndex()
	if match := req.Header.Get("If-None-Match"); match != "" {
		if match == index.GetID() {
			w.WriteHeader(http.StatusNotModified)
			return
		}
	}
	w.Header().Set("Access-Control-Allow-Origin", "*")

	c, hadErrors := s.produceRecommendations(req, index)
	if !hadErrors {
		w.Header().Set("Etag", index.GetID())
		w.Header().Set("Cache-Control", "max-age="+s.config.GetClientCacheMaxAge()+", must-revalidate")
	}

	format := req.URL.Query().Get("f")
	acceptHeader := req.Header.Get("Accept")
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

func (s *Server) produceRecommendations(r *http.Request, index *content.Index) (content.Recommendations, bool) {
	params := make(map[string]interface{})
	params["lang"] = r.Header.Get("Accept-Language")

	params["tags"] = r.URL.Query().Get("t")
	params["query"] = r.URL.Query().Get("q")
	params["provider"] = r.URL.Query().Get("p")
	params["locale"] = r.URL.Query().Get("l")

	recs := make(content.Recommendations, 0)
	cDedupe := make(map[string]bool)
	hadErrors := false
	for _, rec := range s.recommenders {
		crec, err := rec.Recommend(index, params)
		if err != nil {
			log.Printf("%v failed: %v\n", reflect.TypeOf(rec).Elem().Name(), err)
			hadErrors = true
			continue
		}
		for _, rec := range crec {
			if _, ok := cDedupe[rec.ID]; !ok {
				cDedupe[rec.ID] = true
				recs = append(recs, rec)
			}
		}

	}
	return recs, hadErrors
}

func (s *Server) respondWithHTML(w http.ResponseWriter, recs content.Recommendations) {
	t, err := template.ParseFiles(filepath.FromSlash(s.config.GetTemplateDir() + "/item.html"))
	if err != nil {
		log.Fatal("Failed to parse template: ", err)
	}

	w.Header().Set("Content-Type", "text/html;charset=UTF-8")
	for _, r := range recs {
		t.Execute(w, &r)
	}
}

func (s *Server) respondWithJSON(w http.ResponseWriter, recs content.Recommendations) {
	bytes, err := json.Marshal(JSONResponse{Recs: recs})
	if err != nil {
		log.Fatal("Failed to marshal content to JSON: ", err)
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(bytes)
}

// SetIndex atomically updates the server's index to reflect updated content
func (s *Server) SetIndex(index *content.Index) {
	atomic.StorePointer(&s.index, unsafe.Pointer(index))
}

func (s *Server) getIndex() *content.Index {
	return (*content.Index)(atomic.LoadPointer(&s.index))
}
