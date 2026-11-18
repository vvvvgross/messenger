package postgres

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

var ErrUsernameTaken = errors.New("username already taken")
var ErrUserNotFound = errors.New("user not found")

type User struct {
	ID        string
	Username  string
	CreatedAt time.Time
}

type UserCredentials struct {
	User         User
	PasswordHash string `json:"-"`
}

type UserRepository struct {
	pool *pgxpool.Pool
}

func NewUserRepository(pool *pgxpool.Pool) *UserRepository {
	return &UserRepository{pool: pool}
}

func (r *UserRepository) Create(
	ctx context.Context,
	username string,
	passwordHash string,
) (User, error) {
	const query = `
		INSERT INTO users (username, password_hash)
		VALUES ($1, $2)
		RETURNING id::text, username, created_at
	`

	var user User

	err := r.pool.QueryRow(ctx, query, username, passwordHash).Scan(
		&user.ID,
		&user.Username,
		&user.CreatedAt,
	)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) &&
			pgErr.Code == "23505" &&
			pgErr.ConstraintName == "users_username_unique" {
			return User{}, ErrUsernameTaken
		}

		return User{}, fmt.Errorf("create user: %w", err)
	}

	return user, nil
}

func (r *UserRepository) FindByUsername(
	ctx context.Context,
	username string,
) (UserCredentials, error) {
	const query = `
		SELECT id::text, username, created_at, password_hash
		FROM users
		WHERE username = $1
	`

	var credentials UserCredentials

	err := r.pool.QueryRow(ctx, query, username).Scan(
		&credentials.User.ID,
		&credentials.User.Username,
		&credentials.User.CreatedAt,
		&credentials.PasswordHash,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return UserCredentials{}, ErrUserNotFound
		}

		return UserCredentials{}, fmt.Errorf("find user by username: %w", err)
	}

	return credentials, nil
}
