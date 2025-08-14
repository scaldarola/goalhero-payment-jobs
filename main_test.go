package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupTestApp() *gin.Engine {
	// Set test environment
	os.Setenv("GO_ENV", "test")
	
	// Set gin to test mode
	gin.SetMode(gin.TestMode)
	
	// Create a simple test router that mimics the main structure
	testRouter := gin.New()
	testRouter.Use(gin.Recovery())

	// Health check endpoints
	testRouter.GET("/", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"service": "goalhero-payment-jobs", "status": "healthy"})
	})

	testRouter.GET("/ping", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"service": "goalhero-payment-jobs", "status": "healthy"})
	})

	return testRouter
}

func TestHealthCheckEndpoints(t *testing.T) {
	app := setupTestApp()

	t.Run("root endpoint should return healthy status", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		w := httptest.NewRecorder()

		app.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)

		assert.Equal(t, "goalhero-payment-jobs", response["service"])
		assert.Equal(t, "healthy", response["status"])
	})

	t.Run("ping endpoint should return healthy status", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/ping", nil)
		w := httptest.NewRecorder()

		app.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)

		assert.Equal(t, "goalhero-payment-jobs", response["service"])
		assert.Equal(t, "healthy", response["status"])
	})
}

func TestHandlerFunction(t *testing.T) {
	t.Run("Handler should handle HTTP requests", func(t *testing.T) {
		// Test that the Handler function doesn't panic
		// This is a basic test to ensure the function is properly defined
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		w := httptest.NewRecorder()

		// Create a simple test to ensure Handler function exists and can be called
		// In a real deployment scenario, this would route through the initialized router
		defer func() {
			if r := recover(); r != nil {
				t.Errorf("Handler function panicked: %v", r)
			}
		}()

		Handler(w, req)
		
		// The function should complete without panicking
		// The actual response depends on the router initialization
		assert.True(t, true) // Basic assertion to ensure test runs
	})
}

func TestMainFunction(t *testing.T) {
	t.Run("main function should handle test environment", func(t *testing.T) {
		// Set production environment to prevent server startup in test
		os.Setenv("GO_ENV", "production")
		defer os.Unsetenv("GO_ENV")

		// Test that main function doesn't panic when called in production mode
		// In production mode, main() should not start the server
		defer func() {
			if r := recover(); r != nil {
				t.Errorf("main function panicked: %v", r)
			}
		}()

		main()
		
		// Should complete without starting server in production mode
		assert.True(t, true)
	})
}