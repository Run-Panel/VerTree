package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/Run-Panel/VerTree/internal/config"
	"github.com/Run-Panel/VerTree/internal/database"
	"github.com/Run-Panel/VerTree/internal/handlers/admin"
	"github.com/Run-Panel/VerTree/internal/handlers/auth"
	"github.com/Run-Panel/VerTree/internal/handlers/client"
	"github.com/Run-Panel/VerTree/internal/middleware"
	"github.com/Run-Panel/VerTree/internal/utils"
	"github.com/gin-gonic/gin"
)

func main() {
	// Load configuration
	cfg := config.Load()

	// Set Gin mode based on environment
	if cfg.App.Environment == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	// Initialize database
	if err := database.Initialize(cfg); err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	defer database.Close()

	// Run database migrations
	if err := database.AutoMigrate(); err != nil {
		log.Fatalf("Failed to run database migrations: %v", err)
	}

	// Seed default data
	if err := database.SeedDefaultData(); err != nil {
		log.Fatalf("Failed to seed default data: %v", err)
	}

	// Create router
	router := setupRouter(cfg)

	// Create HTTP server
	server := &http.Server{
		Addr:    cfg.Server.Host + ":" + cfg.Server.Port,
		Handler: router,
	}

	// Start server in goroutine
	go func() {
		log.Printf("Server starting on %s:%s", cfg.Server.Host, cfg.Server.Port)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	// Wait for interrupt signal to gracefully shutdown the server
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Shutting down server...")

	// Graceful shutdown with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := server.Shutdown(ctx); err != nil {
		log.Fatal("Server forced to shutdown:", err)
	}

	log.Println("Server exited")
}

func setupRouter(cfg *config.Config) *gin.Engine {
	router := gin.New()

	// Create rate limiters
	rateLimiters := middleware.CreateRateLimiters()

	// Create JWT manager
	jwtManager := utils.NewJWTManager(cfg.App.JWTSecret)

	// Global middleware (applied to all routes)
	router.Use(middleware.Logger())
	router.Use(middleware.RequestID())
	router.Use(middleware.SecurityHeaders())
	router.Use(middleware.XSSProtection())
	router.Use(middleware.SQLInjectionProtection())
	router.Use(middleware.CORS())
	router.Use(gin.Recovery())

	// Public health check endpoint (no auth required)
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":  "ok",
			"service": "runpanel-update-service",
			"version": "1.0.0",
			"region":  cfg.App.Region,
		})
	})

	// Initialize handlers
	authHandler := auth.NewAuthHandler(cfg.App.JWTSecret)
	versionHandler := admin.NewVersionHandler()
	channelHandler := admin.NewChannelHandler()
	statsHandler := admin.NewStatsHandler()
	updateHandler := client.NewUpdateHandler()

	// Auth API routes (public endpoints with rate limiting)
	authGroup := router.Group("/auth/api/v1")
	authGroup.Use(middleware.RateLimitByType(rateLimiters, "auth"))
	{
		authGroup.POST("/login", authHandler.Login)
		authGroup.POST("/refresh", authHandler.RefreshToken)
		authGroup.POST("/logout", authHandler.Logout)
	}

	// Admin API routes (protected with authentication and admin permissions)
	adminV1 := router.Group("/admin/api/v1")
	adminV1.Use(middleware.RateLimitByType(rateLimiters, "admin"))
	adminV1.Use(middleware.AuthMiddleware(jwtManager))
	adminV1.Use(middleware.RequireAdmin()) // Requires admin or superadmin role
	{
		// User profile management (authenticated users)
		adminV1.GET("/profile", authHandler.GetProfile)
		adminV1.POST("/change-password", authHandler.ChangePassword)

		// Version management
		versions := adminV1.Group("/versions")
		{
			versions.POST("", versionHandler.CreateVersion)
			versions.GET("", versionHandler.GetVersions)
			versions.GET("/:id", versionHandler.GetVersion)
			versions.PUT("/:id", versionHandler.UpdateVersion)
			versions.DELETE("/:id", versionHandler.DeleteVersion)
			versions.POST("/:id/publish", versionHandler.PublishVersion)
			versions.POST("/:id/unpublish", versionHandler.UnpublishVersion)
		}

		// Channel management
		channels := adminV1.Group("/channels")
		{
			channels.GET("", channelHandler.GetChannels)
			channels.GET("/:id", channelHandler.GetChannel)
			channels.POST("", channelHandler.CreateChannel)
			channels.PUT("/:id", channelHandler.UpdateChannel)
			channels.DELETE("/:id", channelHandler.DeleteChannel)
		}

		// Statistics
		adminV1.GET("/stats", statsHandler.GetStats)
		adminV1.GET("/stats/distribution", statsHandler.GetVersionDistribution)
		adminV1.GET("/stats/regions", statsHandler.GetRegionDistribution)

		// Admin management (superadmin only)
		admins := adminV1.Group("/admins")
		admins.Use(middleware.RequireSuperAdmin()) // Requires superadmin role
		{
			admins.GET("", authHandler.ListAdmins)
			admins.GET("/:id", authHandler.GetAdmin)
			admins.POST("", authHandler.CreateAdmin)
			admins.PUT("/:id", authHandler.UpdateAdmin)
			admins.DELETE("/:id", authHandler.DeleteAdmin)
		}
	}

	// Client API routes (public with rate limiting)
	clientV1 := router.Group("/api/v1")
	clientV1.Use(middleware.RateLimitByType(rateLimiters, "client"))
	{
		clientV1.POST("/check-update", updateHandler.CheckUpdate)
		clientV1.POST("/download-started", updateHandler.DownloadStarted)
		clientV1.POST("/install-result", updateHandler.InstallResult)
	}

	// Serve static files for admin frontend (public access)
	router.Static("/admin-ui", "./web/admin")

	// SPA fallback for admin frontend - handle all sub-routes
	router.NoRoute(func(c *gin.Context) {
		// Check if the request is for admin-ui
		if strings.HasPrefix(c.Request.URL.Path, "/admin-ui/") {
			// Serve index.html for SPA routing
			c.File("./web/admin/index.html")
			return
		}
		// Default 404 for other routes
		c.JSON(http.StatusNotFound, gin.H{
			"error": "Route not found",
			"path":  c.Request.URL.Path,
		})
	})

	// Redirect root to admin frontend
	router.GET("/", func(c *gin.Context) {
		c.Redirect(http.StatusMovedPermanently, "/admin-ui/")
	})

	return router
}
