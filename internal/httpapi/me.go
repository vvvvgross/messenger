package httpapi

import "net/http"

type meResponse struct {
	ID string `json:"id"`
}

func meHandler(w http.ResponseWriter, r *http.Request) {
	userID, ok := userIDFromContext(r.Context())
	if !ok {
		writeUnauthorized(w)
		return
	}

	w.Header().Set("Cache-Control", "no-store")

	writeJSON(w, http.StatusOK, meResponse{
		ID: userID,
	})
}
