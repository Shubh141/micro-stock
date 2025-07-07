package main

import (
	"context"
	"log"

	"github.com/gin-gonic/gin"
	"go.opentelemetry.io/contrib/instrumentation/github.com/gin-gonic/gin/otelgin"
)

func main() {
	ctx := context.Background()
	shutdown := InitTracer("item-service")
	defer shutdown(ctx)

	r := gin.Default()

	r.Use(otelgin.Middleware("item-service"))

	InitMetrics()
	r.Use(TrackDurationMiddleware(), TrackStatusMiddleware())
	r.GET("/metrics", PrometheusHandler())

	redisClient := InitRedis()

	r.GET("/healthz", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	r.GET("/stock", ListStock(redisClient))
	r.GET("/stock/:item", GetStock(redisClient))
	r.POST("/stock/:item", SetStock(redisClient))
	r.POST("/stock/:item/decrement", DecrementStock(redisClient))
	r.DELETE("/stock/:item", DeleteStock(redisClient))

	log.Println("item-service running on :8080")
	if err := r.Run(":8080"); err != nil {
		log.Fatalf("failed to run server: %v", err)
	}
}
