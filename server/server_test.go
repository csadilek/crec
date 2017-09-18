package server

import (
	"encoding/json"
	"errors"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strconv"
	"testing"

	"net/http"

	"mozilla.org/crec/config"
	"mozilla.org/crec/content"
)

var index *content.Index
var server *Server

func TestMain(m *testing.M) {
	config := config.Create(
		"test-secret01234",
		"../template",
		filepath.FromSlash(os.TempDir()+"/import"),
		filepath.FromSlash(os.TempDir()+"/crec-test-index"),
		"test.bleve")

	index = content.CreateIndex(config)
	server = Create(config, content.Providers{"test": &content.Provider{ID: "test"}}, index)
	os.Exit(m.Run())
}
func TestHandleContentProcessesCacheHeaders(t *testing.T) {
	recorder := httptest.NewRecorder()
	request := httptest.NewRequest("GET", server.config.GetContentPath(), nil)

	request.Header.Set("If-None-Match", index.GetID())
	server.handleContent(recorder, request)
	if recorder.Code != http.StatusNotModified {
		t.Errorf("Expected 304 (Not Modified), but got %v", recorder.Code)
	}

	recorder = httptest.NewRecorder()
	request.Header.Set("If-None-Match", "no-match")
	request.Header.Set("Accept", "application/json")
	server.handleContent(recorder, request)
	if recorder.Code != http.StatusOK {
		t.Errorf("Expected 200 (OK), but got %v", recorder.Code)
	}
	if recorder.Header().Get("Etag") != index.GetID() {
		t.Error("Expected Etag to be set")
	}
	if recorder.Header().Get("Cache-Control") != "max-age="+server.config.GetClientCacheMaxAge()+", must-revalidate" {
		t.Errorf("Unexpected Cache-Control header: %v", recorder.Header().Get("Cache-Control"))
	}
}
func TestHandleContentProcessesAcceptHeaders(t *testing.T) {
	recorder := httptest.NewRecorder()
	request := httptest.NewRequest("GET", server.config.GetContentPath(), nil)

	server.handleContent(recorder, request)
	if recorder.Code != http.StatusNotAcceptable {
		t.Errorf("Expected 406 (Not Acceptable), but got %v", recorder.Code)
	}

	request.Header.Set("Accept", "foo/bar")
	recorder = httptest.NewRecorder()
	server.handleContent(recorder, request)
	if recorder.Code != http.StatusNotAcceptable {
		t.Errorf("Expected 406 (Not Acceptable), but got %v", recorder.Code)
	}

	request.Header.Set("Accept", "text/html")
	recorder = httptest.NewRecorder()
	server.handleContent(recorder, request)
	if recorder.Code != http.StatusOK {
		t.Errorf("Expected 200 (OK), but got %v", recorder.Code)
	}

	request.Header.Set("Accept", "application/json")
	recorder = httptest.NewRecorder()
	server.handleContent(recorder, request)
	if recorder.Code != http.StatusOK {
		t.Errorf("Expected 200 (OK), but got %v", recorder.Code)
	}

	request.Header.Set("Accept", "application/*")
	recorder = httptest.NewRecorder()
	server.handleContent(recorder, request)
	if recorder.Code != http.StatusOK {
		t.Errorf("Expected 200 (OK), but got %v", recorder.Code)
	}
}
func TestHandleContentProducesRecommendations(t *testing.T) {
	index.AddItem(&content.Content{ID: "0", Tags: []string{"t1"}})
	index.AddItem(&content.Content{ID: "1", Excerpt: "q1"})
	index.AddItem(&content.Content{ID: "2", Source: "p1"})

	recorder := httptest.NewRecorder()
	request := httptest.NewRequest("GET", server.config.GetContentPath()+"?t=t1&q=q1&p=p1", nil)

	request.Header.Set("Accept", "application/json")
	server.handleContent(recorder, request)
	if recorder.Code != http.StatusOK {
		t.Errorf("Expected 200 (OK), but got %v", recorder.Code)
	}

	response := JSONResponse{}
	err := json.Unmarshal(recorder.Body.Bytes(), &response)
	if err != nil {
		t.Error(err)
	}

	content := response.Recs
	if len(content) != 3 {
		t.Errorf("Expected exactly 3 recommendations, but got %v", len(content))
	}

	for index := range content {
		if content[index].ID != strconv.Itoa(index) {
			t.Errorf("Expected content with ID %v, but got %v", index, content[index].ID)
		}
	}
}
func TestHandleContentProducesUniqueRecommendations(t *testing.T) {
	index.AddItem(&content.Content{ID: "0", Tags: []string{"t1"}, Excerpt: "q1"})
	index.AddItem(&content.Content{ID: "1", Excerpt: "q1"})
	index.AddItem(&content.Content{ID: "2", Source: "p1", Excerpt: "q1"})

	recorder := httptest.NewRecorder()
	request := httptest.NewRequest("GET", server.config.GetContentPath()+"?t=t1&q=q1&p=p1", nil)

	request.Header.Set("Accept", "application/json")
	server.handleContent(recorder, request)
	if recorder.Code != http.StatusOK {
		t.Errorf("Expected 200 (OK), but got %v", recorder.Code)
	}

	response := JSONResponse{}
	err := json.Unmarshal(recorder.Body.Bytes(), &response)
	if err != nil {
		t.Error(err)
	}

	content := response.Recs
	if len(content) != 3 {
		t.Errorf("Expected exactly 3 recommendations, but got %v", len(content))
	}

	for index := range content {
		if content[index].ID != strconv.Itoa(index) {
			t.Errorf("Expected content with ID %v, but got %v", index, content[index].ID)
		}
	}
}
func TestHandleImportChecksAPIKey(t *testing.T) {
	recorder := httptest.NewRecorder()
	request := httptest.NewRequest("GET", server.config.GetImportPath(), nil)
	apikey := GenerateKey("test", server.config)

	server.handleImport(recorder, request)
	if recorder.Code != http.StatusForbidden {
		t.Errorf("Expected 403 (Status Forbidden), but got %v", recorder.Code)
	}

	request.Header.Set("Authorization", "APIKEY foo")
	recorder = httptest.NewRecorder()
	server.handleImport(recorder, request)
	if recorder.Code != http.StatusForbidden {
		t.Errorf("Expected 403 (Status Forbidden), but got %v", recorder.Code)
	}

	request.Header.Set("Authorization", "APIKEY "+apikey)
	recorder = httptest.NewRecorder()
	server.handleImport(recorder, request)
	if recorder.Code != http.StatusOK {
		t.Errorf("Expected 200 (Status OK), but got %v", recorder.Code)
	}
}

