package httpapi

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestRoutes(t *testing.T) {
	for _, tt := range []struct {
		name, method, path string
		status             int
	}{
		{"health", "GET", "/healthz", 200},
		{"unsupported method", "POST", "/healthz", 405},
		{"unknown route", "GET", "/missing", 404},
		{"not a health subroute", "GET", "/healthz/extra", 404},
	} {
		t.Run(tt.name, func(t *testing.T) {
			response := httptest.NewRecorder()
			NewHandler(nil, nil, nil).ServeHTTP(response, httptest.NewRequest(tt.method, tt.path, nil))
			if response.Code != tt.status {
				t.Fatalf("status = %d, want %d", response.Code, tt.status)
			}
			if tt.status == http.StatusOK {
				if got := response.Header().Get("Content-Type"); got != "application/json" {
					t.Fatalf("content type = %q", got)
				}
				if got := response.Body.String(); got != "{\"status\":\"ok\"}\n" {
					t.Fatalf("body = %q", got)
				}
			}
		})
	}
}
