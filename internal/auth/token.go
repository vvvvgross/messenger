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

var ErrInvalidToken = errors.New("invalid or expired access token")

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

func (m *TokenManager) Parse(tokenString string) (string, error) {
	if strings.TrimSpace(tokenString) == "" {
		return "", ErrInvalidToken
	}

	var claims jwt.RegisteredClaims

	token, err := jwt.ParseWithClaims(
		tokenString,
		&claims,
		func(token *jwt.Token) (any, error) {
			if token.Method != jwt.SigningMethodHS256 {
				return nil, ErrInvalidToken
			}

			return m.secret, nil
		},
		jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Alg()}),
		jwt.WithExpirationRequired(),
		jwt.WithIssuer(tokenIssuer),
		jwt.WithAudience(tokenAudience),
		jwt.WithIssuedAt(),
	)
	if err != nil || !token.Valid {
		return "", ErrInvalidToken
	}

	if strings.TrimSpace(claims.Subject) == "" {
		return "", ErrInvalidToken
	}

	if claims.IssuedAt == nil {
		return "", ErrInvalidToken
	}

	if claims.ExpiresAt == nil ||
		!claims.ExpiresAt.Time.After(claims.IssuedAt.Time) {
		return "", ErrInvalidToken
	}

	return claims.Subject, nil
}
