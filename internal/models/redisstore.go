package models

import (
	"context"
	"log"

	"github.com/redis/go-redis/v9"
)

func SetNewURLRedis(rdb *redis.Client, longURL string, shortURL string) {
	ctx := context.Background()
	err := rdb.Set(ctx, "forward:"+longURL, shortURL, 0).Err()

	if err != nil {
		log.Println("failed to set new forward URL in redis", err)
	}
	err = rdb.Set(ctx, "backward:"+shortURL, longURL, 0).Err()

	if err != nil {
		log.Println("failed to set new backward URL in redis", err)
	}

}
