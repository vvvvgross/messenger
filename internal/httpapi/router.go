package httpapi

import (
	"net/http"

	"github.com/vvvvgross/messenger/internal/auth"
	"github.com/vvvvgross/messenger/internal/chat"
)

func NewHandler(authService *auth.Service, tokens *auth.TokenManager, chatService *chat.Service) http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /healthz", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"status":"ok"}` + "\n"))
	})

	mux.HandleFunc("POST /auth/register", newRegisterHandler(authService))
	mux.HandleFunc("POST /auth/login", newLoginHandler(authService, tokens))
	mux.Handle("GET /auth/me", requireAuth(tokens, http.HandlerFunc(meHandler)))
	mux.Handle("POST /chats/group", requireAuth(tokens, newCreateGroupHandler(chatService)))

	return mux
}
