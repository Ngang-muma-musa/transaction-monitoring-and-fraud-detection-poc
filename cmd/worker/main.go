package worker

import (
	"context"
	"frauddetection/internal/application"
	"frauddetection/internal/infrastructure/fraudengine"
	"frauddetection/internal/infrastructure/postgresql"
	"frauddetection/internal/infrastructure/queue"
	"frauddetection/internal/infrastructure/redis"
	worker "frauddetection/internal/infrastructure/woker"
	"log"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"
)

func main() {
	beanstalkAddr := os.Getenv("BEANSTALK_ADR")
	beanstalkTubeName := os.Getenv("BEANSTALK_TUBE_NAME")
	postgresqlDsn := os.Getenv("POSTGRESQL_DSN")
	maxWorkersStr := os.Getenv("MAX_WORKERS")
	redisDsn := os.Getenv("REDIS_DSN")
	maxWorkers, err := strconv.Atoi(maxWorkersStr)
	if err != nil {
		log.Fatalf("invalid MAX_WORKERS: %v", err)
	}
	conn, err := queue.BuildBeanstalkQueue(beanstalkAddr)
	if err != nil {
		log.Fatalf("failed to connect to beanstalk: %v", err)
	}
	defer conn.Close()
	beanstalkQueue := queue.NewBeanstalkQueue(conn, beanstalkTubeName)

	// Initialize repository
	db, err := postgresql.BuildPostgreSQLDB(postgresqlDsn)
	if err != nil {
		log.Fatalf("failed to connect to postgresql: %v", err)
	}
	defer db.Close()

	paymentRepo := postgresql.NewTransactionRepository(db)

	// Initialize Redis and rate limiter
	cache, err := redis.BuildRedisCache(redisDsn)
	if err != nil {
		log.Fatalf("failed to connect to redis: %v", err)
	}
	defer cache.Close()

	fe := fraudengine.NewFraudEngine(
		paymentRepo,
		fraudengine.NewMLClient(),
		cache,
	)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	for i := 1; i <= maxWorkers; i++ {
		worker := worker.NewWorker(
			ctx,
			beanstalkQueue,
			application.NewTransactionProcessor(paymentRepo, fe),
		)
		go func(id int) {
			log.Printf("Worker #%d started", id)
			worker.Start(ctx)
			log.Printf("Worker #%d stopped", id)
		}(i)
	}

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	cancel()

	time.Sleep(2 * time.Second)
	log.Println("Worker stopped.")
}
