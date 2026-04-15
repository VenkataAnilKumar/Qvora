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
		AllowMethods: []string{http.MethodGet, http.MethodPost, http.MethodPut, http.MethodDelete},
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

	// Workspaces
	workspaces := v1.Group("/workspaces", appmiddleware.RequireWorkspace())
	workspaces.GET("/:orgId", handler.GetWorkspace)
	workspaces.PUT("/:orgId/brand-kit", handler.UpsertBrandKit)

	// Assets
	assets := v1.Group("/assets", appmiddleware.RequireWorkspace())
	assets.GET("", handler.ListAssets)
	assets.DELETE("/:id", handler.DeleteAsset)

	// Variants
	variants := v1.Group("/variants", appmiddleware.RequireWorkspace())
	variants.GET("/:id/playback-url", handler.GetVariantPlaybackURL)

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
