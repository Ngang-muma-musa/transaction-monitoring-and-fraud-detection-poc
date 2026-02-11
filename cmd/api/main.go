package api

import (
	"context"
	"frauddetection/internal/application"
	"frauddetection/internal/infrastructure/postgresql"
	"frauddetection/internal/infrastructure/redis"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	queue "frauddetection/internal/infrastructure/queue"
	"frauddetection/internal/presentation/restapi/handler"
	"frauddetection/internal/presentation/router"
	"frauddetection/pkg/helpers"

	"github.com/labstack/echo/v4"
)

func main() {
	beanstalkAddr := os.Getenv("BEANSTALK_ADR")
	beanstalkTubeName := os.Getenv("BEANSTALK_TUBE_NAME")
	redisDsn := os.Getenv("REDIS_DSN")
	rateLimit := helpers.GetEnvAsInt64("RATELIMIT_LINIT", 5)
	rateLimitWindow := helpers.GetEnvAsInt64("RATELIMIT_WINDOW", 1)
	appPort := helpers.GetEnvAsInt64("APP_PORT", 8080)
	postgresqlDsn := os.Getenv("POSTGRESQL_DSN")
	// Initialize Beanstalk queue
	conn, err := queue.BuildBeanstalkQueue(beanstalkAddr)
	if err != nil {
		log.Fatalf("failed to connect to beanstalk: %v", err)
	}
	defer conn.Close()

	beanstalkQueue := queue.NewBeanstalkQueue(conn, beanstalkTubeName)

	// Initialize Redis and rate limiter
	cache, err := redis.BuildRedisCache(redisDsn)
	if err != nil {
		log.Fatalf("failed to connect to redis: %v", err)
	}
	defer cache.Close()

	window := time.Duration(rateLimitWindow) * time.Minute
	ratelimiter := redis.NewRedisRateLimiter(cache, rateLimit, window)

	// Initialize repository
	db, err := postgresql.BuildPostgreSQLDB(postgresqlDsn)
	if err != nil {
		log.Fatalf("failed to connect to postgresql: %v", err)
	}
	defer db.Close()

	paymentRepo := postgresql.NewTransactionRepository(db)

	// Root context
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Application layer
	paymentServiceApp := application.NewPaymentService(
		paymentRepo,
		beanstalkQueue,
		ratelimiter,
	)

	// Presentation layer
	paymentServiceHandler := handler.NewPaymentServiceHandler(paymentServiceApp)
	apiRouter := router.NewRouter(paymentServiceHandler)

	e := echo.New()
	apiRouter.Register(e)
	echoServer := router.NewServer(e, appPort)

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
