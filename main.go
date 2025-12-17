package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"strings"
)

// Route represents a single proxy route configuration
type Route struct {
	Name   string `json:"name"`
	Path   string `json:"path"`
	Target string `json:"target"`
}

// ProxyServer holds the configuration and routes
type ProxyServer struct {
	routes []Route
}

// NewProxyServer creates a new proxy server from a config file
func NewProxyServer(configPath string) (*ProxyServer, error) {
	file, err := os.Open(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open config file: %w", err)
	}
	defer file.Close()

	data, err := io.ReadAll(file)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var routes []Route
	if err := json.Unmarshal(data, &routes); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	return &ProxyServer{routes: routes}, nil
}

// ServeHTTP handles incoming requests and forwards them to the appropriate target
func (ps *ProxyServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// Remove leading slash for comparison
	requestPath := strings.TrimPrefix(r.URL.Path, "/")
	
	// Find matching route - prioritize exact matches, then prefix matches
	var matchedRoute *Route
	for i := range ps.routes {
		route := &ps.routes[i]
		// Exact match - highest priority
		if requestPath == route.Path {
			matchedRoute = route
			break
		}
		// Prefix match with slash boundary (e.g., 'api' matches 'api/v1' but not 'apikey')
		if strings.HasPrefix(requestPath, route.Path+"/") {
			matchedRoute = route
			break
		}
	}
	
	if matchedRoute != nil {
		ps.proxyRequest(w, r, *matchedRoute)
		return
	}

	// No route found
	http.Error(w, "Not Found", http.StatusNotFound)
	log.Printf("No route found for path: %s", r.URL.Path)
}

// proxyRequest forwards the request to the target URL
func (ps *ProxyServer) proxyRequest(w http.ResponseWriter, r *http.Request, route Route) {
	targetURL, err := url.Parse(route.Target)
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		log.Printf("Failed to parse target URL for route %s: %v", route.Name, err)
		return
	}

	// Create a reverse proxy
	proxy := httputil.NewSingleHostReverseProxy(targetURL)
	
	// Customize the director to forward to the exact target URL
	// Note: This replaces the request path with the target path (not appending subpaths)
	// Example: /github-webhook -> http://localhost:8000/webhook (not /webhook/github-webhook)
	originalDirector := proxy.Director
	proxy.Director = func(req *http.Request) {
		originalDirector(req)
		req.Host = targetURL.Host
		req.URL.Scheme = targetURL.Scheme
		req.URL.Host = targetURL.Host
		req.URL.Path = targetURL.Path
		req.URL.RawQuery = r.URL.RawQuery
	}

	// Log the request
	log.Printf("Proxying request: %s %s -> %s", r.Method, r.URL.Path, route.Target)

	// Proxy the request
	proxy.ServeHTTP(w, r)
}

func main() {
	// Get config file path from environment or use default
	configPath := os.Getenv("CONFIG_PATH")
	if configPath == "" {
		configPath = "config.json"
	}

	// Get port from environment or use default
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	// Create proxy server
	server, err := NewProxyServer(configPath)
	if err != nil {
		log.Fatalf("Failed to create proxy server: %v", err)
	}

	// Log loaded routes
	log.Printf("InnerGate reverse proxy starting on port %s", port)
	log.Printf("Loaded %d routes:", len(server.routes))
	for _, route := range server.routes {
		log.Printf("  - /%s -> %s", route.Path, route.Target)
	}

	// Start server
	addr := fmt.Sprintf(":%s", port)
	log.Printf("Server listening on %s", addr)
	if err := http.ListenAndServe(addr, server); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}
