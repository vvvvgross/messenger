package httpapi

import (
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"mime"
	"net/http"

	"github.com/vvvvgross/messenger/internal/auth"
)

const maxLoginBodyBytes = 4 * 1024

type loginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type loginResponse struct {
	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type"`
}

func newLoginHandler(
	service *auth.Service,
	tokens *auth.TokenManager,
) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Cache-Control", "no-store")

		contentType, _, err := mime.ParseMediaType(
			r.Header.Get("Content-Type"),
		)
		if err != nil || contentType != "application/json" {
			writeJSON(w, http.StatusUnsupportedMediaType, errorResponse{
				Error: "Content-Type must be application/json",
			})
			return
		}

		r.Body = http.MaxBytesReader(w, r.Body, maxLoginBodyBytes)
		defer r.Body.Close()

		decoder := json.NewDecoder(r.Body)
		decoder.DisallowUnknownFields()

		var input loginRequest
		if err := decoder.Decode(&input); err != nil {
			writeLoginDecodeError(w, err)
			return
		}

		if err := decoder.Decode(new(any)); err != io.EOF {
			writeLoginDecodeError(w, err)
			return
		}

		user, err := service.Authenticate(
			r.Context(),
			input.Username,
			input.Password,
		)
		if err != nil {
			if errors.Is(err, auth.ErrInvalidCredentials) {
				writeJSON(w, http.StatusUnauthorized, errorResponse{
					Error: "invalid username or password",
				})
				return
			}

			if r.Context().Err() != nil {
				return
			}

			slog.Error("authentication failed")

			writeJSON(w, http.StatusInternalServerError, errorResponse{
				Error: "internal server error",
			})
			return
		}

		if r.Context().Err() != nil {
			return
		}

		accessToken, err := tokens.Generate(user.ID)
		if err != nil {
			slog.Error("access token generation failed")

			writeJSON(w, http.StatusInternalServerError, errorResponse{
				Error: "internal server error",
			})
			return
		}

		writeJSON(w, http.StatusOK, loginResponse{
			AccessToken: accessToken,
			TokenType:   "Bearer",
		})
	}
}

func writeLoginDecodeError(w http.ResponseWriter, err error) {
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
