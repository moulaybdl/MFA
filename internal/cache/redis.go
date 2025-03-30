package cache

import (
	"context"

	"github.com/redis/go-redis/v9"
)

var ctx = context.Background()

func NewRedisClient() (*redis.Client, error) {
	client := redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
		Password: "",
		DB: 0,
	})

	// test the connection:
	_, err := client.Ping(ctx).Result()
	if err != nil {
		return nil, err
	}

	return client, nil

}

