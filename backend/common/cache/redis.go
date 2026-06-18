package cache

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/dispute-resolve/common/config"
	"github.com/redis/go-redis/v9"
)

var (
	client redis.UniversalClient
	once   sync.Once
)

func InitRedis(cfg *config.RedisConfig) redis.UniversalClient {
	once.Do(func() {
		if cfg.Cluster {
			client = redis.NewClusterClient(&redis.ClusterOptions{
				Addrs:        cfg.Addrs,
				Password:     cfg.Password,
				PoolSize:     cfg.PoolSize,
				MinIdleConns: 10,
				ReadTimeout:  time.Second * 3,
				WriteTimeout: time.Second * 3,
				PoolTimeout:  time.Second * 4,
			})
		} else {
			client = redis.NewClient(&redis.Options{
				Addr:         fmt.Sprintf("%s:%d", cfg.Host, cfg.Port),
				Password:     cfg.Password,
				DB:           cfg.DB,
				PoolSize:     cfg.PoolSize,
				MinIdleConns: 10,
				ReadTimeout:  time.Second * 3,
				WriteTimeout: time.Second * 3,
				PoolTimeout:  time.Second * 4,
			})
		}

		ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
		defer cancel()

		if err := client.Ping(ctx).Err(); err != nil {
			panic(fmt.Errorf("connect redis failed: %v", err))
		}
	})
	return client
}

func GetRedis() redis.UniversalClient {
	return client
}

func Get(ctx context.Context, key string) (string, error) {
	return client.Get(ctx, key).Result()
}

func Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	return client.Set(ctx, key, value, expiration).Err()
}

func Del(ctx context.Context, keys ...string) error {
	return client.Del(ctx, keys...).Err()
}

func Exists(ctx context.Context, key string) (bool, error) {
	result, err := client.Exists(ctx, key).Result()
	return result > 0, err
}

func Expire(ctx context.Context, key string, expiration time.Duration) error {
	return client.Expire(ctx, key, expiration).Err()
}

func Incr(ctx context.Context, key string) (int64, error) {
	return client.Incr(ctx, key).Result()
}

func Decr(ctx context.Context, key string) (int64, error) {
	return client.Decr(ctx, key).Result()
}

func Lock(ctx context.Context, key string, expiration time.Duration) (bool, error) {
	return client.SetNX(ctx, key, "1", expiration).Result()
}

func Unlock(ctx context.Context, key string) error {
	return client.Del(ctx, key).Err()
}

func DelByPrefix(ctx context.Context, prefix string) (int, error) {
	var deleted int
	var cursor uint64

	for {
		var keys []string
		var err error
		keys, cursor, err = client.Scan(ctx, cursor, prefix+"*", 100).Result()
		if err != nil {
			return deleted, err
		}

		if len(keys) > 0 {
			if err := client.Del(ctx, keys...).Err(); err != nil {
				return deleted, err
			}
			deleted += len(keys)
		}

		if cursor == 0 {
			break
		}
	}

	return deleted, nil
}
