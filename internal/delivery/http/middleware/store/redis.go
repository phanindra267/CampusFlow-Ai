package store

import (
	"context"
	"fmt"
	"strconv"
	"time"

	goredis "github.com/redis/go-redis/v9"
	"golang.org/x/time/rate"
)

type RedisStore struct {
	client *goredis.Client
	prefix string
	limit  rate.Limit
	burst  int
	ttl    time.Duration
}

func NewRedisStore(client *goredis.Client, prefix string, perMinute, burst int) *RedisStore {
	if perMinute <= 0 || client == nil {
		return &RedisStore{}
	}
	b := burst
	if b <= 0 {
		b = perMinute
	}
	return &RedisStore{
		client: client,
		prefix: prefix,
		limit:  rate.Limit(float64(perMinute) / 60.0),
		burst:  b,
		ttl:    15 * time.Minute,
	}
}

func (s *RedisStore) key(k string) string {
	if s.prefix == "" {
		return "rl:" + k
	}
	return s.prefix + k
}

var allowLua = goredis.NewScript("local key=KEYS[1]\nlocal now=tonumber(ARGV[1])\nlocal ttl=tonumber(ARGV[2])\nlocal burst=tonumber(ARGV[3])\nlocal count=redis.call('INCR',key)\nif count==1 then redis.call('EXPIRE',key,ttl) end\nif count<=burst then return {1,0} else return {0,ttl} end")

func (s *RedisStore) Allow(ctx context.Context, key string, _ rate.Limit, _ int, now time.Time) (bool, time.Duration) {
	if s.client == nil {
		return true, 0
	}
	k := s.key(key)
	res, err := allowLua.Run(ctx, s.client, []string{k}, now.Unix(), int64(s.ttl.Seconds()), int64(s.burst)).Result()
	if err != nil {
		return true, 0
	}
	arr, ok := res.([]interface{})
	if !ok || len(arr) < 2 {
		return true, 0
	}
	okFlag, _ := strconv.ParseInt(fmt.Sprint(arr[0]), 10, 64)
	retry, _ := strconv.ParseInt(fmt.Sprint(arr[1]), 10, 64)
	if okFlag == 1 {
		return true, 0
	}
	d := time.Duration(retry) * time.Second
	if d < time.Second {
		d = time.Second
	}
	return false, d
}
