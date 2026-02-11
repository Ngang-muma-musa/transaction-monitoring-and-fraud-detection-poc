// cmd/ws-gateway/main.go
package websocket

import (
	"context"
	"encoding/json"
	"frauddetection/internal/domain"
	"frauddetection/internal/infrastructure/redis"
	"frauddetection/internal/infrastructure/websocket"
	"frauddetection/internal/presentation/router"
	"frauddetection/pkg/helpers"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/labstack/echo/v4"
)

func main() {
	redisDsn := os.Getenv("REDIS_DSN")
	wsPort := helpers.GetEnvAsInt64("WEBSOCKET_PORT", 9000)
	ctx := context.Background()

	cache, err := redis.BuildRedisCache(redisDsn)
	if err != nil {
		log.Fatalf("failed to connect to redis: %v", err)
	}
	defer cache.Close()

	wsAdapter := websocket.NewWSAdapter()

	// Subscribe to Redis channel for fraud events
	go func() {
		pubsub := cache.Subscribe(ctx, "fraud_events")
		defer pubsub.Close()

		for msg := range pubsub.Channel() {
			var alert domain.FraudAlert
			if err := json.Unmarshal([]byte(msg.Payload), &alert); err == nil {
				wsAdapter.Notify(ctx, alert)
			}
		}
	}()

	// Register WebSocket routes
	wsRouter := router.NewWsRouter(wsAdapter)
	e := echo.New()
	wsRouter.Register(e)
	echoServer := router.NewServer(e, wsPort)

	// Start server
	echoServer.Start()

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	ctxShutdown, cancelShutdown := context.WithTimeout(ctx, 10*time.Second)
	defer cancelShutdown()

	echoServer.Shutdown(ctxShutdown)
}
