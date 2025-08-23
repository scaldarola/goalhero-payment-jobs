package handler

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/sebastiancaldarola/goalhero-payment-jobs/auth"
	"github.com/sebastiancaldarola/goalhero-payment-jobs/config"
	"github.com/sebastiancaldarola/goalhero-payment-jobs/handlers"
	"github.com/sebastiancaldarola/goalhero-payment-jobs/services"
)

var router *gin.Engine
var jobManager *services.BackgroundJobManager

func init() {
	log.Println("🔧 Initializing GoalHero Payment Jobs Service...")

	// Initialize configuration
	config.InitJobsConfig()

	// Initialize Firebase
	auth.InitFirebase()

	// Start background job manager (Railway supports long-running processes)
	if os.Getenv("DISABLE_BACKGROUND_JOBS") != "true" {
		fmt.Print("⚙️ Starting background jobs...\n")
		jobManager = services.StartBackgroundJobs()
	} else {
		log.Println("⚠️ Background jobs disabled via DISABLE_BACKGROUND_JOBS environment variable")
	}

	// Setup HTTP server
	gin.SetMode(gin.ReleaseMode)
	router = gin.New()
	router.Use(gin.Logger())
	router.Use(gin.Recovery())

	// Health check
	router.GET("/", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"service": "goalhero-payment-jobs", "status": "healthy"})
	})

	router.GET("/ping", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"service": "goalhero-payment-jobs", "status": "healthy"})
	})

	// API routes
	api := router.Group("/api/jobs")
	{
		// Job status and monitoring
		api.GET("/status", handlers.GetJobStatuses)
		api.GET("/health", handlers.GetJobHealth)

		// Job control (admin only)
		adminApi := api.Group("")
		adminApi.Use(auth.FirebaseAuthMiddleware())
		{
			adminApi.POST("/trigger/:jobName", handlers.TriggerJob)
			adminApi.POST("/config", handlers.UpdateJobConfig)
			adminApi.GET("/config", handlers.GetJobConfig)
			adminApi.POST("/restart", handlers.RestartJobs)
		}

		// Inter-service communication (no auth required for internal calls)
		internal := api.Group("/internal")
		{
			internal.POST("/trigger-rating-reminder", handlers.TriggerRatingReminder)
			internal.POST("/trigger-auto-release", handlers.TriggerAutoRelease)
			internal.POST("/trigger-dispute-escalation", handlers.TriggerDisputeEscalation)
		}
	}

	log.Println("✅ GoalHero Payment Jobs Service initialized")
}

func main() {
	// Get port from environment variable (Railway sets PORT automatically)
	port := os.Getenv("PORT")
	if port == "" {
		port = "8081" // Default for local development
	}
	
	log.Printf("🚀 Starting GoalHero Payment Jobs Service on port %s", port)
	router.Run(":" + port)
}

// Handler for Vercel - this is the entry point for serverless functions
func Handler(w http.ResponseWriter, r *http.Request) {
	router.ServeHTTP(w, r)
}
