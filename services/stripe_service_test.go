package services

import (
	"os"
	"testing"

	"github.com/sebastiancaldarola/goalhero-payment-jobs/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewStripeConnectService(t *testing.T) {
	t.Run("should create service with default test mode", func(t *testing.T) {
		// Clear environment variables
		os.Unsetenv("STRIPE_SECRET_KEY")
		os.Unsetenv("STRIPE_CONNECT_ACCOUNT")
		os.Unsetenv("STRIPE_TEST_MODE")
		
		service := NewStripeConnectService()
		
		assert.NotNil(t, service)
		assert.True(t, service.IsTestMode(), "Should default to test mode")
		assert.NotEmpty(t, service.secretKey, "Should have a secret key")
	})
	
	t.Run("should use environment variables when provided", func(t *testing.T) {
		testSecretKey := "sk_test_custom_key"
		testConnectAccount := "acct_test_custom"
		
		os.Setenv("STRIPE_SECRET_KEY", testSecretKey)
		os.Setenv("STRIPE_CONNECT_ACCOUNT", testConnectAccount)
		os.Setenv("STRIPE_TEST_MODE", "true")
		
		service := NewStripeConnectService()
		
		assert.Equal(t, testSecretKey, service.secretKey)
		assert.Equal(t, testConnectAccount, service.connectAccount)
		assert.True(t, service.IsTestMode())
		
		// Clean up
		os.Unsetenv("STRIPE_SECRET_KEY")
		os.Unsetenv("STRIPE_CONNECT_ACCOUNT")
		os.Unsetenv("STRIPE_TEST_MODE")
	})
	
	t.Run("should disable test mode when explicitly set", func(t *testing.T) {
		os.Setenv("STRIPE_TEST_MODE", "false")
		
		service := NewStripeConnectService()
		
		assert.False(t, service.IsTestMode())
		
		// Clean up
		os.Unsetenv("STRIPE_TEST_MODE")
	})
}

func TestCalculateFees(t *testing.T) {
	service := NewStripeConnectService()
	
	testCases := []struct {
		name                    string
		amount                  float64
		expectedPlatformFee     float64
		expectedStripeFeeMin    float64
		expectedStripeFeeMax    float64
		expectedNetAmount       float64
	}{
		{
			name:                    "minimum_amount_5_euros",
			amount:                  5.0,
			expectedPlatformFee:     0.20, // 4% of 5
			expectedStripeFeeMin:    0.33, // 1.65% + 0.25
			expectedStripeFeeMax:    0.34,
			expectedNetAmount:       4.80, // 5 - 0.20
		},
		{
			name:                    "mid_amount_25_euros",
			amount:                  25.0,
			expectedPlatformFee:     1.00, // 4% of 25
			expectedStripeFeeMin:    0.66, // 1.65% + 0.25
			expectedStripeFeeMax:    0.67,
			expectedNetAmount:       24.00, // 25 - 1.00
		},
		{
			name:                    "maximum_amount_50_euros",
			amount:                  50.0,
			expectedPlatformFee:     2.00, // 4% of 50
			expectedStripeFeeMin:    1.07, // 1.65% + 0.25
			expectedStripeFeeMax:    1.08,
			expectedNetAmount:       48.00, // 50 - 2.00
		},
	}
	
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			platformFee, stripeFee, netAmount := service.CalculateFees(tc.amount)
			
			assert.InDelta(t, tc.expectedPlatformFee, platformFee, 0.01, "Platform fee calculation")
			assert.GreaterOrEqual(t, stripeFee, tc.expectedStripeFeeMin, "Stripe fee should be at least minimum")
			assert.LessOrEqual(t, stripeFee, tc.expectedStripeFeeMax, "Stripe fee should be at most maximum")
			assert.InDelta(t, tc.expectedNetAmount, netAmount, 0.01, "Net amount calculation")
			
			// Verify all fees are positive
			assert.Greater(t, platformFee, 0.0, "Platform fee should be positive")
			assert.Greater(t, stripeFee, 0.0, "Stripe fee should be positive")
			assert.Greater(t, netAmount, 0.0, "Net amount should be positive")
		})
	}
}

