package httpapi

import (
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"net/url"
	"strconv"

	"github.com/vvvvgross/messenger/internal/chat"
)

func newListChatsHandler(service *chat.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Cache-Control", "no-store")

		userID, ok := userIDFromContext(r.Context())
		if !ok {
			writeUnauthorized(w)
			return
		}

		query, err := url.ParseQuery(r.URL.RawQuery)
		if err != nil {
			writeJSON(w, http.StatusBadRequest, errorResponse{
				Error: "invalid query parameters",
			})
			return
		}

		limit, err := parseChatListInt(query, "limit", 20)
		if err != nil {
			writeJSON(w, http.StatusBadRequest, errorResponse{
				Error: err.Error(),
			})
			return
		}

		offset, err := parseChatListInt(query, "offset", 0)
		if err != nil {
			writeJSON(w, http.StatusBadRequest, errorResponse{
				Error: err.Error(),
			})
			return
		}

		chats, err := service.ListByUser(
			r.Context(),
			userID,
			limit,
			offset,
		)
		if err != nil {
			switch {
			case errors.Is(err, chat.ErrInvalidListLimit):
				writeJSON(w, http.StatusBadRequest, errorResponse{
					Error: chat.ErrInvalidListLimit.Error(),
				})
			case errors.Is(err, chat.ErrInvalidListOffset):
				writeJSON(w, http.StatusBadRequest, errorResponse{
					Error: chat.ErrInvalidListOffset.Error(),
				})
			default:
				if r.Context().Err() != nil {
					return
				}

				slog.Error("chat listing failed")

				writeJSON(w, http.StatusInternalServerError, errorResponse{
					Error: "internal server error",
				})
			}
			return
		}

		response := make([]chatResponse, 0, len(chats))
		for _, item := range chats {
			response = append(response, chatResponse{
				ID:        item.ID,
				Kind:      item.Kind,
				Title:     item.Title,
				CreatedBy: item.CreatedBy,
				CreatedAt: item.CreatedAt,
			})
		}

		writeJSON(w, http.StatusOK, response)
	}
}

func parseChatListInt(
	query url.Values,
	name string,
	defaultValue int,
) (int, error) {
	values, exists := query[name]
	if !exists {
		return defaultValue, nil
	}

	if len(values) != 1 {
		return 0, fmt.Errorf("%s must be specified once", name)
	}

	value, err := strconv.Atoi(values[0])
	if err != nil {
		return 0, fmt.Errorf("%s must be an integer", name)
	}

	return value, nil
}
