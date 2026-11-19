package auth

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

const (
	accessTokenTTL = 15 * time.Minute
	tokenIssuer    = "messenger"
	tokenAudience  = "messenger-api"
)

type TokenManager struct {
	secret []byte
}

func NewTokenManager(secret string) (*TokenManager, error) {
	if len(secret) < 32 {
		return nil, errors.New("JWT_SECRET must contain at least 32 bytes")
	}

	if strings.TrimSpace(secret) == "" {
		return nil, errors.New("JWT_SECRET must not contain only whitespace")
	}

	return &TokenManager{
		secret: []byte(secret),
	}, nil
}

func (m *TokenManager) Generate(userID string) (string, error) {
	if strings.TrimSpace(userID) == "" {
		return "", errors.New("cannot generate token without user ID")
	}

	now := time.Now()

	claims := jwt.RegisteredClaims{
		Subject:   userID,
		Issuer:    tokenIssuer,
		Audience:  jwt.ClaimStrings{tokenAudience},
		IssuedAt:  jwt.NewNumericDate(now),
		ExpiresAt: jwt.NewNumericDate(now.Add(accessTokenTTL)),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	signed, err := token.SignedString(m.secret)
	if err != nil {
		return "", fmt.Errorf("sign access token: %w", err)
	}

	return signed, nil
}
