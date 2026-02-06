package middleware

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"golang.org/x/time/rate"
)

func TestRateLimiter(t *testing.T) {
	gin.SetMode(gin.TestMode)

	config := RateLimiterConfig{
		Rate:            rate.Every(100 * time.Millisecond), // 10 requests per second
		Burst:           5,
		CleanupInterval: time.Minute,
		EntryTTL:        time.Minute,
	}

	rl := NewRateLimiter(config)
	router := gin.New()
	router.Use(rl.Middleware())

	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "success"})
	})

	// Test that requests within limit are allowed
	for i := 0; i < 5; i++ {
		req := httptest.NewRequest("GET", "/test", nil)
		req.RemoteAddr = "127.0.0.1"
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
	}

	// Test that request exceeding limit is blocked
	req := httptest.NewRequest("GET", "/test", nil)
	req.RemoteAddr = "127.0.0.1"
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusTooManyRequests, w.Code)
}

func TestMultiRateLimiter(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mrl := NewMultiRateLimiter()
	router := gin.New()
	router.Use(mrl.Middleware())

	router.GET("/api/v1/auth/login", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "login"})
	})

	router.GET("/api/v1/public/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "public"})
	})

	// Test auth endpoint rate limiting
	for i := 0; i < 6; i++ {
		req := httptest.NewRequest("GET", "/api/v1/auth/login", nil)
		req.RemoteAddr = "127.0.0.1"
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if i < 5 {
			assert.Equal(t, http.StatusOK, w.Code)
		} else {
			assert.Equal(t, http.StatusTooManyRequests, w.Code)
		}
	}
}

func TestSecurityHeaders(t *testing.T) {
	gin.SetMode(gin.TestMode)

	config := DefaultSecurityHeadersConfig()
	router := gin.New()
	router.Use(SecurityHeaders(config))

	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "success"})
	})

	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "DENY", w.Header().Get("X-Frame-Options"))
	assert.Equal(t, "nosniff", w.Header().Get("X-Content-Type-Options"))
	assert.NotEmpty(t, w.Header().Get("Content-Security-Policy"))
}

func TestCORSSecurity(t *testing.T) {
	gin.SetMode(gin.TestMode)

	config := DefaultCORSSecurityConfig()
	router := gin.New()
	router.Use(CORSSecurityMiddleware(config))

	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "success"})
	})

	// Test preflight request
	req := httptest.NewRequest("OPTIONS", "/test", nil)
	req.Header.Set("Origin", "http://localhost:3000")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNoContent, w.Code)
	assert.Equal(t, "http://localhost:3000", w.Header().Get("Access-Control-Allow-Origin"))

	// Test actual request
	req = httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("Origin", "http://localhost:3000")
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "http://localhost:3000", w.Header().Get("Access-Control-Allow-Origin"))
}

func TestBruteForceProtector(t *testing.T) {
	gin.SetMode(gin.TestMode)

	config := BruteForceConfig{
		MaxAttempts:     3,
		AttemptWindow:   time.Minute,
		BlockDuration:   time.Minute,
		CleanupInterval: time.Minute,
		TrackByIP:       true,
		TrackByEmail:    false,
	}

	bfp := NewBruteForceProtector(config)
	router := gin.New()
	router.Use(bfp.Middleware())

	router.POST("/api/v1/auth/login", func(c *gin.Context) {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid credentials"})
	})

	// Test failed attempts
	for i := 0; i < 4; i++ {
		req := httptest.NewRequest("POST", "/api/v1/auth/login", nil)
		req.RemoteAddr = "127.0.0.1"
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if i < 3 {
			assert.Equal(t, http.StatusUnauthorized, w.Code)
		} else {
			assert.Equal(t, http.StatusTooManyRequests, w.Code)
		}
	}
}

func TestValidationMiddleware(t *testing.T) {
	gin.SetMode(gin.TestMode)

	vm := NewValidationMiddleware()
	router := gin.New()

	type TestRequest struct {
		Name  string `json:"name" validate:"required"`
		Email string `json:"email" validate:"required,email"`
		Slug  string `json:"slug" validate:"slug"`
	}

	router.POST("/test", vm.ValidateBody(TestRequest{}), func(c *gin.Context) {
		req := c.MustGet("validated_request").(*TestRequest)
		c.JSON(http.StatusOK, gin.H{"name": req.Name})
	})

	// Test valid request
	req := httptest.NewRequest("POST", "/test", strings.NewReader(`{"name":"John","email":"john@example.com","slug":"test-slug"}`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	// Test invalid request
	req = httptest.NewRequest("POST", "/test", strings.NewReader(`{"name":"","email":"invalid","slug":"invalid slug"}`))
	req.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestSecurityIntegration(t *testing.T) {
	gin.SetMode(gin.TestMode)

	config := DefaultSecurityConfig()
	config.RateLimiting.Enabled = false // Disable for simpler testing

	sm := NewSecurityMiddleware(nil, config)
	router := gin.New()
	sm.SetupMiddleware(router)

	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "success"})
	})

	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "DENY", w.Header().Get("X-Frame-Options"))
}

func TestSecurityMiddlewareBuilder(t *testing.T) {
	gin.SetMode(gin.TestMode)

	builder := NewSecurityMiddlewareBuilder(nil)
	builder.WithEnvironment("development")
	builder.WithCORSOrigins([]string{"http://localhost:3000"})
	builder.WithRateLimiting(false)

	sm := builder.Build()
	router := gin.New()
	sm.SetupMiddleware(router)

	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "success"})
	})

	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}
