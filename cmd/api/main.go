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
	"wedding-invitation-backend/internal/handlers"
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
	weddingRepo := repo.NewMongoWeddingRepository(db.Database)
	rsvpRepo := repo.NewMongoRSVPRepository(db.Database)

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
	userService := services.NewUserService(userRepo)
	weddingService := services.NewWeddingService(weddingRepo, userRepo)
	rsvpService := services.NewRSVPService(rsvpRepo, weddingRepo)

	// Initialize handlers
	userHandler := handlers.NewUserHandler(userService)
	weddingHandler := handlers.NewWeddingHandler(weddingService)
	rsvpHandler := handlers.NewRSVPHandler(rsvpService)

	// Setup router
	router := setupRouter(cfg, authService, userHandler, weddingHandler, rsvpHandler, jwtManager, logger)

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
	userHandler *handlers.UserHandler,
	weddingHandler *handlers.WeddingHandler,
	rsvpHandler *handlers.RSVPHandler,
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
			auth.POST("/register", func(c *gin.Context) {
				var req services.RegisterRequest
				if err := c.ShouldBindJSON(&req); err != nil {
					c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input: " + err.Error()})
					return
				}

				result, err := authService.Register(c.Request.Context(), req)
				if err != nil {
					c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
					return
				}

				c.JSON(http.StatusCreated, result)
			})

			auth.POST("/login", func(c *gin.Context) {
				var req services.LoginRequest
				if err := c.ShouldBindJSON(&req); err != nil {
					c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input: " + err.Error()})
					return
				}

				result, err := authService.Login(c.Request.Context(), req)
				if err != nil {
					c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
					return
				}

				c.JSON(http.StatusOK, result)
			})

			auth.POST("/refresh", func(c *gin.Context) {
				var req struct {
					RefreshToken string `json:"refresh_token" validate:"required"`
				}
				if err := c.ShouldBindJSON(&req); err != nil {
					c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input: " + err.Error()})
					return
				}

				result, err := authService.RefreshToken(c.Request.Context(), req.RefreshToken)
				if err != nil {
					c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
					return
				}

				c.JSON(http.StatusOK, result)
			})

			auth.POST("/logout", func(c *gin.Context) {
				// TODO: Extract user ID from token for proper logout
				c.JSON(http.StatusOK, gin.H{"message": "Logged out successfully"})
			})

			auth.POST("/forgot-password", func(c *gin.Context) {
				var req struct {
					Email string `json:"email" validate:"required,email"`
				}
				if err := c.ShouldBindJSON(&req); err != nil {
					c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input: " + err.Error()})
					return
				}

				result, err := authService.ForgotPassword(c.Request.Context(), req.Email)
				if err != nil {
					c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
					return
				}

				c.JSON(http.StatusOK, result)
			})

			auth.POST("/reset-password", func(c *gin.Context) {
				var req services.ResetPasswordRequest
				if err := c.ShouldBindJSON(&req); err != nil {
					c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input: " + err.Error()})
					return
				}

				err := authService.ResetPassword(c.Request.Context(), req)
				if err != nil {
					c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
					return
				}

				c.JSON(http.StatusOK, gin.H{"message": "Password reset successfully"})
			})

			auth.POST("/verify-email", func(c *gin.Context) {
				var req struct {
					Token string `json:"token" validate:"required"`
				}
				if err := c.ShouldBindJSON(&req); err != nil {
					c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input: " + err.Error()})
					return
				}

				err := authService.VerifyEmail(c.Request.Context(), req.Token)
				if err != nil {
					c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
					return
				}

				c.JSON(http.StatusOK, gin.H{"message": "Email verified successfully"})
			})
		}

		// Protected routes (temporarily without auth middleware)
		protected := v1.Group("/")
		{
			// User profile routes
			protected.GET("/users/profile", userHandler.GetProfile)
			protected.PUT("/users/profile", userHandler.UpdateProfile)

			// User wedding routes
			protected.GET("/users/weddings", userHandler.GetUserWeddings)
			protected.POST("/users/weddings/:wedding_id", userHandler.AddWeddingToUser)
			protected.DELETE("/users/weddings/:wedding_id", userHandler.RemoveWeddingFromUser)

			// Wedding management routes
			protected.POST("/weddings", weddingHandler.CreateWedding)
			protected.GET("/weddings", weddingHandler.GetUserWeddings)
			protected.GET("/weddings/:id", weddingHandler.GetWedding)
			protected.PUT("/weddings/:id", weddingHandler.UpdateWedding)
			protected.DELETE("/weddings/:id", weddingHandler.DeleteWedding)
			protected.POST("/weddings/:id/publish", weddingHandler.PublishWedding)
			protected.GET("/weddings/slug/:slug", weddingHandler.GetWeddingBySlug)

			// RSVP management routes
			protected.GET("/weddings/:id/rsvps", rsvpHandler.GetRSVPs)
			protected.GET("/weddings/:id/rsvps/statistics", rsvpHandler.GetRSVPStatistics)
			protected.GET("/weddings/:id/rsvps/export", rsvpHandler.ExportRSVPs)
		}

		// Individual RSVP routes
		v1.PUT("/rsvps/:id", rsvpHandler.UpdateRSVP)
		v1.DELETE("/rsvps/:id", rsvpHandler.DeleteRSVP)

		// Public routes
		public := v1.Group("/public")
		{
			// Public wedding listings
			public.GET("/weddings", weddingHandler.ListPublicWeddings)
			
			// Public RSVP submission
			public.POST("/weddings/:id/rsvp", rsvpHandler.SubmitRSVP)
		}

		// Admin routes (temporarily without auth middleware)
		admin := v1.Group("/admin")
		{
			// User management routes
			admin.GET("/users", userHandler.GetUsersList)
			admin.GET("/users/search", userHandler.SearchUsers)
			admin.PUT("/users/:id/status", userHandler.UpdateUserStatus)
			admin.DELETE("/users/:id", userHandler.DeleteUser)
			admin.GET("/users/stats", userHandler.GetUserStats)
		}
	}

	return router
}
