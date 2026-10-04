package auth

import (
	"os"
	"strconv"
	"sync"

	"golang.org/x/crypto/bcrypt"
)

const (
	// DefaultBcryptCost is the work factor used unless BCRYPT_COST overrides it.
	// OWASP recommends 10-12; 12 keeps a login well under a second while
	// remaining expensive to attack offline.
	DefaultBcryptCost = 12

	// MinimumTestCost is the cheapest legal work factor. Tests drop to it so
	// key stretching does not dominate the suite runtime. Never use it in
	// production configuration.
	MinimumTestCost = bcrypt.MinCost

	minBcryptCost = bcrypt.MinCost
	maxBcryptCost = 14
)

var (
	costMu      sync.RWMutex
	defaultCost = resolveConfiguredCost()
)

func resolveConfiguredCost() int {
	raw := os.Getenv("BCRYPT_COST")
	if raw == "" {
		return DefaultBcryptCost
	}
	cost, err := strconv.Atoi(raw)
	if err != nil || cost < minBcryptCost || cost > maxBcryptCost {
		return DefaultBcryptCost
	}
	return cost
}

// DefaultCost reports the bcrypt work factor that HashPassword will use.
func DefaultCost() int {
	costMu.RLock()
	defer costMu.RUnlock()
	return defaultCost
}

// SetDefaultCost overrides the work factor for subsequent HashPassword calls.
// Tests use this to avoid paying the full cost on every case.
func SetDefaultCost(cost int) {
	if cost < minBcryptCost || cost > maxBcryptCost {
		return
	}
	costMu.Lock()
	defaultCost = cost
	costMu.Unlock()
}

func HashPassword(password string) (string, error) {
	hashed, err := bcrypt.GenerateFromPassword([]byte(password), DefaultCost())
	if err != nil {
		return "", err
	}
	return string(hashed), nil
}

func CheckPasswordHash(password, hash string) bool {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(password)) == nil
}
