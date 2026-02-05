package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"wedding-invitation-backend/internal/config"
	repo "wedding-invitation-backend/internal/repository/mongodb"
	"wedding-invitation-backend/internal/services"
	"wedding-invitation-backend/internal/utils"
	"wedding-invitation-backend/pkg/database"
)

func main() {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Initialize logger
	var logger *zap.Logger
	if cfg.IsProduction() {
		logger, _ = zap.NewProduction()
	} else {
		logger, _ = zap.NewDevelopment()
	}
	defer logger.Sync()

	// Connect to database
	db, err := database.NewMongoDB(&cfg.Database)
	if err != nil {
		logger.Fatal("Failed to connect to database", zap.Error(err))
	}
	defer db.Close(context.Background())

	// Ensure indexes
	if err := db.EnsureIndexes(context.Background()); err != nil {
		logger.Fatal("Failed to ensure indexes", zap.Error(err))
	}

	// Initialize repositories
	userRepo := repo.NewMongoUserRepository(db.Database)

	// Initialize JWT manager
	jwtManager := utils.NewJWTManager(
		cfg.Auth.JWTSecret,
		cfg.Auth.JWTRefreshSecret,
		cfg.Auth.AccessTokenTTL,
		cfg.Auth.RefreshTokenTTL,
		"wedding-invitation-api",
	)

	// Initialize services
	authService := services.NewAuthService(userRepo, jwtManager)

	// Setup router
	router := setupRouter(cfg, authService, jwtManager, logger)

	// Create HTTP server
	server := &http.Server{
		Addr:         ":" + cfg.Server.Port,
		Handler:      router,
		ReadTimeout:  cfg.Server.ReadTimeout,
		WriteTimeout: cfg.Server.WriteTimeout,
	}

	// Start server in a goroutine
	go func() {
		logger.Info("Starting server", zap.String("port", cfg.Server.Port))
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatal("Failed to start server", zap.Error(err))
		}
	}()

	// Wait for interrupt signal to gracefully shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("Shutting down server...")

	// Give outstanding requests 30 seconds to complete
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		logger.Fatal("Server forced to shutdown", zap.Error(err))
	}

	logger.Info("Server exited")
}

func setupRouter(
	cfg *config.Config,
	authService services.AuthService,
	jwtManager *utils.JWTManager,
	logger *zap.Logger,
) *gin.Engine {
	// Set gin mode
	if cfg.IsProduction() {
		gin.SetMode(gin.ReleaseMode)
	} else {
		gin.SetMode(gin.DebugMode)
	}

	router := gin.New()

	// Add basic middleware
	router.Use(gin.Logger())
	router.Use(gin.Recovery())

	// Health check endpoint
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":    "ok",
			"timestamp": time.Now().UTC(),
			"version":   "1.0.0",
		})
	})

	// API v1 routes
	v1 := router.Group("/api/v1")
	{
		// Authentication routes (public)
		auth := v1.Group("/auth")
		{
			auth.POST("/register", handleRegister(authService))
			auth.POST("/login", handleLogin(authService))
			auth.POST("/refresh", handleRefreshToken(authService, jwtManager))
			auth.POST("/logout", handleLogout())
			auth.POST("/forgot-password", handleForgotPassword(authService))
			auth.POST("/reset-password", handleResetPassword(authService))
			auth.POST("/verify-email", handleVerifyEmail(authService))
		}

		// Protected routes (temporarily without auth middleware)
		protected := v1.Group("/")
		{
			// User profile route
			protected.GET("/users/profile", handleGetProfile(authService))
			protected.PUT("/users/profile", handleUpdateProfile(authService))
			protected.PUT("/users/password", handleChangePassword(authService))
		}
	}

	return router
}

// Temporary handlers until proper handlers are implemented
func handleRegister(authService services.AuthService) gin.HandlerFunc {
	return func(c *gin.Context) {
		// TODO: Implement registration handler
		c.JSON(http.StatusNotImplemented, gin.H{"error": "Not implemented yet"})
	}
}

func handleLogin(authService services.AuthService) gin.HandlerFunc {
	return func(c *gin.Context) {
		// TODO: Implement login handler
		c.JSON(http.StatusNotImplemented, gin.H{"error": "Not implemented yet"})
	}
}

func handleRefreshToken(authService services.AuthService, jwtManager *utils.JWTManager) gin.HandlerFunc {
	return func(c *gin.Context) {
		// TODO: Implement refresh token handler
		c.JSON(http.StatusNotImplemented, gin.H{"error": "Not implemented yet"})
	}
}

func handleLogout() gin.HandlerFunc {
	return func(c *gin.Context) {
		// TODO: Implement logout handler
		c.JSON(http.StatusNotImplemented, gin.H{"error": "Not implemented yet"})
	}
}

func handleForgotPassword(authService services.AuthService) gin.HandlerFunc {
	return func(c *gin.Context) {
		// TODO: Implement forgot password handler
		c.JSON(http.StatusNotImplemented, gin.H{"error": "Not implemented yet"})
	}
}

func handleResetPassword(authService services.AuthService) gin.HandlerFunc {
	return func(c *gin.Context) {
		// TODO: Implement reset password handler
		c.JSON(http.StatusNotImplemented, gin.H{"error": "Not implemented yet"})
	}
}

func handleVerifyEmail(authService services.AuthService) gin.HandlerFunc {
	return func(c *gin.Context) {
		// TODO: Implement verify email handler
		c.JSON(http.StatusNotImplemented, gin.H{"error": "Not implemented yet"})
	}
}

func handleGetProfile(authService services.AuthService) gin.HandlerFunc {
	return func(c *gin.Context) {
		// TODO: Implement get profile handler
		c.JSON(http.StatusNotImplemented, gin.H{"error": "Not implemented yet"})
	}
}

func handleUpdateProfile(authService services.AuthService) gin.HandlerFunc {
	return func(c *gin.Context) {
		// TODO: Implement update profile handler
		c.JSON(http.StatusNotImplemented, gin.H{"error": "Not implemented yet"})
	}
}

func handleChangePassword(authService services.AuthService) gin.HandlerFunc {
	return func(c *gin.Context) {
		// TODO: Implement change password handler
		c.JSON(http.StatusNotImplemented, gin.H{"error": "Not implemented yet"})
	}
}
