package redis

import (
	"context"
	"fmt"
	"os"

	"github.com/redis/go-redis/v9"
)

var ctx = context.Background()

type RedisClient struct {
	Client *redis.Client
}

func New() (*RedisClient, error) {
	conn := redis.NewClient(&redis.Options{
		Addr:     os.Getenv("REDIS_ADDR"),     // Redis server address
		Password: os.Getenv("REDIS_PASSWORD"), // No password set
		DB:       0,                           // Use default DB
	})
	// Ping the Redis server to check if the connection is successful
	_, err := conn.Ping(ctx).Result()
	if err != nil {
		panic(err) // Handle error appropriately in production code
	}
	fmt.Println("Connected to Redis successfully!")
	return &RedisClient{Client: conn}, nil
}
