package config

import (
	"os"
	"testing"
)

func TestLoad(t *testing.T) {
	// Set an environment variable
	os.Setenv("PORT", "9999")
	defer os.Unsetenv("PORT")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if cfg.Server.Port != "9999" {
		t.Errorf("Expected port to be 9999, got %s", cfg.Server.Port)
	}
}

func TestDatabaseConfig_DSN(t *testing.T) {
	cfg := DatabaseConfig{
		Host:     "host",
		Port:     "5432",
		User:     "user",
		Password: "password",
		Name:     "name",
		SSLMode:  "disable",
	}
	expected := "host=host port=5432 user=user password=password dbname=name sslmode=disable"
	if cfg.DSN() != expected {
		t.Errorf("Expected DSN %s, got %s", expected, cfg.DSN())
	}
}
