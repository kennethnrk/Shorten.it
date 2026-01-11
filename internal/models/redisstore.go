package models

import (
	"context"

	"github.com/redis/go-redis/v9"
)

func SetNewURLRedis(rdb *redis.Client, longURL string, shortURL string) error {
	ctx := context.Background()
	err := rdb.Set(ctx, "forward:"+longURL, shortURL, 0).Err()

	if err != nil {
		return err
	}
	err = rdb.Set(ctx, "backward:"+shortURL, longURL, 0).Err()

	if err != nil {
		return err
	}
	return nil
}
