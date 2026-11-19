package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"
	"unicode"

	"github.com/vvvvgross/messenger/internal/auth"
	"github.com/vvvvgross/messenger/internal/httpapi"
	"github.com/vvvvgross/messenger/internal/storage/postgres"
)

func main() {
	if err := run(); err != nil {
		slog.Error("server stopped", "error", err)
		os.Exit(1)
	}
}

func run() error {
	addr, err := parseHTTPAddr(os.Getenv("HTTP_ADDR"))
	if err != nil {
		return err
	}

	ctx, stop := signal.NotifyContext(
		context.Background(),
		os.Interrupt,
		syscall.SIGTERM,
	)

	defer stop()

	pool, err := postgres.NewPool(ctx, os.Getenv("DATABASE_URL"))
	if err != nil {
		return fmt.Errorf("initialize database: %w", err)
	}

	defer pool.Close()

	slog.Info("PostgreSQL connection established")

	users := postgres.NewUserRepository(pool)
	authService := auth.NewService(users)

	server := &http.Server{
		Addr:              addr,
		Handler:           httpapi.NewHandler(authService),
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       10 * time.Second,
		WriteTimeout:      10 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	result := make(chan error, 1)
	go func() {
		result <- server.ListenAndServe()
	}()

	slog.Info("starting HTTP server", "addr", addr)

	select {
	case err := <-result:
		return fmt.Errorf("listen: %w", err)
	case <-ctx.Done():
		stop()
	}

	slog.Info("shutting down HTTP server")

	shutdownCtx, cancel := context.WithTimeout(
		context.Background(),
		10*time.Second,
	)
	defer cancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		_ = server.Close()
		return fmt.Errorf("shutdown: %w", err)
	}

	if err := <-result; err != nil && !errors.Is(err, http.ErrServerClosed) {
		return fmt.Errorf("serve: %w", err)
	}

	return nil
}

func parseHTTPAddr(value string) (string, error) {
	if value == "" {
		return ":8080", nil
	}

	if strings.IndexFunc(value, unicode.IsSpace) >= 0 {
		return "", fmt.Errorf("HTTP_ADDR %q must not contain whitespace", value)
	}

	_, port, err := net.SplitHostPort(value)
	if err != nil {
		return "", fmt.Errorf(
			"HTTP_ADDR %q must use host:port format: %w",
			value,
			err,
		)
	}

	if port == "" || strings.IndexFunc(port, func(r rune) bool {
		return r < '0' || r > '9'
	}) >= 0 {
		return "", fmt.Errorf(
			"HTTP_ADDR %q must contain a numeric port",
			value,
		)
	}

	number, err := strconv.ParseUint(port, 10, 16)
	if err != nil || number == 0 {
		return "", fmt.Errorf(
			"HTTP_ADDR %q port must be between 1 and 65535",
			value,
		)
	}

	return value, nil
}