type FailingRecommender struct{}

func (r *FailingRecommender) Recommend(
	index *content.Index,
	params map[string]interface{}) (content.Recommendations, error) {

	return nil, errors.New("Expected error for testing purposes")
}

func TestCacheHeadersOmittedIfRecommenderFailing(t *testing.T) {
	failingRecommender := &FailingRecommender{}
	server.recommenders = append(server.recommenders, failingRecommender)

	recorder := httptest.NewRecorder()
	request := httptest.NewRequest("GET", server.config.GetContentPath()+"?t=t1&q=q1&p=p1", nil)

	request.Header.Set("If-None-Match", "no-match")
	request.Header.Set("Accept", "application/json")
	server.handleContent(recorder, request)
	if recorder.Code != http.StatusOK {
		t.Errorf("Expected 200 (OK), but got %v", recorder.Code)
	}
	if recorder.Header().Get("Etag") != "" {
		t.Error("Expected Etag to be empty")
	}
	if recorder.Header().Get("Cache-Control") != "" {
		t.Errorf("Expected Cache-Control header to be empty")
	}

	server.recommenders = make([]content.Recommender, 0)
}

func BenchmarkHandleContent(b *testing.B) {
	for i := 0; i < b.N; i++ {
		recorder := httptest.NewRecorder()
		request := httptest.NewRequest("GET", server.config.GetContentPath()+"?t=t1&q=q1&p=p1", nil)
		server.handleContent(recorder, request)
	}
}
