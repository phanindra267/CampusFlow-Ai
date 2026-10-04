package auth

import (
	"encoding/base64"
	"encoding/json"
	"strings"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

const (
	jwtSecret = "unit-test-secret-that-is-long-enough-1234"
	jwtIssuer = "campuscare-test"
)

func testOpts() TokenOptions {
	return TokenOptions{Issuer: jwtIssuer, AccessTTL: time.Hour}
}

func TestAccessAndRefreshTokensAreDistinctKinds(t *testing.T) {
	access, err := GenerateAccessToken("user-1", "MEMBER", jwtSecret, testOpts())
	if err != nil {
		t.Fatalf("generate access token: %v", err)
	}
	refresh, err := GenerateRefreshToken("user-1", "MEMBER", jwtSecret, testOpts())
	if err != nil {
		t.Fatalf("generate refresh token: %v", err)
	}
	if access == refresh {
		t.Fatal("access and refresh tokens are byte-identical")
	}

	if _, err := ValidateAccessToken(access, jwtSecret, jwtIssuer); err != nil {
		t.Errorf("access token failed access validation: %v", err)
	}
	if _, err := ValidateRefreshToken(access, jwtSecret, jwtIssuer); err == nil {
		t.Error("an access token must not validate as a refresh token")
	}

	if _, err := ValidateRefreshToken(refresh, jwtSecret, jwtIssuer); err != nil {
		t.Errorf("refresh token failed refresh validation: %v", err)
	}
	if _, err := ValidateAccessToken(refresh, jwtSecret, jwtIssuer); err == nil {
		t.Error("a refresh token must not validate as an access token")
	}
}

// A negative TTL is a configuration error and must be reported rather than
// silently coerced to the default, otherwise a misconfigured TTL would hand out
// long-lived tokens.
func TestGenerateRejectsNegativeTTL(t *testing.T) {
	if _, err := GenerateAccessToken("user-1", "MEMBER", jwtSecret, TokenOptions{
		Issuer: jwtIssuer, AccessTTL: -time.Minute,
	}); err == nil {
		t.Error("a negative access TTL was silently accepted")
	}
	if _, err := SignTransient("nonce", "", jwtSecret, jwtIssuer, -time.Minute); err == nil {
		t.Error("a negative state TTL was silently accepted")
	}
}

func TestGenerateAppliesDefaultWhenTTLUnset(t *testing.T) {
	token, err := GenerateAccessToken("user-1", "MEMBER", jwtSecret, TokenOptions{Issuer: jwtIssuer})
	if err != nil {
		t.Fatalf("generate token: %v", err)
	}
	claims, err := ValidateAccessToken(token, jwtSecret, jwtIssuer)
	if err != nil {
		t.Fatalf("validate token: %v", err)
	}
	if lifetime := claims.ExpiresAt.Sub(claims.IssuedAt.Time); lifetime < 23*time.Hour || lifetime > 25*time.Hour {
		t.Errorf("expected the 24h default lifetime, got %s", lifetime)
	}
}

// The original keyfunc ignored the token's declared algorithm, which allows a
// crafted "alg: none" token to bypass signature verification entirely.
func TestRejectsUnsignedAlgNoneToken(t *testing.T) {
	claims := CustomClaims{
		UserID: "attacker",
		Role:   "SUPER_ADMIN",
		Type:   TokenTypeAccess,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    jwtIssuer,
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour)),
		},
	}

	raw, err := jwt.NewWithClaims(jwt.SigningMethodNone, claims).
		SignedString(jwt.UnsafeAllowNoneSignatureType)
	if err != nil {
		t.Fatalf("build unsigned token: %v", err)
	}

	if _, err := ValidateAccessToken(raw, jwtSecret, jwtIssuer); err == nil {
		t.Fatal("an unsigned alg=none token was accepted")
	}
}

func TestRejectsForeignSigningKey(t *testing.T) {
	forged, err := GenerateAccessToken("attacker", "SUPER_ADMIN", "a-completely-different-secret-value", testOpts())
	if err != nil {
		t.Fatalf("generate forged token: %v", err)
	}

	if _, err := ValidateAccessToken(forged, jwtSecret, jwtIssuer); err == nil {
		t.Fatal("a token signed with a different secret was accepted")
	}
}

func TestRejectsWrongIssuer(t *testing.T) {
	token, err := GenerateAccessToken("user-1", "MEMBER", jwtSecret, TokenOptions{Issuer: "somebody-else", AccessTTL: time.Hour})
	if err != nil {
		t.Fatalf("generate token: %v", err)
	}

	if _, err := ValidateAccessToken(token, jwtSecret, jwtIssuer); err == nil {
		t.Fatal("a token from an unexpected issuer was accepted")
	}
}

// An expired token must be rejected even when it carries a valid signature.
// The token is forged by hand because the generator refuses negative TTLs.
func TestRejectsExpiredToken(t *testing.T) {
	past := time.Now().Add(-2 * time.Hour)
	forged := jwt.NewWithClaims(jwt.SigningMethodHS256, CustomClaims{
		UserID: "user-1",
		Role:   "MEMBER",
		Type:   TokenTypeAccess,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    jwtIssuer,
			IssuedAt:  jwt.NewNumericDate(past),
			NotBefore: jwt.NewNumericDate(past),
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(-time.Hour)),
		},
	})

	raw, err := forged.SignedString([]byte(jwtSecret))
	if err != nil {
		t.Fatalf("sign forged token: %v", err)
	}

	if _, err := ValidateAccessToken(raw, jwtSecret, jwtIssuer); err == nil {
		t.Fatal("an expired token was accepted")
	}
}

