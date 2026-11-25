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

func (r *ChatRepository) ListByUser(
	ctx context.Context,
	userID string,
	limit int,
	offset int,
) ([]Chat, error) {
	const query = `
		SELECT
			c.id::text,
			c.kind,
			c.title,
			c.created_by::text,
			c.created_at
		FROM chats AS c
		JOIN chat_members AS cm ON cm.chat_id = c.id
		WHERE cm.user_id = $1
		ORDER BY c.created_at DESC, c.id DESC
		LIMIT $2 OFFSET $3
	`

	rows, err := r.pool.Query(ctx, query, userID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("query user chats: %w", err)
	}
	defer rows.Close()

	chats := make([]Chat, 0)

	for rows.Next() {
		var chat Chat

		if err := rows.Scan(
			&chat.ID,
			&chat.Kind,
			&chat.Title,
			&chat.CreatedBy,
			&chat.CreatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan user chat: %w", err)
		}

		chats = append(chats, chat)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("read user chats: %w", err)
	}

	return chats, nil
}
