package server

import (
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"net/http"

	"mozilla.org/crec/config"
	"mozilla.org/crec/ingester"
	"mozilla.org/crec/provider"
)

var indexer *ingester.Indexer
var server *Server

func TestMain(m *testing.M) {
	indexer = ingester.CreateIndexer(filepath.FromSlash(os.TempDir()+"/crec-test-index"), "test.bleve")
	config := config.Create("test-secret01234", "../template")
	server = Create(config, indexer, provider.Providers{"test": &provider.Provider{ID: "test"}})
	os.Exit(m.Run())
}

func TestHandleContentProcessesCacheHeaders(t *testing.T) {
	recorder := httptest.NewRecorder()
	request := httptest.NewRequest("GET", server.config.GetContentPath(), nil)

	request.Header.Set("If-None-Match", indexer.GetID())
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
	if recorder.Header().Get("Etag") != indexer.GetID() {
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

func TestHandleImportChecksAPIKey(t *testing.T) {
	recorder := httptest.NewRecorder()
	request := httptest.NewRequest("GET", server.config.GetImportPath(), nil)
	apikey := GenerateAPIKey("test", server.config)

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
