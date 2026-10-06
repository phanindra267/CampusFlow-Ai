package config

import (
	"fmt"
	"os"
	"strings"
	"time"
)

type RedisConfig struct {
	Host         string
	Port         string
	Password     string
	DB           int
	PoolSize     int
	MinIdleConns int
	DialTimeout  time.Duration
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
	IdleTimeout  time.Duration
}

func (r RedisConfig) Enabled() bool {
	if strings.TrimSpace(r.Host) != "" {
		return true
	}
	if strings.TrimSpace(os.Getenv("REDIS_URL")) != "" {
		return true
	}
	return false
}

func (r RedisConfig) Addr() string {
	if r.Port == "" {
		r.Port = "6379"
	}
	return fmt.Sprintf("%s:%s", r.Host, r.Port)
}
