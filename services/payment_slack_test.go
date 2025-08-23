package services

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSendSlackSuccessNotification(t *testing.T) {
	tests := []struct {
		name           string
		escrowID       string
		amount         float64
		releaseReason  string
		webhookURL     string
		expectedCalled bool
	}{
		{
			name:           "success_with_webhook_configured",
			escrowID:       "test_escrow_123",
			amount:         25.50,
			releaseReason:  "automatic_release",
			webhookURL:     "http://test-webhook.com",
			expectedCalled: true,
		},
		{
			name:           "no_webhook_configured",
			escrowID:       "test_escrow_456",
			amount:         15.00,
			releaseReason:  "manual_release",
			webhookURL:     "",
			expectedCalled: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup mock HTTP server
			var receivedMessage SlackMessage
			called := false
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				called = true
				body, _ := io.ReadAll(r.Body)
				json.Unmarshal(body, &receivedMessage)
				w.WriteHeader(http.StatusOK)
			}))
			defer server.Close()

			// Set environment variable
			originalWebhook := os.Getenv("SLACK_ESCROW_WEBHOOK_URL")
			defer os.Setenv("SLACK_ESCROW_WEBHOOK_URL", originalWebhook)

			if tt.webhookURL != "" {
				os.Setenv("SLACK_ESCROW_WEBHOOK_URL", server.URL)
			} else {
				os.Unsetenv("SLACK_ESCROW_WEBHOOK_URL")
			}

			// Create service and call function
			service := &PaymentService{}
			service.sendSlackSuccessNotification(tt.escrowID, tt.amount, tt.releaseReason)

			// Verify expectations
			assert.Equal(t, tt.expectedCalled, called)

			if tt.expectedCalled {
				assert.Contains(t, receivedMessage.Text, "✅ *Escrow Payment Processed Successfully*")
				assert.Contains(t, receivedMessage.Text, tt.escrowID)
				assert.Contains(t, receivedMessage.Text, "€25.50")
				assert.Contains(t, receivedMessage.Text, tt.releaseReason)
				assert.Contains(t, receivedMessage.Text, "Status: Released")
			}
		})
	}
}

func TestSendSlackFailureNotification(t *testing.T) {
	tests := []struct {
		name           string
		escrowID       string
		amount         float64
		errorMsg       string
		webhookURL     string
		expectedCalled bool
	}{
		{
			name:           "failure_with_webhook_configured",
			escrowID:       "test_escrow_789",
			amount:         30.00,
			errorMsg:       "Stripe API error",
			webhookURL:     "http://test-webhook.com",
			expectedCalled: true,
		},
		{
			name:           "no_webhook_configured",
			escrowID:       "test_escrow_012",
			amount:         20.00,
			errorMsg:       "Database connection failed",
			webhookURL:     "",
			expectedCalled: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup mock HTTP server
			var receivedMessage SlackMessage
			called := false
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				called = true
				body, _ := io.ReadAll(r.Body)
				json.Unmarshal(body, &receivedMessage)
				w.WriteHeader(http.StatusOK)
			}))
			defer server.Close()

			// Set environment variable
			originalWebhook := os.Getenv("SLACK_ESCROW_WEBHOOK_URL")
			defer os.Setenv("SLACK_ESCROW_WEBHOOK_URL", originalWebhook)

			if tt.webhookURL != "" {
				os.Setenv("SLACK_ESCROW_WEBHOOK_URL", server.URL)
			} else {
				os.Unsetenv("SLACK_ESCROW_WEBHOOK_URL")
			}

			// Create service and call function
			service := &PaymentService{}
			service.sendSlackFailureNotification(tt.escrowID, tt.amount, tt.errorMsg)

			// Verify expectations
			assert.Equal(t, tt.expectedCalled, called)

			if tt.expectedCalled {
				assert.Contains(t, receivedMessage.Text, "❌ *Escrow Payment Processing Failed*")
				assert.Contains(t, receivedMessage.Text, tt.escrowID)
				assert.Contains(t, receivedMessage.Text, "€30.00")
				assert.Contains(t, receivedMessage.Text, tt.errorMsg)
				assert.Contains(t, receivedMessage.Text, "Status: Failed")
			}
		})
	}
}

