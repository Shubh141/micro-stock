package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/propagation"
)

var tracer = otel.Tracer("order-service")

func PlaceOrder(rdb *redis.Client) gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, span := tracer.Start(c.Request.Context(), "PlaceOrder")
		defer span.End()

		fmt.Println("order-service TraceID:", span.SpanContext().TraceID().String())

		var req struct {
			Item     string `json:"item"`
			Quantity int    `json:"quantity"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		span.SetAttributes(
			attribute.String("order.item", req.Item),
			attribute.Int("order.quantity", req.Quantity),
		)

		client := http.Client{
			Transport: otelhttp.NewTransport(http.DefaultTransport),
		}

		// GET /stock/:item from item-service
		getURL := fmt.Sprintf("http://item-service:8080/stock/%s", req.Item)

		getCtx, getSpan := tracer.Start(ctx, "GET item-service /stock/:item")
		fmt.Println("order-service GET TraceID:", getSpan.SpanContext().TraceID().String())

		getReq, err := http.NewRequestWithContext(getCtx, "GET", getURL, nil)
		if err != nil {
			getSpan.End()
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create stock request"})
			return
		}
		otel.GetTextMapPropagator().Inject(getCtx, propagation.HeaderCarrier(getReq.Header))

		getResp, err := client.Do(getReq)
		getSpan.End()
		if err != nil {
			c.JSON(http.StatusServiceUnavailable, gin.H{"error": "item-service unreachable"})
			return
		}
		defer getResp.Body.Close()

		if getResp.StatusCode != http.StatusOK {
			c.JSON(http.StatusNotFound, gin.H{"error": "item not found"})
			return
		}

		body, _ := io.ReadAll(getResp.Body)
		var itemResp struct {
			Item  string `json:"item"`
			Stock string `json:"stock"`
		}
		if err := json.Unmarshal(body, &itemResp); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "invalid response from item-service"})
			return
		}

		var currentStock int
		fmt.Sscanf(itemResp.Stock, "%d", &currentStock)

		if req.Quantity > currentStock {
			c.JSON(http.StatusConflict, gin.H{"error": "not enough stock"})
			return
		}

		// Redis order increment
		redisCtx, redisSpan := tracer.Start(ctx, "INCR Redis order count")
		orderKey := fmt.Sprintf("order:%s", req.Item)
		err = rdb.IncrBy(redisCtx, orderKey, int64(req.Quantity)).Err()
		redisSpan.End()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to save order"})
			return
		}

		// POST /stock/:item/decrement to item-service
		decURL := fmt.Sprintf("http://item-service:8080/stock/%s/decrement", req.Item)
		payload := map[string]int{"quantity": req.Quantity}
		jsonBody, _ := json.Marshal(payload)

		decCtx, decSpan := tracer.Start(ctx, "POST item-service /stock/:item/decrement")
		fmt.Println("order-service POST TraceID:", decSpan.SpanContext().TraceID().String())

		postReq, err := http.NewRequestWithContext(decCtx, "POST", decURL, bytes.NewBuffer(jsonBody))
		if err != nil {
			decSpan.End()
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create decrement request"})
			return
		}
		postReq.Header.Set("Content-Type", "application/json")
		otel.GetTextMapPropagator().Inject(decCtx, propagation.HeaderCarrier(postReq.Header))

		decResp, err := client.Do(postReq)
		decSpan.End()
		if err != nil || decResp.StatusCode != http.StatusOK {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update stock in item-service"})
			return
		}

		OrdersPlacedCounter.WithLabelValues(req.Item).Inc()

		c.JSON(http.StatusOK, gin.H{
			"message":        "order placed",
			"item":           req.Item,
			"quantity":       req.Quantity,
			"remainingStock": currentStock - req.Quantity,
		})
	}
}