// A token with no expiry at all must also be rejected, otherwise a leaked
// token would be valid forever.
func TestRejectsTokenWithoutExpiry(t *testing.T) {
	forged := jwt.NewWithClaims(jwt.SigningMethodHS256, CustomClaims{
		UserID: "user-1",
		Role:   "SUPER_ADMIN",
		Type:   TokenTypeAccess,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer: jwtIssuer,
		},
	})

	raw, err := forged.SignedString([]byte(jwtSecret))
	if err != nil {
		t.Fatalf("sign forged token: %v", err)
	}

	if _, err := ValidateAccessToken(raw, jwtSecret, jwtIssuer); err == nil {
		t.Fatal("a token with no expiry claim was accepted")
	}
}

func TestRejectsMalformedTokens(t *testing.T) {
	cases := map[string]string{
		"empty":           "",
		"not a jwt":       "definitely-not-a-jwt",
		"two segments":    "aaa.bbb",
		"four segments":   "aaa.bbb.ccc.ddd",
		"garbage payload": "aaa." + base64.RawURLEncoding.EncodeToString([]byte("!!!")) + ".ccc",
		"empty signature": "aaa.bbb.",
	}
	for name, token := range cases {
		if _, err := ValidateToken(token, jwtSecret, jwtIssuer); err == nil {
			t.Errorf("%s: malformed token was accepted", name)
		}
	}
}

func TestValidateRequiresSecret(t *testing.T) {
	token, err := GenerateAccessToken("user-1", "MEMBER", jwtSecret, testOpts())
	if err != nil {
		t.Fatalf("generate token: %v", err)
	}
	if _, err := ValidateAccessToken(token, "", jwtIssuer); err == nil {
		t.Fatal("validation succeeded without a secret")
	}
}

func TestGenerateRejectsEmptyInputs(t *testing.T) {
	if _, err := GenerateAccessToken("", "MEMBER", jwtSecret, testOpts()); err == nil {
		t.Error("empty user id was accepted")
	}
	if _, err := GenerateAccessToken("user-1", "MEMBER", "", testOpts()); err == nil {
		t.Error("empty secret was accepted")
	}
}

func TestTokenExpiryHonoursConfiguredTTL(t *testing.T) {
	token, err := GenerateAccessToken("user-1", "MEMBER", jwtSecret, TokenOptions{
		Issuer: jwtIssuer, AccessTTL: 2 * time.Hour,
	})
	if err != nil {
		t.Fatalf("generate token: %v", err)
	}

	claims, err := ValidateAccessToken(token, jwtSecret, jwtIssuer)
	if err != nil {
		t.Fatalf("validate token: %v", err)
	}

	lifetime := claims.ExpiresAt.Sub(claims.IssuedAt.Time)
	if lifetime < 119*time.Minute || lifetime > 121*time.Minute {
		t.Errorf("expected roughly a 2h lifetime, got %s", lifetime)
	}
}

func TestTransientStateRoundTrip(t *testing.T) {
	signed, err := SignTransient("nonce-123", "/dashboard", jwtSecret, jwtIssuer, time.Minute)
	if err != nil {
		t.Fatalf("sign transient: %v", err)
	}

	claims, err := VerifyTransient(signed, jwtSecret, jwtIssuer)
	if err != nil {
		t.Fatalf("verify transient: %v", err)
	}
	if claims.Nonce != "nonce-123" {
		t.Errorf("unexpected nonce %q", claims.Nonce)
	}
	if claims.RedirectURI != "/dashboard" {
		t.Errorf("unexpected redirect %q", claims.RedirectURI)
	}
}

func TestTransientStateRejectsTampering(t *testing.T) {
	signed, err := SignTransient("nonce-123", "/dashboard", jwtSecret, jwtIssuer, time.Minute)
	if err != nil {
		t.Fatalf("sign transient: %v", err)
	}

	parts := strings.Split(signed, ".")
	if len(parts) != 3 {
		t.Fatalf("unexpected token shape: %d segments", len(parts))
	}

	// Flip a character in the payload and re-encode it.
	payload, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		t.Fatalf("decode payload: %v", err)
	}
	var claims map[string]any
	if err := json.Unmarshal(payload, &claims); err != nil {
		t.Fatalf("parse payload: %v", err)
	}
	claims["nonce"] = "nonce-attacker"
	forgedPayload, _ := json.Marshal(claims)
	tampered := parts[0] + "." + base64.RawURLEncoding.EncodeToString(forgedPayload) + "." + parts[2]

	if _, err := VerifyTransient(tampered, jwtSecret, jwtIssuer); err == nil {
		t.Fatal("a tampered state token was accepted")
	}
}

// An expired state token must be rejected, which is what bounds the window in
// which a leaked OAuth callback can be replayed.
func TestTransientStateExpires(t *testing.T) {
	past := time.Now().Add(-time.Hour)
	forged := jwt.NewWithClaims(jwt.SigningMethodHS256, TransientClaims{
		Nonce: "nonce-replay",
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    jwtIssuer,
			IssuedAt:  jwt.NewNumericDate(past),
			NotBefore: jwt.NewNumericDate(past),
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(-time.Minute)),
		},
	})

	raw, err := forged.SignedString([]byte(jwtSecret))
	if err != nil {
		t.Fatalf("sign forged state: %v", err)
	}
	if _, err := VerifyTransient(raw, jwtSecret, jwtIssuer); err == nil {
		t.Fatal("an expired state token was accepted")
	}
}