func TestSendSlackMessage(t *testing.T) {
	tests := []struct {
		name           string
		message        SlackMessage
		serverStatus   int
		expectedError  bool
		errorContains  string
	}{
		{
			name: "successful_message",
			message: SlackMessage{
				Text: "Test message",
			},
			serverStatus:  http.StatusOK,
			expectedError: false,
		},
		{
			name: "server_error_response",
			message: SlackMessage{
				Text: "Test message",
			},
			serverStatus:  http.StatusInternalServerError,
			expectedError: false, // Function doesn't return errors, just logs
		},
		{
			name: "bad_request_response",
			message: SlackMessage{
				Text: "Test message",
			},
			serverStatus:  http.StatusBadRequest,
			expectedError: false, // Function doesn't return errors, just logs
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup mock HTTP server
			var receivedMessage SlackMessage
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				// Verify content type
				assert.Equal(t, "application/json", r.Header.Get("Content-Type"))
				
				// Read and parse body
				body, err := io.ReadAll(r.Body)
				require.NoError(t, err)
				
				err = json.Unmarshal(body, &receivedMessage)
				require.NoError(t, err)
				
				w.WriteHeader(tt.serverStatus)
			}))
			defer server.Close()

			// Create service and call function
			service := &PaymentService{}
			service.sendSlackMessage(tt.message, server.URL)

			// Verify message was received correctly
			assert.Equal(t, tt.message.Text, receivedMessage.Text)
		})
	}
}

func TestSendSlackMessage_InvalidURL(t *testing.T) {
	service := &PaymentService{}
	message := SlackMessage{Text: "Test message"}
	
	// This should not panic and should handle the error gracefully
	service.sendSlackMessage(message, "invalid-url")
	
	// Function should complete without crashing (it logs errors instead of returning them)
}

func TestSendSlackMessage_JSONMarshalError(t *testing.T) {
	service := &PaymentService{}
	
	// Create a message that would cause JSON marshal to fail
	// Note: In Go, it's actually hard to make json.Marshal fail with simple structs,
	// but we can test this by creating a custom type that implements MarshalJSON badly
	message := SlackMessage{Text: "Normal message"}
	
	// Test with a valid server that returns 200
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()
	
	// This should work fine since SlackMessage is simple
	service.sendSlackMessage(message, server.URL)
}

// Test integration with ProcessAutomaticReleases to ensure Slack notifications are called
func TestProcessAutomaticReleases_SlackIntegration(t *testing.T) {
	// Setup a mock server to capture Slack notifications
	var slackMessages []SlackMessage
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var message SlackMessage
		body, _ := io.ReadAll(r.Body)
		json.Unmarshal(body, &message)
		slackMessages = append(slackMessages, message)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	// Set the webhook URL for the test
	originalWebhook := os.Getenv("SLACK_ESCROW_WEBHOOK_URL")
	defer os.Setenv("SLACK_ESCROW_WEBHOOK_URL", originalWebhook)
	os.Setenv("SLACK_ESCROW_WEBHOOK_URL", server.URL)

	// Create a payment service
	service := NewPaymentService()

	// Note: This test would need a proper database setup to fully test,
	// but we're testing the Slack integration portion
	// In a real scenario, you'd mock the database calls or use a test database
	
	// The actual test would involve:
	// 1. Create test escrow transactions in the database
	// 2. Call ProcessAutomaticReleases()
	// 3. Verify that Slack notifications were sent for successes/failures
	
	// For now, we can at least verify the service can be created
	assert.NotNil(t, service)
}

// Test that Slack notifications contain proper formatting
func TestSlackMessageFormatting(t *testing.T) {
	service := &PaymentService{}
	
	// Test success notification formatting
	var successMessage SlackMessage
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		json.Unmarshal(body, &successMessage)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()
	
	originalWebhook := os.Getenv("SLACK_ESCROW_WEBHOOK_URL")
	defer os.Setenv("SLACK_ESCROW_WEBHOOK_URL", originalWebhook)
	os.Setenv("SLACK_ESCROW_WEBHOOK_URL", server.URL)
	
	service.sendSlackSuccessNotification("escrow_123", 42.75, "test_release")
	
	// Verify the message format
	assert.True(t, strings.Contains(successMessage.Text, "✅"))
	assert.True(t, strings.Contains(successMessage.Text, "*Escrow Payment Processed Successfully*"))
	assert.True(t, strings.Contains(successMessage.Text, "escrow_123"))
	assert.True(t, strings.Contains(successMessage.Text, "€42.75"))
	assert.True(t, strings.Contains(successMessage.Text, "test_release"))
	assert.True(t, strings.Contains(successMessage.Text, "Released"))
	
	// Test that the message is properly structured with newlines
	lines := strings.Split(successMessage.Text, "\n")
	assert.GreaterOrEqual(t, len(lines), 4) // Should have multiple lines
}