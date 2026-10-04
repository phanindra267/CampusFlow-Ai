package auth

import (
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// TransientClaims carries short-lived, signed state that must survive a
// redirect round-trip through a third party — most importantly the OAuth
// "state" parameter. It is signed rather than stored server-side so no session
// store is required, while still being unforgeable and single-purpose.
type TransientClaims struct {
	Nonce       string `json:"nonce"`
	RedirectURI string `json:"redirect_uri,omitempty"`
	jwt.RegisteredClaims
}

// SignTransient issues a signed, short-lived state token.
func SignTransient(nonce, redirectURI, secret, issuer string, ttl time.Duration) (string, error) {
	if secret == "" {
		return "", fmt.Errorf("signing secret is required")
	}
	switch {
	case ttl < 0:
		return "", fmt.Errorf("state TTL must be positive, got %s", ttl)
	case ttl == 0:
		ttl = 10 * time.Minute
	}

	now := time.Now()
	claims := TransientClaims{
		Nonce:       nonce,
		RedirectURI: redirectURI,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    issuer,
			Subject:   "oauth-state",
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(ttl)),
		},
	}

	signed, err := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString([]byte(secret))
	if err != nil {
		return "", err
	}
	return signed, nil
}

// VerifyTransient validates a state token produced by SignTransient.
func VerifyTransient(tokenString, secret, issuer string) (*TransientClaims, error) {
	if tokenString == "" || secret == "" {
		return nil, fmt.Errorf("%w: state token missing", ErrInvalidToken)
	}

	claims := &TransientClaims{}
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
		return nil, fmt.Errorf("%w: %s", ErrInvalidToken, err)
	}
	if !token.Valid || claims.Nonce == "" {
		return nil, fmt.Errorf("%w: unusable state token", ErrInvalidToken)
	}
	if issuer != "" && claims.Issuer != issuer {
		return nil, fmt.Errorf("%w: unexpected issuer", ErrInvalidToken)
	}
	return claims, nil
}
