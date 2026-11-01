package cache

import (
	"github.com/redis/go-redis/v9"
	"context"
)

var Client *redis.Client

func InitRedis(ctx context.Context, url string) error {
	opts, err := redis.ParseURL(url)
	if err != nil {
		return err
	}
	c := redis.NewClient(opts)
	if err := c.Ping(ctx).Err(); err != nil {
		c.Close()
		return err
	}
	Client = c
	return nil
}