package config

import (
	"log"
	"os"
	"strconv"
	"time"

)

// JobsConfig holds configuration specific to jobs service
type JobsConfig struct {
	Port                     string
	MainAPIURL               string
	RatingReminderInterval   time.Duration
	AutoReleaseInterval      time.Duration
	DisputeEscalationInterval time.Duration
	RatingDeadlineDays       int
	MinRatingForAutoRelease  float64
	DisputeEscalationHours   int
	MaxRetries               int
	RetryDelay               time.Duration
}

var jobsConfig *JobsConfig

// InitJobsConfig initializes the jobs service configuration
func InitJobsConfig() {
	// Initialize base config first
	InitConfig()
	
	jobsConfig = &JobsConfig{
		Port:                      getEnv("JOBS_PORT", "8081"),
		MainAPIURL:                getEnv("MAIN_API_URL", "http://localhost:8080"),
		RatingReminderInterval:    getDurationEnv("RATING_REMINDER_INTERVAL", 24*time.Hour),
		AutoReleaseInterval:       getDurationEnv("AUTO_RELEASE_INTERVAL", 1*time.Hour),
		DisputeEscalationInterval: getDurationEnv("DISPUTE_ESCALATION_INTERVAL", 24*time.Hour),
		RatingDeadlineDays:        getIntEnv("RATING_DEADLINE_DAYS", 7),
		MinRatingForAutoRelease:   getFloatEnv("MIN_RATING_FOR_AUTO_RELEASE", 3.0),
		DisputeEscalationHours:    getIntEnv("DISPUTE_ESCALATION_HOURS", 72),
		MaxRetries:                getIntEnv("MAX_RETRIES", 3),
		RetryDelay:                getDurationEnv("RETRY_DELAY", 30*time.Second),
	}

	log.Printf("🔧 Jobs Service Config: Port=%s, MainAPI=%s", jobsConfig.Port, jobsConfig.MainAPIURL)
}

// GetJobsConfig returns the jobs configuration
func GetJobsConfig() *JobsConfig {
	if jobsConfig == nil {
		InitJobsConfig()
	}
	return jobsConfig
}

// Helper functions
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getIntEnv(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}

func getFloatEnv(key string, defaultValue float64) float64 {
	if value := os.Getenv(key); value != "" {
		if floatValue, err := strconv.ParseFloat(value, 64); err == nil {
			return floatValue
		}
	}
	return defaultValue
}

func getDurationEnv(key string, defaultValue time.Duration) time.Duration {
	if value := os.Getenv(key); value != "" {
		if duration, err := time.ParseDuration(value); err == nil {
			return duration
		}
	}
	return defaultValue
}