package redis

import (
	"context"
	"fmt"

	"github.com/redis/go-redis/v9"
)

var ctx = context.Background()

type RedisClient struct {
	Client *redis.Client
}

func New() (*RedisClient, error) {
	conn := redis.NewClient(&redis.Options{
		Addr:     "redis-19433.c240.us-east-1-3.ec2.redns.redis-cloud.com:19433", // Redis server address
		Password: "sEOfYdii9huc4aqsRouD5vI1RdJzTb28",                             // No password set
		DB:       0,                                                              // Use default DB
	})
	// Ping the Redis server to check if the connection is successful
	_, err := conn.Ping(ctx).Result()
	if err != nil {
		panic(err) // Handle error appropriately in production code
	}
	fmt.Println("Connected to Redis successfully!")
	return &RedisClient{Client: conn}, nil
}
