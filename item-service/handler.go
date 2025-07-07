package main

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
	"go.opentelemetry.io/otel"
)

var tracer = otel.Tracer("item-service")

func GetStock(rdb *redis.Client) gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, span := tracer.Start(c.Request.Context(), "GetStock")
		defer span.End()

		traceID := span.SpanContext().TraceID()
		println("TraceID:", traceID.String())

		item := c.Param("item")
		val, err := rdb.Get(ctx, item).Result()
		if err == redis.Nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "item not found"})
			return
		} else if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"item": item, "stock": val})
	}
}

func SetStock(rdb *redis.Client) gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, span := tracer.Start(c.Request.Context(), "SetStock")
		defer span.End()

		traceID := span.SpanContext().TraceID()
		println("TraceID:", traceID.String())

		item := c.Param("item")
		var req struct {
			Stock int `json:"stock"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		if err := rdb.Set(ctx, item, req.Stock, 0).Err(); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		ItemCreatedCounter.Inc()
		ItemStockGauge.WithLabelValues(item).Set(float64(req.Stock))

		c.JSON(http.StatusOK, gin.H{"item": item, "stock": req.Stock})
	}
}

func DecrementStock(rdb *redis.Client) gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, span := tracer.Start(c.Request.Context(), "DecrementStock")
		defer span.End()

		item := c.Param("item")
		var req struct {
			Quantity int `json:"quantity"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		if err := rdb.DecrBy(ctx, item, int64(req.Quantity)).Err(); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to decrement stock"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"message": "stock decremented"})
	}
}

func DeleteStock(rdb *redis.Client) gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, span := tracer.Start(c.Request.Context(), "DeleteStock")
		defer span.End()

		traceID := span.SpanContext().TraceID()
		println("TraceID:", traceID.String())

		item := c.Param("item")
		_, err := rdb.Del(ctx, item).Result()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		ItemDeletedCounter.Inc()
		ItemStockGauge.DeleteLabelValues(item)

		c.JSON(http.StatusOK, gin.H{"message": "item deleted", "item": item})
	}
}

func ListStock(rdb *redis.Client) gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, span := tracer.Start(c.Request.Context(), "ListStock")
		defer span.End()

		traceID := span.SpanContext().TraceID()
		println("TraceID:", traceID.String())

		keys, err := rdb.Keys(ctx, "*").Result()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		result := make(map[string]int)
		for _, key := range keys {
			val, err := rdb.Get(ctx, key).Int()
			if err != nil {
				continue
			}
			result[key] = val
			ItemStockGauge.WithLabelValues(key).Set(float64(val))
		}

		c.JSON(http.StatusOK, result)
	}
}
