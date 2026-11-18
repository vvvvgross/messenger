package auth

import (
	"context"
	"fmt"

	"github.com/vvvvgross/messenger/internal/storage/postgres"
)

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
