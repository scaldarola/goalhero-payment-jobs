package config

import (
	"context"
	"log"
	"os"
	"strconv"
	"time"

	"cloud.google.com/go/firestore"
	"firebase.google.com/go/v4"
	"google.golang.org/api/option"
)

type Config struct {
	PaymentTestMode    bool
	StripeTestMode     bool
	AutoAcceptPayments bool
	Port               string
	Environment        string
}

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

var (
	AppConfig *Config
	jobsConfig *JobsConfig
	firestoreClient *firestore.Client
)

func InitConfig() {
	AppConfig = &Config{
		PaymentTestMode:    getBoolEnv("PAYMENT_TEST_MODE", true),
		StripeTestMode:     getBoolEnv("STRIPE_TEST_MODE", true),
		AutoAcceptPayments: getBoolEnv("AUTO_ACCEPT_PAYMENTS", true),
		Port:              getEnv("PORT", "8080"),
		Environment:       getEnv("ENVIRONMENT", "development"),
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getBoolEnv(key string, defaultValue bool) bool {
	if value := os.Getenv(key); value != "" {
		if parsed, err := strconv.ParseBool(value); err == nil {
			return parsed
		}
	}
	return defaultValue
}

func IsTestMode() bool {
	return AppConfig.PaymentTestMode || AppConfig.Environment == "development"
}

func IsAutoAcceptEnabled() bool {
	return AppConfig.AutoAcceptPayments && IsTestMode()
}

// InitJobsConfig initializes the jobs service configuration
func InitJobsConfig() {
	// Initialize base config first
	InitConfig()
	
	// Initialize Firestore
	InitFirestore()
	
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

// InitFirestore initializes Firestore client
func InitFirestore() {
	ctx := context.Background()
	
	var opt option.ClientOption
	
	// Check for environment variable first (for Vercel/production)
	if credentialsJSON := os.Getenv("GOOGLE_APPLICATION_CREDENTIALS"); credentialsJSON != "" {
		opt = option.WithCredentialsJSON([]byte(credentialsJSON))
	} else {
		// Fallback to file for local development
		credentialsPath := "auth/firebase_credentials.json"
		if _, err := os.Stat(credentialsPath); os.IsNotExist(err) {
			log.Printf("⚠️ No Firebase credentials found, Firestore will be disabled")
			return
		}
		opt = option.WithCredentialsFile(credentialsPath)
	}
	
	app, err := firebase.NewApp(ctx, nil, opt)
	if err != nil {
		log.Printf("⚠️ Error initializing Firebase app: %v", err)
		return
	}
	
	client, err := app.Firestore(ctx)
	if err != nil {
		log.Printf("⚠️ Error initializing Firestore: %v", err)
		return
	}
	
	firestoreClient = client
	SetFirestoreClient(client)
	log.Println("✅ Firestore initialized")
}

// Helper functions for jobs config
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

