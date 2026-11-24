package httpapi

import (
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"mime"
	"net/http"
	"time"

	"github.com/vvvvgross/messenger/internal/chat"
)

const maxCreateGroupBodyBytes = 4 * 1024

type createGroupRequest struct {
	Title string `json:"title"`
}

type chatResponse struct {
	ID        string    `json:"id"`
	Kind      string    `json:"kind"`
	Title     *string   `json:"title"`
	CreatedBy string    `json:"created_by"`
	CreatedAt time.Time `json:"created_at"`
}

func newCreateGroupHandler(service *chat.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Cache-Control", "no-store")

		creatorID, ok := userIDFromContext(r.Context())
		if !ok {
			writeUnauthorized(w)
			return
		}

		contentType, _, err := mime.ParseMediaType(
			r.Header.Get("Content-Type"),
		)
		if err != nil || contentType != "application/json" {
			writeJSON(w, http.StatusUnsupportedMediaType, errorResponse{
				Error: "Content-Type must be application/json",
			})
			return
		}

		r.Body = http.MaxBytesReader(w, r.Body, maxCreateGroupBodyBytes)
		defer r.Body.Close()

		decoder := json.NewDecoder(r.Body)
		decoder.DisallowUnknownFields()

		var input createGroupRequest
		if err := decoder.Decode(&input); err != nil {
			writeCreateGroupDecodeError(w, err)
			return
		}

		if err := decoder.Decode(new(any)); err != io.EOF {
			writeCreateGroupDecodeError(w, err)
			return
		}

		group, err := service.CreateGroup(
			r.Context(),
			creatorID,
			input.Title,
		)
		if err != nil {
			if errors.Is(err, chat.ErrInvalidGroupTitle) {
				writeJSON(w, http.StatusBadRequest, errorResponse{
					Error: chat.ErrInvalidGroupTitle.Error(),
				})
				return
			}

			if r.Context().Err() != nil {
				return
			}

			slog.Error("group creation failed")

			writeJSON(w, http.StatusInternalServerError, errorResponse{
				Error: "internal server error",
			})
			return
		}

		writeJSON(w, http.StatusCreated, chatResponse{
			ID:        group.ID,
			Kind:      group.Kind,
			Title:     group.Title,
			CreatedBy: group.CreatedBy,
			CreatedAt: group.CreatedAt,
		})
	}
}

func writeCreateGroupDecodeError(w http.ResponseWriter, err error) {
	var sizeError *http.MaxBytesError

	if errors.As(err, &sizeError) {
		writeJSON(w, http.StatusRequestEntityTooLarge, errorResponse{
			Error: "request body is too large",
		})
		return
	}

	writeJSON(w, http.StatusBadRequest, errorResponse{
		Error: "body must contain one JSON object with title",
	})
}
