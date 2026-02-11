package testutil

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
)

// MockSpotifyServer provides a mock HTTP server for testing Spotify API interactions.
type MockSpotifyServer struct {
	*httptest.Server
	handlers map[string]http.HandlerFunc
}

// NewMockSpotifyServer creates a new mock Spotify server.
func NewMockSpotifyServer() *MockSpotifyServer {
	m := &MockSpotifyServer{
		handlers: make(map[string]http.HandlerFunc),
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		path := r.URL.Path
		if handler, ok := m.handlers[path]; ok {
			handler(w, r)
			return
		}
		http.NotFound(w, r)
	})

	m.Server = httptest.NewServer(mux)
	return m
}

// AddHandler registers a handler for a specific path.
func (m *MockSpotifyServer) AddHandler(path string, handler http.HandlerFunc) {
	m.handlers[path] = handler
}

// RespondJSON writes a JSON response with the given status code.
func RespondJSON(w http.ResponseWriter, status int, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}

// RespondError writes a Spotify API error response.
func RespondError(w http.ResponseWriter, status int, message string) {
	RespondJSON(w, status, map[string]interface{}{
		"error": map[string]interface{}{
			"status":  status,
			"message": message,
		},
	})
}

// CheckAuthorization verifies the Authorization header contains the expected token.
// Returns true if valid, writes 401 response and returns false if invalid.
func CheckAuthorization(w http.ResponseWriter, r *http.Request, expectedToken string) bool {
	auth := r.Header.Get("Authorization")
	if auth != "Bearer "+expectedToken {
		w.WriteHeader(http.StatusUnauthorized)
		RespondError(w, http.StatusUnauthorized, "Invalid token")
		return false
	}
	return true
}
