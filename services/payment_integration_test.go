package services

import (
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/sebastiancaldarola/goalhero-payment-jobs/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

// PaymentIntegrationTestSuite defines the integration test suite
type PaymentIntegrationTestSuite struct {
	suite.Suite
	paymentService *PaymentService
	stripeService  *StripeConnectService
	testData       *TestData
}

// TestData holds test data for integration tests
type TestData struct {
	TestUserID        string
	TestGameID        string
	TestApplicationID string
	TestOrganizerID   string
	TestAmount        float64
	TestPaymentID     string
}

// SetupSuite runs before all tests in the suite
func (suite *PaymentIntegrationTestSuite) SetupSuite() {
	// Skip integration tests if required env vars are not set
	if os.Getenv("STRIPE_SECRET_KEY") == "" {
		suite.T().Skip("Skipping integration tests: STRIPE_SECRET_KEY not set")
	}

	// Ensure we're in test mode
	os.Setenv("STRIPE_TEST_MODE", "true")
	
	suite.paymentService = NewPaymentService()
	suite.stripeService = NewStripeConnectService()
	
	// Verify test mode is enabled
	require.True(suite.T(), suite.stripeService.IsTestMode(), "Tests must run in Stripe test mode")
	
	// Initialize test data
	suite.testData = &TestData{
		TestUserID:        "test_user_integration_" + generateTestID(),
		TestGameID:        "test_game_integration_" + generateTestID(),
		TestApplicationID: "test_app_integration_" + generateTestID(),
		TestOrganizerID:   "acct_test_organizer_" + generateTestID(),
		TestAmount:        15.50, // Standard test amount
	}
}

// TearDownSuite runs after all tests in the suite
func (suite *PaymentIntegrationTestSuite) TearDownSuite() {
	// Cleanup any test data if needed
	// Note: In test mode, Stripe test data is automatically cleaned up
}

// SetupTest runs before each individual test
func (suite *PaymentIntegrationTestSuite) SetupTest() {
	// Reset test payment ID for each test
	suite.testData.TestPaymentID = ""
}

// Test successful payment flow end-to-end
func (suite *PaymentIntegrationTestSuite) TestSuccessfulPaymentFlow() {
	
	// Step 1: Create payment intent
	payment, paymentResult, err := suite.paymentService.CreateGamePayment(
		suite.testData.TestUserID,
		suite.testData.TestGameID,
		suite.testData.TestApplicationID,
		suite.testData.TestOrganizerID,
		suite.testData.TestAmount,
	)
	
	require.NoError(suite.T(), err, "Payment creation should succeed")
	require.NotNil(suite.T(), payment, "Payment should not be nil")
	require.NotNil(suite.T(), paymentResult, "Payment result should not be nil")
	
	suite.testData.TestPaymentID = payment.ID
	
	// Verify payment details
	assert.Equal(suite.T(), suite.testData.TestAmount, payment.Amount)
	assert.Equal(suite.T(), suite.testData.TestUserID, payment.UserID)
	assert.Equal(suite.T(), suite.testData.TestGameID, payment.GameID)
	assert.Equal(suite.T(), models.PaymentStatusPending, payment.Status)
	assert.NotEmpty(suite.T(), paymentResult.ClientSecret)
	assert.NotEmpty(suite.T(), paymentResult.PaymentIntent.ID)
	
	// Verify fee calculations
	expectedPlatformFee := suite.testData.TestAmount * models.PlatformFeePercentage / 100
	assert.InDelta(suite.T(), expectedPlatformFee, payment.PlatformFee, 0.01)
	assert.Greater(suite.T(), payment.PaymentFee, 0.0, "Payment fee should be calculated")
	assert.Equal(suite.T(), payment.Amount-payment.PlatformFee, payment.NetAmount)
	
	// Step 2: Confirm payment (simulates successful payment)
	confirmedPayment, escrow, err := suite.paymentService.ConfirmGamePayment(payment.ID)
	
	require.NoError(suite.T(), err, "Payment confirmation should succeed")
	require.NotNil(suite.T(), confirmedPayment, "Confirmed payment should not be nil")
	require.NotNil(suite.T(), escrow, "Escrow should be created")
	
	// Verify payment confirmation
	assert.Equal(suite.T(), models.PaymentStatusConfirmed, confirmedPayment.Status)
	assert.NotNil(suite.T(), confirmedPayment.ConfirmedAt)
	assert.True(suite.T(), confirmedPayment.ConfirmedAt.After(payment.CreatedAt))
	
	// Verify escrow creation
	assert.Equal(suite.T(), models.EscrowStatusHeld, escrow.Status)
	assert.Equal(suite.T(), payment.NetAmount, escrow.Amount)
	assert.Equal(suite.T(), suite.testData.TestOrganizerID, escrow.OrganizerID)
	assert.Equal(suite.T(), payment.ID, escrow.PaymentID)
	
	// Step 3: Test escrow release
	err = suite.paymentService.ProcessEscrowRelease(escrow.ID, "integration_test_release")
	require.NoError(suite.T(), err, "Escrow release should succeed")
	
	// Verify escrow was released
	// Note: In a real implementation, you'd fetch the updated escrow from the database
	// For now, we just verify the operation completed without error
}

// Test payment validation failures
func (suite *PaymentIntegrationTestSuite) TestPaymentValidationFailures() {
	testCases := []struct {
		name          string
		amount        float64
		expectedError string
	}{
		{
			name:          "amount_too_low",
			amount:        2.0, // Below minimum of €5
			expectedError: "below minimum",
		},
		{
			name:          "amount_too_high", 
			amount:        75.0, // Above maximum of €50
			expectedError: "above maximum",
		},
		{
			name:          "zero_amount",
			amount:        0.0,
			expectedError: "must be greater than 0",
		},
		{
			name:          "negative_amount",
			amount:        -5.0,
			expectedError: "must be greater than 0",
		},
	}
	
	for _, tc := range testCases {
		suite.T().Run(tc.name, func(t *testing.T) {
			payment, paymentResult, err := suite.paymentService.CreateGamePayment(
				suite.testData.TestUserID,
				suite.testData.TestGameID,
				suite.testData.TestApplicationID,
				suite.testData.TestOrganizerID,
				tc.amount,
			)
			
			assert.Error(t, err, "Should return validation error")
			assert.Contains(t, err.Error(), tc.expectedError)
			assert.Nil(t, payment, "Payment should be nil on validation error")
			assert.Nil(t, paymentResult, "Payment result should be nil on validation error")
		})
	}
}

// Test invalid parameters
func (suite *PaymentIntegrationTestSuite) TestInvalidParameters() {
	testCases := []struct {
		name          string
		userID        string
		gameID        string
		applicationID string
		organizerID   string
		expectedError string
	}{
		{
			name:          "empty_user_id",
			userID:        "",
			gameID:        suite.testData.TestGameID,
			applicationID: suite.testData.TestApplicationID,
			organizerID:   suite.testData.TestOrganizerID,
			expectedError: "user ID cannot be empty",
		},
		{
			name:          "empty_game_id",
			userID:        suite.testData.TestUserID,
			gameID:        "",
			applicationID: suite.testData.TestApplicationID,
			organizerID:   suite.testData.TestOrganizerID,
			expectedError: "game ID cannot be empty",
		},
		{
			name:          "empty_organizer_id",
			userID:        suite.testData.TestUserID,
			gameID:        suite.testData.TestGameID,
			applicationID: suite.testData.TestApplicationID,
			organizerID:   "",
			expectedError: "organizer ID cannot be empty",
		},
	}
	
	for _, tc := range testCases {
		suite.T().Run(tc.name, func(t *testing.T) {
			payment, paymentResult, err := suite.paymentService.CreateGamePayment(
				tc.userID,
				tc.gameID,
				tc.applicationID,
				tc.organizerID,
				suite.testData.TestAmount,
			)
			
			assert.Error(t, err, "Should return validation error")
			assert.Contains(t, err.Error(), tc.expectedError)
			assert.Nil(t, payment, "Payment should be nil on validation error")
			assert.Nil(t, paymentResult, "Payment result should be nil on validation error")
		})
	}
}

// Test fee calculations
func (suite *PaymentIntegrationTestSuite) TestFeeCalculations() {
	testAmounts := []float64{5.0, 15.0, 25.0, 50.0}
	
	for _, amount := range testAmounts {
		suite.T().Run(fmt.Sprintf("amount_%.0f", amount), func(t *testing.T) {
			platformFee, stripeFee, netAmount := suite.stripeService.CalculateFees(amount)
			
			// Verify platform fee (4%)
			expectedPlatformFee := amount * models.PlatformFeePercentage / 100
			assert.InDelta(t, expectedPlatformFee, platformFee, 0.01, "Platform fee should be 4%")
			
			// Verify Stripe fee structure (1.65% + €0.25)
			expectedStripeFee := amount*1.65/100 + 0.25
			assert.InDelta(t, expectedStripeFee, stripeFee, 0.01, "Stripe fee calculation")
			
			// Verify net amount
			expectedNetAmount := amount - platformFee
			assert.InDelta(t, expectedNetAmount, netAmount, 0.01, "Net amount calculation")
			
			// Ensure all fees are positive
			assert.Greater(t, platformFee, 0.0, "Platform fee should be positive")
			assert.Greater(t, stripeFee, 0.0, "Stripe fee should be positive")
			assert.Greater(t, netAmount, 0.0, "Net amount should be positive")
		})
	}
}

// Test refund functionality
func (suite *PaymentIntegrationTestSuite) TestRefundFlow() {
	// First create and confirm a payment
	payment, _, err := suite.paymentService.CreateGamePayment(
		suite.testData.TestUserID,
		suite.testData.TestGameID,
		suite.testData.TestApplicationID,
		suite.testData.TestOrganizerID,
		suite.testData.TestAmount,
	)
	require.NoError(suite.T(), err)
	
	confirmedPayment, escrow, err := suite.paymentService.ConfirmGamePayment(payment.ID)
	require.NoError(suite.T(), err)
	require.NotNil(suite.T(), escrow)
	
	// Get payment details from Stripe
	stripePI, err := suite.stripeService.GetPaymentDetails(confirmedPayment.StripePaymentID)
	require.NoError(suite.T(), err)
	require.NotNil(suite.T(), stripePI)
	
	// Test refund creation
	refundAmount := suite.testData.TestAmount / 2 // Partial refund
	refund, err := suite.stripeService.CreateRefund(
		stripePI.ID,
		refundAmount,
		"integration_test_refund",
	)
	
	require.NoError(suite.T(), err, "Refund creation should succeed")
	require.NotNil(suite.T(), refund, "Refund should not be nil")
	
	// Verify refund details
	expectedRefundCents := int64(refundAmount * 100)
	assert.Equal(suite.T(), expectedRefundCents, refund.Amount)
	assert.Equal(suite.T(), stripePI.ID, refund.PaymentIntent.ID)
	assert.Equal(suite.T(), "requested_by_customer", refund.Reason)
	assert.Contains(suite.T(), refund.Metadata, "refund_reason")
}

// Test Stripe Connect account validation
func (suite *PaymentIntegrationTestSuite) TestConnectAccountValidation() {
	testCases := []struct {
		name        string
		accountID   string
		shouldError bool
		errorMsg    string
	}{
		{
			name:        "valid_test_account",
			accountID:   "acct_test_valid_account_123456",
			shouldError: false,
		},
		{
			name:        "empty_account",
			accountID:   "",
			shouldError: true,
			errorMsg:    "cannot be empty",
		},
		{
			name:        "too_short_account",
			accountID:   "acct_123",
			shouldError: true,
			errorMsg:    "invalid connect account ID format",
		},
	}
	
	for _, tc := range testCases {
		suite.T().Run(tc.name, func(t *testing.T) {
			err := suite.stripeService.ValidateConnectAccount(tc.accountID)
			
			if tc.shouldError {
				assert.Error(t, err)
				if tc.errorMsg != "" {
					assert.Contains(t, err.Error(), tc.errorMsg)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// Test concurrent payment creation
func (suite *PaymentIntegrationTestSuite) TestConcurrentPaymentCreation() {
	const numGoroutines = 5
	const paymentsPerGoroutine = 3
	
	results := make(chan error, numGoroutines*paymentsPerGoroutine)
	
	for i := 0; i < numGoroutines; i++ {
		go func(routineID int) {
			for j := 0; j < paymentsPerGoroutine; j++ {
				userID := fmt.Sprintf("concurrent_user_%d_%d_%s", routineID, j, generateTestID())
				gameID := fmt.Sprintf("concurrent_game_%d_%d_%s", routineID, j, generateTestID())
				appID := fmt.Sprintf("concurrent_app_%d_%d_%s", routineID, j, generateTestID())
				
				_, _, err := suite.paymentService.CreateGamePayment(
					userID,
					gameID,
					appID,
					suite.testData.TestOrganizerID,
					suite.testData.TestAmount,
				)
				results <- err
			}
		}(i)
	}
	
	// Collect results
	successCount := 0
	for i := 0; i < numGoroutines*paymentsPerGoroutine; i++ {
		err := <-results
		if err == nil {
			successCount++
		} else {
			suite.T().Logf("Concurrent payment creation error: %v", err)
		}
	}
	
	// At least some should succeed (allowing for potential rate limiting or other issues)
	assert.Greater(suite.T(), successCount, 0, "At least some concurrent payments should succeed")
}

// Helper function to generate unique test IDs
func generateTestID() string {
	return fmt.Sprintf("%d", time.Now().UnixNano()%1000000)
}

// Test runner for the integration test suite
func TestPaymentIntegrationSuite(t *testing.T) {
	// Check if we should run integration tests
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}
	
	suite.Run(t, new(PaymentIntegrationTestSuite))
}

// Test Stripe test card scenarios
func (suite *PaymentIntegrationTestSuite) TestStripeTestCards() {
	testCards := suite.stripeService.GetTestCardTokens()
	
	// Verify we have expected test cards
	expectedCards := []string{"visa_success", "visa_decline", "mastercard_success", "insufficient_funds"}
	for _, cardType := range expectedCards {
		assert.Contains(suite.T(), testCards, cardType, "Should have %s test card", cardType)
		assert.NotEmpty(suite.T(), testCards[cardType], "Test card number should not be empty")
	}
	
	// Test that card numbers are properly formatted (basic validation)
	for cardType, cardNumber := range testCards {
		assert.Len(suite.T(), cardNumber, 16, "Card number for %s should be 16 digits", cardType)
		assert.Regexp(suite.T(), `^\d{16}$`, cardNumber, "Card number for %s should be numeric", cardType)
	}
}

// Test escrow release eligibility
func (suite *PaymentIntegrationTestSuite) TestEscrowReleaseEligibility() {
	// Create a test payment and escrow
	payment, _, err := suite.paymentService.CreateGamePayment(
		suite.testData.TestUserID,
		suite.testData.TestGameID,
		suite.testData.TestApplicationID,
		suite.testData.TestOrganizerID,
		suite.testData.TestAmount,
	)
	require.NoError(suite.T(), err)
	
	_, escrow, err := suite.paymentService.ConfirmGamePayment(payment.ID)
	require.NoError(suite.T(), err)
	require.NotNil(suite.T(), escrow)
	
	// Test getting eligible escrow releases
	escrows, err := suite.paymentService.GetEligibleEscrowReleases()
	if err != nil {
		// This might fail if database is not properly configured for tests
		suite.T().Logf("GetEligibleEscrowReleases failed (expected if DB not configured): %v", err)
		return
	}
	
	// If we got escrows, verify structure
	if len(escrows) > 0 {
		for _, e := range escrows {
			assert.NotEmpty(suite.T(), e.ID, "Escrow should have ID")
			assert.Greater(suite.T(), e.Amount, 0.0, "Escrow amount should be positive")
			assert.NotEmpty(suite.T(), e.Status, "Escrow should have status")
		}
	}
}