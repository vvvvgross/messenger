package httpapi

import (
	"context"
	"net/http"
	"strings"

	"github.com/vvvvgross/messenger/internal/auth"
)

type userIDContextKey struct{}

func userIDFromContext(ctx context.Context) (string, bool) {
	userID, ok := ctx.Value(userIDContextKey{}).(string)
	return userID, ok && userID != ""
}

func requireAuth(
	tokens *auth.TokenManager,
	next http.Handler,
) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		headers := r.Header.Values("Authorization")
		if len(headers) != 1 {
			writeUnauthorized(w)
			return
		}

		parts := strings.Fields(headers[0])
		if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
			writeUnauthorized(w)
			return
		}

		userID, err := tokens.Parse(parts[1])
		if err != nil {
			writeUnauthorized(w)
			return
		}

		ctx := context.WithValue(
			r.Context(),
			userIDContextKey{},
			userID,
		)

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func writeUnauthorized(w http.ResponseWriter) {
	w.Header().Set("WWW-Authenticate", "Bearer")
	w.Header().Set("Cache-Control", "no-store")

	writeJSON(w, http.StatusUnauthorized, errorResponse{
		Error: "authentication required",
	})
}
