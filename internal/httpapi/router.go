package httpapi

import (
	"net/http"

	"github.com/vvvvgross/messenger/internal/auth"
)

func NewHandler(authService *auth.Service, tokens *auth.TokenManager) http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /healthz", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"status":"ok"}` + "\n"))
	})

	mux.HandleFunc("POST /auth/register", newRegisterHandler(authService))
	mux.HandleFunc("POST /auth/login", newLoginHandler(authService, tokens))

	return mux
}
