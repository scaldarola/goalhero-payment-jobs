package config

import (
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestConfig(t *testing.T) {
	t.Run("should initialize config with default values", func(t *testing.T) {
		// Clear any existing environment variables
		os.Unsetenv("PAYMENT_TEST_MODE")
		os.Unsetenv("STRIPE_TEST_MODE")
		os.Unsetenv("AUTO_ACCEPT_PAYMENTS")
		os.Unsetenv("PORT")
		os.Unsetenv("ENVIRONMENT")

		// Reset config to nil to force re-initialization
		AppConfig = nil

		InitConfig()

		require.NotNil(t, AppConfig)
		assert.True(t, AppConfig.PaymentTestMode)
		assert.True(t, AppConfig.StripeTestMode)
		assert.True(t, AppConfig.AutoAcceptPayments)
		assert.Equal(t, "8080", AppConfig.Port)
		assert.Equal(t, "development", AppConfig.Environment)
	})

	t.Run("should initialize config with environment variables", func(t *testing.T) {
		// Set environment variables
		os.Setenv("PAYMENT_TEST_MODE", "false")
		os.Setenv("STRIPE_TEST_MODE", "false")
		os.Setenv("AUTO_ACCEPT_PAYMENTS", "false")
		os.Setenv("PORT", "9000")
		os.Setenv("ENVIRONMENT", "production")

		// Reset config to force re-initialization
		AppConfig = nil

		InitConfig()

		require.NotNil(t, AppConfig)
		assert.False(t, AppConfig.PaymentTestMode)
		assert.False(t, AppConfig.StripeTestMode)
		assert.False(t, AppConfig.AutoAcceptPayments)
		assert.Equal(t, "9000", AppConfig.Port)
		assert.Equal(t, "production", AppConfig.Environment)

		// Clean up
		os.Unsetenv("PAYMENT_TEST_MODE")
		os.Unsetenv("STRIPE_TEST_MODE")
		os.Unsetenv("AUTO_ACCEPT_PAYMENTS")
		os.Unsetenv("PORT")
		os.Unsetenv("ENVIRONMENT")
	})
}

func TestJobsConfig(t *testing.T) {
	t.Run("should create jobs config structure", func(t *testing.T) {
		config := &JobsConfig{
			Port:                      "8081",
			MainAPIURL:               "http://localhost:8080",
			RatingReminderInterval:    24 * time.Hour,
			AutoReleaseInterval:       1 * time.Hour,
			DisputeEscalationInterval: 24 * time.Hour,
			RatingDeadlineDays:        7,
			MinRatingForAutoRelease:   3.0,
			DisputeEscalationHours:    72,
			MaxRetries:               3,
			RetryDelay:               30 * time.Second,
		}

		assert.Equal(t, "8081", config.Port)
		assert.Equal(t, "http://localhost:8080", config.MainAPIURL)
		assert.Equal(t, 24*time.Hour, config.RatingReminderInterval)
		assert.Equal(t, 1*time.Hour, config.AutoReleaseInterval)
		assert.Equal(t, 24*time.Hour, config.DisputeEscalationInterval)
		assert.Equal(t, 7, config.RatingDeadlineDays)
		assert.Equal(t, 3.0, config.MinRatingForAutoRelease)
		assert.Equal(t, 72, config.DisputeEscalationHours)
		assert.Equal(t, 3, config.MaxRetries)
		assert.Equal(t, 30*time.Second, config.RetryDelay)
	})

	t.Run("should return existing jobs config when already initialized", func(t *testing.T) {
		// Pre-initialize config
		existingConfig := &JobsConfig{
			Port:       "test-port",
			MainAPIURL: "test-url",
		}
		jobsConfig = existingConfig

		config := GetJobsConfig()
		assert.Equal(t, existingConfig, config)
		assert.Equal(t, "test-port", config.Port)
		assert.Equal(t, "test-url", config.MainAPIURL)
	})
}

func TestHelperFunctions(t *testing.T) {
	t.Run("getEnv should return environment value when set", func(t *testing.T) {
		os.Setenv("TEST_ENV_VAR", "test_value")
		defer os.Unsetenv("TEST_ENV_VAR")

		result := getEnv("TEST_ENV_VAR", "default_value")
		assert.Equal(t, "test_value", result)
	})

	t.Run("getEnv should return default value when env var not set", func(t *testing.T) {
		os.Unsetenv("NON_EXISTENT_VAR")

		result := getEnv("NON_EXISTENT_VAR", "default_value")
		assert.Equal(t, "default_value", result)
	})

	t.Run("getBoolEnv should return environment value when set to true", func(t *testing.T) {
		os.Setenv("TEST_BOOL_VAR", "true")
		defer os.Unsetenv("TEST_BOOL_VAR")

		result := getBoolEnv("TEST_BOOL_VAR", false)
		assert.True(t, result)
	})

	t.Run("getBoolEnv should return environment value when set to false", func(t *testing.T) {
		os.Setenv("TEST_BOOL_VAR", "false")
		defer os.Unsetenv("TEST_BOOL_VAR")

		result := getBoolEnv("TEST_BOOL_VAR", true)
		assert.False(t, result)
	})

	t.Run("getBoolEnv should return default value when env var not set", func(t *testing.T) {
		os.Unsetenv("NON_EXISTENT_BOOL_VAR")

		result := getBoolEnv("NON_EXISTENT_BOOL_VAR", true)
		assert.True(t, result)
	})

	t.Run("getBoolEnv should return default value when env var has invalid format", func(t *testing.T) {
		os.Setenv("INVALID_BOOL_VAR", "not_a_boolean")
		defer os.Unsetenv("INVALID_BOOL_VAR")

		result := getBoolEnv("INVALID_BOOL_VAR", false)
		assert.False(t, result)
	})

	t.Run("getIntEnv should return environment value when set", func(t *testing.T) {
		os.Setenv("TEST_INT_VAR", "42")
		defer os.Unsetenv("TEST_INT_VAR")

		result := getIntEnv("TEST_INT_VAR", 0)
		assert.Equal(t, 42, result)
	})

	t.Run("getIntEnv should return default value when env var not set", func(t *testing.T) {
		os.Unsetenv("NON_EXISTENT_INT_VAR")

		result := getIntEnv("NON_EXISTENT_INT_VAR", 100)
		assert.Equal(t, 100, result)
	})

	t.Run("getIntEnv should return default value when env var has invalid format", func(t *testing.T) {
		os.Setenv("INVALID_INT_VAR", "not_a_number")
		defer os.Unsetenv("INVALID_INT_VAR")

		result := getIntEnv("INVALID_INT_VAR", 50)
		assert.Equal(t, 50, result)
	})

	t.Run("getFloatEnv should return environment value when set", func(t *testing.T) {
		os.Setenv("TEST_FLOAT_VAR", "3.14")
		defer os.Unsetenv("TEST_FLOAT_VAR")

		result := getFloatEnv("TEST_FLOAT_VAR", 0.0)
		assert.Equal(t, 3.14, result)
	})

	t.Run("getFloatEnv should return default value when env var not set", func(t *testing.T) {
		os.Unsetenv("NON_EXISTENT_FLOAT_VAR")

		result := getFloatEnv("NON_EXISTENT_FLOAT_VAR", 2.5)
		assert.Equal(t, 2.5, result)
	})

	t.Run("getFloatEnv should return default value when env var has invalid format", func(t *testing.T) {
		os.Setenv("INVALID_FLOAT_VAR", "not_a_float")
		defer os.Unsetenv("INVALID_FLOAT_VAR")

		result := getFloatEnv("INVALID_FLOAT_VAR", 1.5)
		assert.Equal(t, 1.5, result)
	})

	t.Run("getDurationEnv should return environment value when set", func(t *testing.T) {
		os.Setenv("TEST_DURATION_VAR", "30s")
		defer os.Unsetenv("TEST_DURATION_VAR")

		result := getDurationEnv("TEST_DURATION_VAR", 0)
		assert.Equal(t, 30*time.Second, result)
	})

	t.Run("getDurationEnv should return default value when env var not set", func(t *testing.T) {
		os.Unsetenv("NON_EXISTENT_DURATION_VAR")

		result := getDurationEnv("NON_EXISTENT_DURATION_VAR", 1*time.Minute)
		assert.Equal(t, 1*time.Minute, result)
	})

	t.Run("getDurationEnv should return default value when env var has invalid format", func(t *testing.T) {
		os.Setenv("INVALID_DURATION_VAR", "not_a_duration")
		defer os.Unsetenv("INVALID_DURATION_VAR")

		result := getDurationEnv("INVALID_DURATION_VAR", 5*time.Second)
		assert.Equal(t, 5*time.Second, result)
	})
}

func TestConfigUtilities(t *testing.T) {
	t.Run("IsTestMode should return true when PaymentTestMode is true", func(t *testing.T) {
		AppConfig = &Config{
			PaymentTestMode: true,
			Environment:     "production",
		}

		assert.True(t, IsTestMode())
	})

	t.Run("IsTestMode should return true when Environment is development", func(t *testing.T) {
		AppConfig = &Config{
			PaymentTestMode: false,
			Environment:     "development",
		}

		assert.True(t, IsTestMode())
	})

	t.Run("IsTestMode should return false when neither condition is met", func(t *testing.T) {
		AppConfig = &Config{
			PaymentTestMode: false,
			Environment:     "production",
		}

		assert.False(t, IsTestMode())
	})

	t.Run("IsAutoAcceptEnabled should return true when conditions are met", func(t *testing.T) {
		AppConfig = &Config{
			PaymentTestMode:    true,
			AutoAcceptPayments: true,
			Environment:        "development",
		}

		assert.True(t, IsAutoAcceptEnabled())
	})

	t.Run("IsAutoAcceptEnabled should return false when AutoAcceptPayments is false", func(t *testing.T) {
		AppConfig = &Config{
			PaymentTestMode:    true,
			AutoAcceptPayments: false,
			Environment:        "development",
		}

		assert.False(t, IsAutoAcceptEnabled())
	})

	t.Run("IsAutoAcceptEnabled should return false when not in test mode", func(t *testing.T) {
		AppConfig = &Config{
			PaymentTestMode:    false,
			AutoAcceptPayments: true,
			Environment:        "production",
		}

		assert.False(t, IsAutoAcceptEnabled())
	})
}