package auth

import (
	"errors"
	"regexp"
	"strings"
	"unicode/utf8"
)

const minPasswordCharacters = 8

var (
	ErrInvalidUsername = errors.New(
		"username must contain 3–32 Latin letters, digits or underscores",
	)
	ErrInvalidPasswordEncoding = errors.New("password must be valid UTF-8")
	ErrPasswordTooShort        = errors.New(
		"password must contain at least 8 characters",
	)
)

var usernamePattern = regexp.MustCompile(`^[a-zA-Z0-9_]{3,32}$`)

func NormalizeUsername(username string) (string, error) {
	if !usernamePattern.MatchString(username) {
		return "", ErrInvalidUsername
	}

	return strings.ToLower(username), nil
}

func ValidatePassword(password string) error {
	if len(password) > maxPasswordBytes {
		return ErrPasswordTooLong
	}

	if !utf8.ValidString(password) {
		return ErrInvalidPasswordEncoding
	}

	if utf8.RuneCountInString(password) < minPasswordCharacters {
		return ErrPasswordTooShort
	}

	return nil
}
