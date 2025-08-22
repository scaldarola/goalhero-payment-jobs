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

	// Start background job manager (will be handled differently in serverless)
	if os.Getenv("GO_ENV") != "production" {
		fmt.Print("⚙️ Starting background jobs in non-production environment...\n")
		jobManager = services.StartBackgroundJobs()
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
	// For local development
	//if os.Getenv("GO_ENV") != "production" {
	log.Println("🔧 Running in development mode on :8081")
	router.Run(":8081")
	//}
}

// Handler for Vercel - this is the entry point for serverless functions
func Handler(w http.ResponseWriter, r *http.Request) {
	router.ServeHTTP(w, r)
}
