package auth

import (
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// TokenType distinguishes an access token from a refresh token so a refresh
// token can never be replayed as a bearer credential.
type TokenType string

const (
	TokenTypeAccess  TokenType = "access"
	TokenTypeRefresh TokenType = "refresh"
)

var (
	ErrInvalidToken = errors.New("invalid token")
	ErrExpiredToken = errors.New("token expired")
)

type TokenOptions struct {
	Issuer    string
	AccessTTL time.Duration
}

type CustomClaims struct {
	UserID string    `json:"user_id"`
	Role   string    `json:"role"`
	Type   TokenType `json:"typ"`
	jwt.RegisteredClaims
}

// GenerateAccessToken issues a short-lived bearer credential.
func GenerateAccessToken(userID, role, secret string, opts TokenOptions) (string, error) {
	return generate(userID, role, TokenTypeAccess, secret, opts)
}

// GenerateRefreshToken issues a long-lived credential usable only against the
// refresh endpoint.
func GenerateRefreshToken(userID, role, secret string, opts TokenOptions) (string, error) {
	return generate(userID, role, TokenTypeRefresh, secret, opts)
}

func generate(userID, role string, tokenType TokenType, secret string, opts TokenOptions) (string, error) {
	if userID == "" {
		return "", errors.New("user id is required")
	}
	if secret == "" {
		return "", errors.New("signing secret is required")
	}

	ttl := opts.AccessTTL
	if tokenType == TokenTypeRefresh {
		ttl = 30 * 24 * time.Hour
	}
	switch {
	case ttl < 0:
		return "", fmt.Errorf("token TTL must be positive, got %s", ttl)
	case ttl == 0:
		ttl = 24 * time.Hour
	}

	now := time.Now()
	claims := CustomClaims{
		UserID: userID,
		Role:   role,
		Type:   tokenType,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    opts.Issuer,
			Subject:   userID,
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(ttl)),
			ID:        fmt.Sprintf("%s-%d", tokenType, now.UnixNano()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := token.SignedString([]byte(secret))
	if err != nil {
		return "", err
	}
	return signed, nil
}

// ValidateToken parses a token and verifies its signature, expiry and issuer.
//
// The key function pins the expected signing algorithm. Without that check a
// crafted token could select a different algorithm (including "none") and
// bypass signature verification entirely.
func ValidateToken(tokenString, secret, issuer string) (*CustomClaims, error) {
	if tokenString == "" || secret == "" {
		return nil, ErrInvalidToken
	}

	claims := &CustomClaims{}
	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method %q", token.Header["alg"])
		}
		return []byte(secret), nil
	},
		jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Alg()}),
		jwt.WithExpirationRequired(),
	)
	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			return nil, ErrExpiredToken
		}
		return nil, fmt.Errorf("%w: %s", ErrInvalidToken, err)
	}
	if !token.Valid {
		return nil, ErrInvalidToken
	}
	if issuer != "" && claims.Issuer != issuer {
		return nil, fmt.Errorf("%w: unexpected issuer %q", ErrInvalidToken, claims.Issuer)
	}
	if claims.UserID == "" {
		return nil, fmt.Errorf("%w: missing user id", ErrInvalidToken)
	}
	if claims.Type == "" {
		return nil, fmt.Errorf("%w: missing token type", ErrInvalidToken)
	}
	return claims, nil
}

// ValidateAccessToken additionally rejects refresh tokens presented as bearer
// credentials.
func ValidateAccessToken(tokenString, secret, issuer string) (*CustomClaims, error) {
	claims, err := ValidateToken(tokenString, secret, issuer)
	if err != nil {
		return nil, err
	}
	if claims.Type != TokenTypeAccess {
		return nil, fmt.Errorf("%w: expected an access token, got %q", ErrInvalidToken, claims.Type)
	}
	return claims, nil
}

// ValidateRefreshToken rejects access tokens presented to the refresh endpoint.
func ValidateRefreshToken(tokenString, secret, issuer string) (*CustomClaims, error) {
	claims, err := ValidateToken(tokenString, secret, issuer)
	if err != nil {
		return nil, err
	}
	if claims.Type != TokenTypeRefresh {
		return nil, fmt.Errorf("%w: expected a refresh token, got %q", ErrInvalidToken, claims.Type)
	}
	return claims, nil
}
