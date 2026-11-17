package postgres

import (
	"context"
	"errors"
	"strings"
	"testing"
)

func TestNewPoolRejectsEmptyURL(t *testing.T) {
	for _, value := range []string{"", " ", "\t\n"} {
		t.Run(value, func(t *testing.T) {
			pool, err := NewPool(context.Background(), value)

			if pool != nil {
				pool.Close()
				t.Fatal("expected nil pool")
			}
			if err == nil {
				t.Fatal("expected error for empty DATABASE_URL")
			}
		})
	}
}

func TestNewPoolRejectsInvalidURL(t *testing.T) {
	const databaseURL = "postgres://user:secret-password@localhost:5432/db%ZZ"

	pool, err := NewPool(context.Background(), databaseURL)

	if pool != nil {
		pool.Close()
		t.Fatal("expected nil pool")
	}
	if err == nil {
		t.Fatal("expected error for invalid DATABASE_URL")
	}

	if strings.Contains(err.Error(), "secret-password") {
		t.Fatal("error exposes the database password")
	}
	if strings.Contains(err.Error(), databaseURL) {
		t.Fatal("error exposes the connection string")
	}
}

func TestNewPoolWithCancelledContext(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	pool, err := NewPool(
		ctx,
		"postgres://user:password@127.0.0.1:5432/messenger?sslmode=disable",
	)

	if pool != nil {
		pool.Close()
		t.Fatal("expected nil pool")
	}
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("expected context.Canceled, got %v", err)
	}
}
