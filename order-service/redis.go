package main

import (
	"context"
	"log"
	"time"

	"github.com/go-redis/redis/extra/redisotel"
	"github.com/go-redis/redis/v8"
)

var Ctx = context.Background()

func InitRedis() *redis.Client {
	client := redis.NewClient(&redis.Options{
		Addr: "redis-order:6379",
	})

	client.AddHook(redisotel.TracingHook{})

	log.Println("Waiting for Redis (order-service) to start...")
	time.Sleep(3 * time.Second)

	if err := client.Ping(Ctx).Err(); err != nil {
		log.Fatalf("Failed to connect to Redis: %v", err)
	}

	return client
}