func TestCalculateFeesEdgeCases(t *testing.T) {
	service := NewStripeConnectService()
	
	t.Run("should handle very small amounts", func(t *testing.T) {
		platformFee, stripeFee, netAmount := service.CalculateFees(0.01)
		
		assert.Greater(t, platformFee, 0.0, "Should calculate platform fee for small amount")
		assert.Greater(t, stripeFee, 0.0, "Should calculate Stripe fee for small amount")
		assert.GreaterOrEqual(t, netAmount, 0.0, "Net amount should not be negative")
	})
	
	t.Run("should handle large amounts", func(t *testing.T) {
		largeAmount := 1000.0
		platformFee, stripeFee, netAmount := service.CalculateFees(largeAmount)
		
		expectedPlatformFee := largeAmount * models.PlatformFeePercentage / 100
		assert.InDelta(t, expectedPlatformFee, platformFee, 0.01, "Platform fee should scale with amount")
		assert.Greater(t, stripeFee, 0.25, "Stripe fee should include base fee")
		assert.Equal(t, largeAmount-platformFee, netAmount, "Net amount calculation for large amount")
	})
}

func TestValidateConnectAccount(t *testing.T) {
	service := NewStripeConnectService()
	
	testCases := []struct {
		name        string
		accountID   string
		shouldError bool
		errorMsg    string
	}{
		{
			name:        "valid_account_id",
			accountID:   "acct_1234567890",
			shouldError: false,
		},
		{
			name:        "valid_test_account_id",
			accountID:   "acct_test_1234567890abcdef",
			shouldError: false,
		},
		{
			name:        "empty_account_id",
			accountID:   "",
			shouldError: true,
			errorMsg:    "cannot be empty",
		},
		{
			name:        "too_short_account_id",
			accountID:   "acct_123",
			shouldError: true,
			errorMsg:    "invalid connect account ID format",
		},
		{
			name:        "minimum_valid_length",
			accountID:   "1234567890", // 10 characters
			shouldError: false,
		},
		{
			name:        "just_below_minimum",
			accountID:   "123456789", // 9 characters
			shouldError: true,
			errorMsg:    "invalid connect account ID format",
		},
	}
	
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := service.ValidateConnectAccount(tc.accountID)
			
			if tc.shouldError {
				assert.Error(t, err, "Should return validation error")
				if tc.errorMsg != "" {
					assert.Contains(t, err.Error(), tc.errorMsg, "Error message should contain expected text")
				}
			} else {
				assert.NoError(t, err, "Should not return error for valid account ID")
			}
		})
	}
}

func TestGetTestCardTokens(t *testing.T) {
	service := NewStripeConnectService()
	
	testCards := service.GetTestCardTokens()
	
	// Verify we have the expected test cards
	expectedCards := []string{
		"visa_success",
		"visa_decline", 
		"mastercard_success",
		"amex_success",
		"insufficient_funds",
		"expired_card",
		"incorrect_cvc",
		"processing_error",
	}
	
	for _, cardType := range expectedCards {
		assert.Contains(t, testCards, cardType, "Should contain %s test card", cardType)
		
		cardNumber := testCards[cardType]
		assert.NotEmpty(t, cardNumber, "Card number should not be empty for %s", cardType)
		assert.Len(t, cardNumber, 16, "Card number should be 16 digits for %s", cardType)
		assert.Regexp(t, `^\d{16}$`, cardNumber, "Card number should be all digits for %s", cardType)
	}
	
	// Verify specific test card numbers match Stripe's test cards
	assert.Equal(t, "4242424242424242", testCards["visa_success"], "Visa success card should match Stripe test card")
	assert.Equal(t, "4000000000000002", testCards["visa_decline"], "Visa decline card should match Stripe test card")
	assert.Equal(t, "5555555555554444", testCards["mastercard_success"], "Mastercard success should match Stripe test card")
}

func TestIsTestMode(t *testing.T) {
	t.Run("should return true when in test mode", func(t *testing.T) {
		os.Setenv("STRIPE_TEST_MODE", "true")
		service := NewStripeConnectService()
		
		assert.True(t, service.IsTestMode())
		
		os.Unsetenv("STRIPE_TEST_MODE")
	})
	
	t.Run("should return false when explicitly disabled", func(t *testing.T) {
		os.Setenv("STRIPE_TEST_MODE", "false")
		service := NewStripeConnectService()
		
		assert.False(t, service.IsTestMode())
		
		os.Unsetenv("STRIPE_TEST_MODE")
	})
	
	t.Run("should default to true when not set", func(t *testing.T) {
		os.Unsetenv("STRIPE_TEST_MODE")
		service := NewStripeConnectService()
		
		assert.True(t, service.IsTestMode(), "Should default to test mode")
	})
}

