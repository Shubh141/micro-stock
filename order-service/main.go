package main

import (
	"log"

	"github.com/gin-gonic/gin"
)

func main() {
	r := gin.Default()

	InitMetrics()
	r.Use(TrackDurationMiddleware(), TrackStatusMiddleware())
	r.GET("/metrics", PrometheusHandler())

	shutdown := InitTracer("order-service")
	defer shutdown(Ctx)

	redisClient := InitRedis()

	r.GET("/healthz", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	r.POST("/order", PlaceOrder(redisClient))

	log.Println("order-service running on :8081")
	if err := r.Run(":8081"); err != nil {
		log.Fatalf("failed to run server: %v", err)
	}
}
