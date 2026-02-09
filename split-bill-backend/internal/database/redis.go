package database

import (
	"context"
	"log"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/splitbill/backend/internal/config"
)

type RedisClient struct {
	Client *redis.Client
}

func NewRedisClient(cfg *config.RedisConfig) *RedisClient {
	client := redis.NewClient(&redis.Options{
		Addr:     cfg.Addr,
		Password: cfg.Password,
		DB:       cfg.DB,
	})

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		log.Printf("⚠️  Warning: Could not connect to Redis: %v (continuing without cache)", err)
		return &RedisClient{Client: client}
	}

	log.Println("✅ Connected to Redis successfully")
	return &RedisClient{Client: client}
}

func (r *RedisClient) Close() {
	if err := r.Client.Close(); err != nil {
		log.Printf("Error closing Redis: %v", err)
	}
	log.Println("Disconnected from Redis")
}
