package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/hibiken/asynq"
	"go.uber.org/zap"

	"github.com/qvora/worker/internal/task"
)

func main() {
	log, _ := zap.NewProduction()
	defer log.Sync() //nolint:errcheck

	// -------------------------------------------------------------------------
	// Railway Redis — TCP only (asynq requires persistent BLPOP connection)
	// NEVER use Upstash HTTP Redis here — it does not support BLPOP
	// -------------------------------------------------------------------------
	redisURL := os.Getenv("RAILWAY_REDIS_URL")
	if redisURL == "" {
		log.Fatal("RAILWAY_REDIS_URL is required for asynq worker")
	}

	redisOpt, err := asynq.ParseRedisURI(redisURL)
	if err != nil {
		log.Fatal("invalid RAILWAY_REDIS_URL", zap.Error(err))
	}

	srv := asynq.NewServer(redisOpt, asynq.Config{
		Concurrency: 10,
		Queues: map[string]int{
			"critical": 6, // postprocess, webhook delivery
			"default":  3, // generation jobs
			"low":      1, // cleanup, analytics sync
		},
		ErrorHandler: asynq.ErrorHandlerFunc(func(ctx context.Context, t *asynq.Task, err error) {
			log.Error("task failed",
				zap.String("type", t.Type()),
				zap.Error(err),
			)
		}),
	})

	mux := asynq.NewServeMux()

	// Register task handlers
	mux.HandleFunc(task.TypeScrape, task.HandleScrape)
	mux.HandleFunc(task.TypeGenerate, task.HandleGenerate)
	mux.HandleFunc(task.TypePostprocess, task.HandlePostprocess)

	log.Info("starting qvora worker")

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)

	go func() {
		if err := srv.Run(mux); err != nil {
			log.Fatal("worker run failed", zap.Error(err))
		}
	}()

	<-quit
	log.Info("shutting down worker")
	srv.Shutdown()
}
