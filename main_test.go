package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
)

// writeConfigFile writes a JSON config to a temp file and returns the path.
func writeConfigFile(t *testing.T, routes []Route) string {
	t.Helper()
	data, err := json.Marshal(routes)
	if err != nil {
		t.Fatalf("failed to marshal routes: %v", err)
	}
	f, err := os.CreateTemp(t.TempDir(), "config-*.json")
	if err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}
	if _, err := f.Write(data); err != nil {
		t.Fatalf("failed to write config: %v", err)
	}
	if err := f.Close(); err != nil {
		t.Fatalf("failed to close temp file: %v", err)
	}
	return f.Name()
}

// TestNewProxyServer_ValidConfig verifies that a valid config file is loaded correctly.
func TestNewProxyServer_ValidConfig(t *testing.T) {
	routes := []Route{
		{Name: "svc-a", Path: "a", Target: "http://localhost:9001/hook"},
		{Name: "svc-b", Path: "b", Target: "http://localhost:9002/hook"},
	}
	path := writeConfigFile(t, routes)

	ps, err := NewProxyServer(path)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(ps.routes) != 2 {
		t.Fatalf("expected 2 routes, got %d", len(ps.routes))
	}
	if ps.routes[0].Name != "svc-a" {
		t.Errorf("expected first route name 'svc-a', got %q", ps.routes[0].Name)
	}
}

// TestNewProxyServer_MissingFile verifies an error is returned when the config file is absent.
func TestNewProxyServer_MissingFile(t *testing.T) {
	_, err := NewProxyServer(filepath.Join(t.TempDir(), "nonexistent.json"))
	if err == nil {
		t.Fatal("expected error for missing config file, got nil")
	}
}

// TestNewProxyServer_InvalidJSON verifies an error is returned for malformed JSON.
func TestNewProxyServer_InvalidJSON(t *testing.T) {
	f, err := os.CreateTemp(t.TempDir(), "bad-*.json")
	if err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}
	if _, err := f.WriteString("not-valid-json"); err != nil {
		t.Fatalf("failed to write invalid JSON: %v", err)
	}
	if err := f.Close(); err != nil {
		t.Fatalf("failed to close temp file: %v", err)
	}

	_, err = NewProxyServer(f.Name())
	if err == nil {
		t.Fatal("expected error for invalid JSON, got nil")
	}
}

// TestServeHTTP_NoRoutes verifies a 404 is returned when no routes are configured.
func TestServeHTTP_NoRoutes(t *testing.T) {
	ps := &ProxyServer{routes: []Route{}}

	req := httptest.NewRequest(http.MethodGet, "/anything", nil)
	rec := httptest.NewRecorder()
	ps.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Errorf("expected 404, got %d", rec.Code)
	}
}

// TestServeHTTP_ExactMatch verifies that an exact path match routes to the correct backend.
func TestServeHTTP_ExactMatch(t *testing.T) {
	backend := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer backend.Close()

	ps := &ProxyServer{routes: []Route{
		{Name: "backend", Path: "webhook", Target: backend.URL + "/hook"},
	}}

	req := httptest.NewRequest(http.MethodPost, "/webhook", nil)
	rec := httptest.NewRecorder()
	ps.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
}

// TestServeHTTP_PrefixMatch verifies that a sub-path of a route is still matched.
func TestServeHTTP_PrefixMatch(t *testing.T) {
	backend := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer backend.Close()

	ps := &ProxyServer{routes: []Route{
		{Name: "api", Path: "api", Target: backend.URL + "/"},
	}}

	req := httptest.NewRequest(http.MethodGet, "/api/v1/resource", nil)
	rec := httptest.NewRecorder()
	ps.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
}

// TestServeHTTP_NoPrefixPartialMatch ensures partial string matches do not route
// (e.g. path "api" must not match request "/apikey").
func TestServeHTTP_NoPrefixPartialMatch(t *testing.T) {
	ps := &ProxyServer{routes: []Route{
		{Name: "api", Path: "api", Target: "http://localhost:9999/"},
	}}

	req := httptest.NewRequest(http.MethodGet, "/apikey", nil)
	rec := httptest.NewRecorder()
	ps.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Errorf("expected 404, got %d", rec.Code)
	}
}

// TestServeHTTP_ExactMatchBeforePrefix verifies that an exact path match takes
// priority over a prefix match when both a shorter prefix route and an exact
// route are present in the route list (prefix listed first).
func TestServeHTTP_ExactMatchBeforePrefix(t *testing.T) {
	// prefixBackend handles requests matching the "api" prefix (e.g. /api/v1).
	prefixBackend := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Backend", "prefix")
		w.WriteHeader(http.StatusOK)
	}))
	defer prefixBackend.Close()

	// exactBackend should be reached only for the exact path /api-exact.
	exactBackend := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Backend", "exact")
		w.WriteHeader(http.StatusOK)
	}))
	defer exactBackend.Close()

	// "api" is a prefix route (matches /api/...) and "api-exact" is an exact route.
	ps := &ProxyServer{routes: []Route{
		{Name: "prefix", Path: "api", Target: prefixBackend.URL + "/"},
		{Name: "exact", Path: "api-exact", Target: exactBackend.URL + "/hook"},
	}}

	// A request to /api-exact must NOT be captured by the "api" prefix route.
	req := httptest.NewRequest(http.MethodGet, "/api-exact", nil)
	rec := httptest.NewRecorder()
	ps.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
	if got := rec.Header().Get("X-Backend"); got != "exact" {
		t.Errorf("expected X-Backend=exact, got %q", got)
	}
}

// TestProxyRequest_InvalidTargetURL verifies that an unparseable target URL
// results in a 500 response.
func TestProxyRequest_InvalidTargetURL(t *testing.T) {
	ps := &ProxyServer{}
	route := Route{Name: "bad", Path: "bad", Target: "://invalid-url"}

	req := httptest.NewRequest(http.MethodGet, "/bad", nil)
	rec := httptest.NewRecorder()
	ps.proxyRequest(rec, req, route)

	if rec.Code != http.StatusInternalServerError {
		t.Errorf("expected 500, got %d", rec.Code)
	}
}

// TestProxyRequest_HeadersForwarded verifies that request headers are forwarded to the backend.
func TestProxyRequest_HeadersForwarded(t *testing.T) {
	var receivedAuth string
	backend := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		receivedAuth = r.Header.Get("Authorization")
		w.WriteHeader(http.StatusOK)
	}))
	defer backend.Close()

	ps := &ProxyServer{}
	route := Route{Name: "svc", Path: "svc", Target: backend.URL + "/hook"}

	req := httptest.NewRequest(http.MethodPost, "/svc", nil)
	req.Header.Set("Authorization", "Bearer test-token")
	rec := httptest.NewRecorder()
	ps.proxyRequest(rec, req, route)

	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
	if receivedAuth != "Bearer test-token" {
		t.Errorf("expected Authorization header forwarded, got %q", receivedAuth)
	}
}
