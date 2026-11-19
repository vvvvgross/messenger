package auth

import (
	"context"
	"errors"
	"fmt"
	"unicode/utf8"

	"github.com/vvvvgross/messenger/internal/storage/postgres"
)

var ErrInvalidCredentials = errors.New("invalid username or password")

type Service struct {
	users *postgres.UserRepository
}

func NewService(users *postgres.UserRepository) *Service {
	return &Service{users: users}
}

func (s *Service) Register(
	ctx context.Context,
	username string,
	password string,
) (postgres.User, error) {
	username, err := NormalizeUsername(username)
	if err != nil {
		return postgres.User{}, err
	}

	if err := ValidatePassword(password); err != nil {
		return postgres.User{}, err
	}

	if err := ctx.Err(); err != nil {
		return postgres.User{}, err
	}

	passwordHash, err := HashPassword(password)
	if err != nil {
		return postgres.User{}, fmt.Errorf("register user: %w", err)
	}

	user, err := s.users.Create(ctx, username, passwordHash)
	if err != nil {
		return postgres.User{}, fmt.Errorf("register user: %w", err)
	}

	return user, nil
}

func (s *Service) Authenticate(
	ctx context.Context,
	username string,
	password string,
) (postgres.User, error) {
	username, err := NormalizeUsername(username)
	if err != nil {
		return postgres.User{}, ErrInvalidCredentials
	}

	if password == "" ||
		len(password) > maxPasswordBytes ||
		!utf8.ValidString(password) {
		return postgres.User{}, ErrInvalidCredentials
	}

	credentials, err := s.users.FindByUsername(ctx, username)
	if err != nil {
		if errors.Is(err, postgres.ErrUserNotFound) {
			return postgres.User{}, ErrInvalidCredentials
		}

		return postgres.User{}, fmt.Errorf("authenticate user: %w", err)
	}

	if err := ctx.Err(); err != nil {
		return postgres.User{}, err
	}

	matched, err := CheckPassword(credentials.PasswordHash, password)
	if err != nil {
		return postgres.User{}, fmt.Errorf("authenticate user: %w", err)
	}

	if !matched {
		return postgres.User{}, ErrInvalidCredentials
	}

	return credentials.User, nil
}
