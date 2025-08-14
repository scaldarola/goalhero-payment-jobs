package handlers

import (
	"fmt"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/sebastiancaldarola/goalhero-payment-jobs/services"
)

// GetJobStatuses handles GET /api/jobs/status
func GetJobStatuses(c *gin.Context) {
	log.Printf("[GetJobStatuses] Retrieving job statuses")

	statuses := services.GetJobStatuses()

	c.JSON(http.StatusOK, gin.H{
		"success":    true,
		"totalJobs":  len(statuses),
		"statuses":   statuses,
		"timestamp": c.GetHeader("X-Request-Time"),
	})
}

// GetJobHealth handles GET /api/jobs/health
func GetJobHealth(c *gin.Context) {
	log.Printf("[GetJobHealth] Performing health check")

	health := services.GetJobHealth()

	status := http.StatusOK
	if !health.Healthy {
		status = http.StatusServiceUnavailable
	}

	c.JSON(status, gin.H{
		"success": true,
		"health":  health,
	})
}

// TriggerJob handles POST /api/jobs/trigger/:jobName
func TriggerJob(c *gin.Context) {
	jobName := c.Param("jobName")
	log.Printf("[TriggerJob] Manual trigger requested for job: %s", jobName)

	var err error
	switch jobName {
	case "rating-reminder":
		err = services.TriggerRatingReminder()
	case "auto-release":
		err = services.TriggerAutoRelease()
	case "dispute-escalation":
		err = services.TriggerDisputeEscalation()
	default:
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Invalid job name",
			"validJobs": []string{"rating-reminder", "auto-release", "dispute-escalation"},
		})
		return
	}

	if err != nil {
		log.Printf("[TriggerJob] Failed to trigger job %s: %v", jobName, err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": fmt.Sprintf("Job %s triggered successfully", jobName),
		"jobName": jobName,
	})
}

// UpdateJobConfig handles POST /api/jobs/config
func UpdateJobConfig(c *gin.Context) {
	log.Printf("[UpdateJobConfig] Updating job configuration")

	var newConfig services.JobConfig
	if err := c.ShouldBindJSON(&newConfig); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Invalid configuration format",
			"details": err.Error(),
		})
		return
	}

	if err := services.UpdateJobConfig(&newConfig); err != nil {
		log.Printf("[UpdateJobConfig] Failed to update configuration: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Job configuration updated successfully",
		"config":  newConfig,
	})
}

// GetJobConfig handles GET /api/jobs/config
func GetJobConfig(c *gin.Context) {
	log.Printf("[GetJobConfig] Retrieving job configuration")

	config := services.GetJobConfig()
	if config == nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Job manager not initialized",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"config":  config,
	})
}

// RestartJobs handles POST /api/jobs/restart
func RestartJobs(c *gin.Context) {
	log.Printf("[RestartJobs] Restarting job system")

	// TODO: Implement job restart functionality
	c.JSON(http.StatusNotImplemented, gin.H{
		"success": false,
		"error":   "Job restart functionality not yet implemented",
	})
}

// Internal trigger handlers for inter-service communication
func TriggerRatingReminder(c *gin.Context) {
	log.Printf("[Internal] Rating reminder trigger received")

	err := services.TriggerRatingReminder()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Rating reminder job triggered",
	})
}

func TriggerAutoRelease(c *gin.Context) {
	log.Printf("[Internal] Auto release trigger received")

	err := services.TriggerAutoRelease()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Auto release job triggered",
	})
}

func TriggerDisputeEscalation(c *gin.Context) {
	log.Printf("[Internal] Dispute escalation trigger received")

	err := services.TriggerDisputeEscalation()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Dispute escalation job triggered",
	})
}