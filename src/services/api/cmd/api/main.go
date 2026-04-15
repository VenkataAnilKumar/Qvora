package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"go.uber.org/zap"

	"github.com/qvora/api/internal/handler"
	appmiddleware "github.com/qvora/api/internal/middleware"
)

func main() {
	log, _ := zap.NewProduction()
	defer log.Sync() //nolint:errcheck

	e := echo.New()
	e.HideBanner = true
	e.HidePort = true

	// -------------------------------------------------------------------------
	// Global middleware
	// -------------------------------------------------------------------------
	e.Use(middleware.RecoverWithConfig(middleware.RecoverConfig{
		LogErrorFunc: func(c echo.Context, err error, stack []byte) error {
			log.Error("panic recovered", zap.Error(err), zap.ByteString("stack", stack))
			return nil
		},
	}))
	e.Use(middleware.RequestID())
	e.Use(middleware.SecureWithConfig(middleware.SecureConfig{
		XSSProtection:         "1; mode=block",
		ContentTypeNosniff:    "nosniff",
		XFrameOptions:         "DENY",
		HSTSMaxAge:            31536000,
		ContentSecurityPolicy: "default-src 'none'",
	}))
	e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins: []string{
			os.Getenv("NEXT_PUBLIC_APP_URL"),
		},
		AllowMethods: []string{http.MethodGet, http.MethodPost, http.MethodPut, http.MethodPatch, http.MethodDelete},
		AllowHeaders: []string{echo.HeaderAuthorization, echo.HeaderContentType, echo.HeaderAccept},
	}))
	e.Use(appmiddleware.ClerkAuth())
	e.Use(appmiddleware.RateLimiter())

	// -------------------------------------------------------------------------
	// Routes
	// -------------------------------------------------------------------------
	v1 := e.Group("/api/v1")

	// Health
	v1.GET("/health", handler.Health)

	// Jobs
	jobs := v1.Group("/jobs", appmiddleware.RequireWorkspace())
	jobs.POST("", handler.SubmitJob)
	jobs.GET("", handler.ListJobs)
	jobs.GET("/:id", handler.GetJob)
	jobs.PATCH("/:id/status", handler.UpdateJobStatus)
	jobs.GET("/:id/stream", handler.StreamJob)

	// Briefs
	briefs := v1.Group("/briefs", appmiddleware.RequireWorkspace())
	briefs.POST("", handler.CreateBrief)
	briefs.GET("", handler.ListBriefs)
	briefs.POST("/:briefId/batch-generate", handler.BatchGenerateVariants)

	// Workspaces
	workspaces := v1.Group("/workspaces", appmiddleware.RequireWorkspace())
	workspaces.GET("/:orgId", handler.GetWorkspace)
	workspaces.GET("/:orgId/brand-kit", handler.GetBrandKit)
	workspaces.GET("/:orgId/usage", handler.GetWorkspaceUsage)
	workspaces.PUT("/:orgId/brand-kit", handler.UpsertBrandKit)
	workspaces.PATCH("/:orgId/usage/reset", handler.ResetWorkspaceUsage)
	workspaces.PATCH("/:orgId/memberships/sync", handler.SyncWorkspaceMembership)
	workspaces.PATCH("/:orgId/lifecycle", handler.UpdateWorkspaceLifecycle)
	workspaces.PATCH("/:orgId/subscription", handler.UpdateWorkspaceSubscription)

	// Assets
	assets := v1.Group("/assets", appmiddleware.RequireWorkspace())
	assets.GET("", handler.ListAssets)
	assets.DELETE("/:id", handler.DeleteAsset)

	// Exports
	exports := v1.Group("/exports", appmiddleware.RequireWorkspace())
	exports.GET("", handler.ListExports)

	// Signal (Phase 2 kickoff)
	signalRoutes := v1.Group("/signal", appmiddleware.RequireWorkspace())
	signalRoutes.GET("/connections", handler.ListSignalConnections)
	signalRoutes.PUT("/connections/:platform", handler.UpsertSignalConnection)
	signalRoutes.PATCH("/connections/:platform/:accountId/health", handler.PatchSignalConnectionHealth)
	signalRoutes.GET("/oauth/:platform/initiate", handler.InitiateSignalOAuth)
	signalRoutes.GET("/dashboard", handler.GetSignalDashboard)
	signalRoutes.POST("/metrics", handler.UpsertSignalMetrics)
	signalRoutes.GET("/fatigue", handler.DetectSignalFatigue)
	signalRoutes.GET("/fatigue/events", handler.ListSignalFatigueEvents)
	signalRoutes.GET("/recommendations", handler.GetSignalRecommendations)
	signalRoutes.POST("/recommendations/feedback", handler.CreateSignalRecommendationFeedback)
	signalRoutes.GET("/recommendations/feedback", handler.ListSignalRecommendationFeedbackByAngle)
	v1.GET("/signal/oauth/:platform/callback", handler.HandleSignalOAuthCallback)
	v1.POST("/internal/signal/metrics/sync-all", handler.SyncSignalMetricsAll)
	v1.POST("/internal/signal/gdpr/cleanup", handler.HandleSignalGDPRCleanup)
	v1.POST("/internal/signal/recommendations/refresh-all", handler.RefreshAllSignalRecommendations)
	v1.POST("/internal/postprocess/callback", handler.HandlePostprocessCallback)
	v1.POST("/internal/jobs/reconcile-stuck", handler.ReconcileStuckJobs)

	// Variants
	variants := v1.Group("/variants", appmiddleware.RequireWorkspace())
	variants.GET("/:id/playback-url", handler.GetVariantPlaybackURL)
	variants.PATCH("/:id/fal-request", handler.UpdateVariantFalRequest)

	// Webhooks (no auth middleware — verified by signature)
	webhooks := e.Group("/webhooks")
	webhooks.POST("/mux", handler.MuxWebhook)
	webhooks.POST("/fal", handler.FalWebhook)

	// -------------------------------------------------------------------------
	// Start with graceful shutdown
	// -------------------------------------------------------------------------
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	go func() {
		log.Info("starting qvora api", zap.String("port", port))
		if err := e.Start(":" + port); err != nil && err != http.ErrServerClosed {
			log.Fatal("server error", zap.Error(err))
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
	<-quit

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := e.Shutdown(ctx); err != nil {
		log.Fatal("graceful shutdown failed", zap.Error(err))
	}
	log.Info("server stopped")
}
