package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"
	"time"

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
	mux.HandleFunc(task.TypeSignalRecommendationsRefresh, task.HandleSignalRecommendationsRefresh)
	mux.HandleFunc(task.TypeSignalMetricsSync, task.HandleSignalMetricsSync)
	mux.HandleFunc(task.TypeSignalGDPRCleanup, task.HandleSignalGDPRCleanup)
	mux.HandleFunc(task.TypeJobReconcileStuck, task.HandleJobReconcile)

	scheduler := asynq.NewScheduler(redisOpt, &asynq.SchedulerOpts{
		Location: time.UTC,
	})
	weeklyRefreshTask, err := task.NewSignalRecommendationsRefreshTask(90)
	if err != nil {
		log.Fatal("failed to create signal recommendations refresh task", zap.Error(err))
	}
	if _, err := scheduler.Register("@weekly", weeklyRefreshTask); err != nil {
		log.Fatal("failed to register weekly signal recommendations refresh", zap.Error(err))
	}

	metricsSyncTask, err := task.NewSignalMetricsSyncTask()
	if err != nil {
		log.Fatal("failed to create signal metrics sync task", zap.Error(err))
	}
	if _, err := scheduler.Register("@every 6h", metricsSyncTask); err != nil {
		log.Fatal("failed to register signal metrics sync scheduler", zap.Error(err))
	}

	gdprCleanupTask, err := task.NewSignalGDPRCleanupTask()
	if err != nil {
		log.Fatal("failed to create signal gdpr cleanup task", zap.Error(err))
	}
	if _, err := scheduler.Register("@daily", gdprCleanupTask); err != nil {
		log.Fatal("failed to register signal gdpr cleanup scheduler", zap.Error(err))
	}

	reconcileTask, err := task.NewJobReconcileTask(45)
	if err != nil {
		log.Fatal("failed to create job reconcile task", zap.Error(err))
	}
	if _, err := scheduler.Register("@every 15m", reconcileTask); err != nil {
		log.Fatal("failed to register job reconcile scheduler", zap.Error(err))
	}

	log.Info("starting qvora worker")

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)

	go func() {
		if err := srv.Run(mux); err != nil {
			log.Fatal("worker run failed", zap.Error(err))
		}
	}()

	go func() {
		if err := scheduler.Run(); err != nil {
			log.Fatal("worker scheduler run failed", zap.Error(err))
		}
	}()

	<-quit
	log.Info("shutting down worker")
	scheduler.Shutdown()
	srv.Shutdown()
}
