package postgres

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Chat struct {
	ID        string
	Kind      string
	Title     *string
	CreatedBy string
	CreatedAt time.Time
}

type ChatRepository struct {
	pool *pgxpool.Pool
}

func NewChatRepository(pool *pgxpool.Pool) *ChatRepository {
	return &ChatRepository{pool: pool}
}

func (r *ChatRepository) CreateGroup(
	ctx context.Context,
	creatorID string,
	title string,
) (Chat, error) {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return Chat{}, fmt.Errorf("begin group creation: %w", err)
	}

	defer func() {
		rollbackCtx, cancel := context.WithTimeout(
			context.Background(),
			5*time.Second,
		)
		defer cancel()

		_ = tx.Rollback(rollbackCtx)
	}()

	const createChatQuery = `
		INSERT INTO chats (kind, title, created_by)
		VALUES ('group', $1, $2)
		RETURNING id::text, kind, title, created_by::text, created_at
	`

	var chat Chat

	err = tx.QueryRow(ctx, createChatQuery, title, creatorID).Scan(
		&chat.ID,
		&chat.Kind,
		&chat.Title,
		&chat.CreatedBy,
		&chat.CreatedAt,
	)
	if err != nil {
		return Chat{}, fmt.Errorf("insert group chat: %w", err)
	}

	const addAdminQuery = `
		INSERT INTO chat_members (chat_id, user_id, role)
		VALUES ($1, $2, 'admin')
	`

	_, err = tx.Exec(ctx, addAdminQuery, chat.ID, creatorID)
	if err != nil {
		return Chat{}, fmt.Errorf("add group administrator: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return Chat{}, fmt.Errorf("commit group creation: %w", err)
	}

	return chat, nil
}
