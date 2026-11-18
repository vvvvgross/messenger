package httpapi

import (
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"mime"
	"net/http"
	"time"

	"github.com/vvvvgross/messenger/internal/auth"
	"github.com/vvvvgross/messenger/internal/storage/postgres"
)

const maxRegisterBodyBytes = 4 * 1024

type registerRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type registerResponse struct {
	ID        string    `json:"id"`
	Username  string    `json:"username"`
	CreatedAt time.Time `json:"created_at"`
}

type errorResponse struct {
	Error string `json:"error"`
}

func newRegisterHandler(service *auth.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		contentType, _, err := mime.ParseMediaType(
			r.Header.Get("Content-Type"),
		)
		if err != nil || contentType != "application/json" {
			writeJSON(w, http.StatusUnsupportedMediaType, errorResponse{
				Error: "Content-Type must be application/json",
			})
			return
		}

		r.Body = http.MaxBytesReader(w, r.Body, maxRegisterBodyBytes)
		defer r.Body.Close()

		decoder := json.NewDecoder(r.Body)
		decoder.DisallowUnknownFields()

		var input registerRequest
		if err := decoder.Decode(&input); err != nil {
			writeRegisterDecodeError(w, err)
			return
		}

		if err := decoder.Decode(new(any)); err != io.EOF {
			writeRegisterDecodeError(w, err)
			return
		}

		user, err := service.Register(
			r.Context(),
			input.Username,
			input.Password,
		)
		if err != nil {
			switch {
			case errors.Is(err, auth.ErrInvalidUsername),
				errors.Is(err, auth.ErrInvalidPasswordEncoding),
				errors.Is(err, auth.ErrPasswordTooShort),
				errors.Is(err, auth.ErrPasswordTooLong):
				writeJSON(w, http.StatusBadRequest, errorResponse{
					Error: err.Error(),
				})

			case errors.Is(err, postgres.ErrUsernameTaken):
				writeJSON(w, http.StatusConflict, errorResponse{
					Error: "username already taken",
				})

			default:
				if r.Context().Err() != nil {
					return
				}

				slog.Error("registration failed")

				writeJSON(w, http.StatusInternalServerError, errorResponse{
					Error: "internal server error",
				})
			}

			return
		}

		writeJSON(w, http.StatusCreated, registerResponse{
			ID:        user.ID,
			Username:  user.Username,
			CreatedAt: user.CreatedAt,
		})
	}
}

func writeRegisterDecodeError(w http.ResponseWriter, err error) {
	var sizeError *http.MaxBytesError

	if errors.As(err, &sizeError) {
		writeJSON(w, http.StatusRequestEntityTooLarge, errorResponse{
			Error: "request body is too large",
		})
		return
	}

	writeJSON(w, http.StatusBadRequest, errorResponse{
		Error: "body must contain one JSON object with username and password",
	})
}

func writeJSON(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	if err := json.NewEncoder(w).Encode(value); err != nil {
		slog.Error("failed to write JSON response", "error", err)
	}
}