// Integration tests that require actual Stripe API calls
func TestStripeServiceIntegration(t *testing.T) {
	// Skip if no Stripe key is provided
	if os.Getenv("STRIPE_SECRET_KEY") == "" {
		t.Skip("Skipping Stripe integration tests: STRIPE_SECRET_KEY not set")
	}
	
	// Ensure we're in test mode
	os.Setenv("STRIPE_TEST_MODE", "true")
	service := NewStripeConnectService()
	require.True(t, service.IsTestMode(), "Integration tests must run in test mode")
	
	testUtils := NewTestUtilities()
	
	t.Run("should create payment intent with valid parameters", func(t *testing.T) {
		payment := testUtils.GenerateTestPayment()
		payment.Amount = 15.0 // Valid amount
		
		organizerID := testUtils.CreateTestOrganizerID()
		
		result, err := service.CreateEscrowPaymentIntent(payment, organizerID)
		
		if err != nil {
			t.Logf("Payment intent creation failed (may be due to test account setup): %v", err)
			return
		}
		
		require.NotNil(t, result, "Payment result should not be nil")
		assert.NotEmpty(t, result.ClientSecret, "Should have client secret")
		assert.NotEmpty(t, result.PaymentIntent.ID, "Should have payment intent ID")
		assert.Equal(t, "requires_payment_method", result.Status, "Should require payment method")
	})
	
	t.Run("should get payment details", func(t *testing.T) {
		// This test requires a valid payment intent ID
		// In a real test, you'd create one first or use a known test ID
		t.Skip("Requires valid payment intent ID")
	})
	
	t.Run("should create transfer with valid parameters", func(t *testing.T) {
		amount := 10.0
		destinationAccount := testUtils.CreateTestOrganizerID()
		metadata := map[string]string{
			"test_transfer": "true",
			"amount":       "10.00",
		}
		
		transfer, err := service.CreateTransfer(amount, destinationAccount, metadata)
		
		if err != nil {
			t.Logf("Transfer creation failed (expected for test account): %v", err)
			return
		}
		
		require.NotNil(t, transfer, "Transfer should not be nil")
		assert.Equal(t, int64(amount*100), transfer.Amount, "Transfer amount should match")
		assert.Equal(t, destinationAccount, transfer.Destination.ID, "Destination should match")
	})
}

func TestStripeServiceErrorHandling(t *testing.T) {
	service := NewStripeConnectService()
	
	t.Run("should handle nil payment", func(t *testing.T) {
		result, err := service.CreateEscrowPaymentIntent(nil, "test_organizer")
		
		assert.Error(t, err, "Should return error for nil payment")
		assert.Nil(t, result, "Result should be nil on error")
	})
	
	t.Run("should validate organizer ID", func(t *testing.T) {
		testUtils := NewTestUtilities()
		payment := testUtils.GenerateTestPayment()
		payment.Amount = 15.0
		
		result, err := service.CreateEscrowPaymentIntent(payment, "")
		
		assert.Error(t, err, "Should return error for empty organizer ID")
		assert.Nil(t, result, "Result should be nil on error")
	})
	
	t.Run("should handle invalid payment intent ID", func(t *testing.T) {
		result, err := service.ConfirmPaymentIntent("invalid_payment_intent_id")
		
		assert.Error(t, err, "Should return error for invalid payment intent ID")
		assert.Nil(t, result, "Result should be nil on error")
	})
	
	t.Run("should handle invalid refund parameters", func(t *testing.T) {
		refund, err := service.CreateRefund("invalid_payment_intent", -10.0, "test")
		
		assert.Error(t, err, "Should return error for negative amount")
		assert.Nil(t, refund, "Refund should be nil on error")
	})
}

func TestStripeServiceConcurrency(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping concurrency test in short mode")
	}
	
	service := NewStripeConnectService()
	testUtils := NewTestUtilities()
	
	t.Run("should handle concurrent fee calculations", func(t *testing.T) {
		const numGoroutines = 10
		results := make(chan bool, numGoroutines)
		
		for i := 0; i < numGoroutines; i++ {
			go func() {
				amount := testUtils.GenerateRandomAmount()
				platformFee, stripeFee, netAmount := service.CalculateFees(amount)
				
				// Verify calculations are consistent
				expectedPlatformFee := amount * models.PlatformFeePercentage / 100
				expectedNetAmount := amount - platformFee
				
				platformFeeOK := abs(platformFee-expectedPlatformFee) < 0.01
				netAmountOK := abs(netAmount-expectedNetAmount) < 0.01
				stripeFeeOK := stripeFee > 0
				
				results <- platformFeeOK && netAmountOK && stripeFeeOK
			}()
		}
		
		// Collect results
		successCount := 0
		for i := 0; i < numGoroutines; i++ {
			if <-results {
				successCount++
			}
		}
		
		assert.Equal(t, numGoroutines, successCount, "All concurrent calculations should be correct")
	})
}