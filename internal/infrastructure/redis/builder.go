package redis

import (
	"strings"

	"github.com/go-redis/redis/v8"
)

func BuildRedisCache(redisDsn string) (*redis.Client, error) {
	redisDsn = strings.Trim(redisDsn, `"`)

	// Parse connection string from dsn
	opt, err := redis.ParseURL(redisDsn)
	if err != nil {
		return nil, err
	}

	// Connect to redis
	db := redis.NewClient(opt)
	return db, nil
}
