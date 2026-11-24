package chat

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"unicode/utf8"

	"github.com/vvvvgross/messenger/internal/storage/postgres"
)

const maxGroupTitleCharacters = 100

var ErrInvalidGroupTitle = errors.New(
	"group title must be valid text containing 1–100 characters",
)

type Service struct {
	chats *postgres.ChatRepository
}

func NewService(chats *postgres.ChatRepository) *Service {
	return &Service{chats: chats}
}

func (s *Service) CreateGroup(
	ctx context.Context,
	creatorID string,
	title string,
) (postgres.Chat, error) {
	if !utf8.ValidString(title) || strings.ContainsRune(title, '\x00') {
		return postgres.Chat{}, ErrInvalidGroupTitle
	}

	title = strings.TrimSpace(title)

	length := utf8.RuneCountInString(title)
	if length == 0 || length > maxGroupTitleCharacters {
		return postgres.Chat{}, ErrInvalidGroupTitle
	}

	group, err := s.chats.CreateGroup(ctx, creatorID, title)
	if err != nil {
		return postgres.Chat{}, fmt.Errorf("create group: %w", err)
	}

	return group, nil
}
