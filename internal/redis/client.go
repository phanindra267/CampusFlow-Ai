package redis

import (
	"context"
	"time"

	"github.com/campuscare/api/internal/config"
	goredis "github.com/redis/go-redis/v9"
)

var (
	client *goredis.Client
)

func Init(cfg config.RedisConfig) error {
	if !cfg.Enabled() {
		return nil
	}
		c := goredis.NewClient(&goredis.Options{
			Addr:         cfg.Addr(),
			Password:     cfg.Password,
			DB:           cfg.DB,
			PoolSize:     cfg.PoolSize,
			MinIdleConns: cfg.MinIdleConns,
			DialTimeout:  cfg.DialTimeout,
			ReadTimeout:  cfg.ReadTimeout,
			WriteTimeout: cfg.WriteTimeout,
		})
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	if err := c.Ping(ctx).Err(); err != nil {
		return err
	}
	client = c
	return nil
}

func Client() *goredis.Client {
	return client
}

func Close() error {
	if client == nil {
		return nil
	}
	return client.Close()
}
