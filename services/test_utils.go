package services

import (
	"fmt"
	"math/rand"
	"time"

	"github.com/sebastiancaldarola/goalhero-payment-jobs/models"
)

// TestUtilities provides helper functions for testing
type TestUtilities struct{}

// NewTestUtilities creates a new test utilities instance
func NewTestUtilities() *TestUtilities {
	return &TestUtilities{}
}

// GenerateTestPayment creates a test payment with realistic data
func (tu *TestUtilities) GenerateTestPayment() *models.Payment {
	now := time.Now()
	testID := tu.GenerateTestID()
	
	return &models.Payment{
		ID:                fmt.Sprintf("test_payment_%s", testID),
		UserID:            fmt.Sprintf("test_user_%s", testID),
		GameID:            fmt.Sprintf("test_game_%s", testID),
		ApplicationID:     fmt.Sprintf("test_app_%s", testID),
		Amount:            tu.GenerateRandomAmount(),
		PlatformFee:       0.0, // Will be calculated
		PaymentFee:        0.0, // Will be calculated
		NetAmount:         0.0, // Will be calculated
		Currency:          string(models.DefaultCurrency),
		Status:            models.PaymentStatusPending,
		PaymentMethod:     models.PaymentMethodStripe,
		StripePaymentID:   "",
		ClientSecret:      "",
		CreatedAt:         now,
		ConfirmedAt:       nil,
		Metadata:          make(map[string]interface{}),
	}
}

// GenerateTestEscrow creates a test escrow transaction
func (tu *TestUtilities) GenerateTestEscrow(paymentID string, organizerID string, amount float64) *models.EscrowTransaction {
	now := time.Now()
	testID := tu.GenerateTestID()
	
	return &models.EscrowTransaction{
		ID:                  fmt.Sprintf("test_escrow_%s", testID),
		GameID:              fmt.Sprintf("test_game_%s", testID),
		OrganizerID:         organizerID,
		PaymentID:           paymentID,
		Amount:              amount,
		Status:              models.EscrowStatusHeld,
		HeldAt:              now,
		ReleasedAt:          nil,
		ReleaseReason:       "",
		DisputeID:           "",
		ReleaseEligibleAt:   now.Add(time.Duration(models.EscrowHoldHours) * time.Hour),
		RatingReceived:      false,
		RatingApproved:      false,
		MinRatingRequired:   3.0,
		ActualRating:        0.0,
		ReviewedBy:          "",
	}
}

// GenerateTestID creates a unique test identifier
func (tu *TestUtilities) GenerateTestID() string {
	return fmt.Sprintf("%d_%d", time.Now().UnixNano()%1000000, rand.Intn(1000))
}

// GenerateRandomAmount generates a random valid payment amount
func (tu *TestUtilities) GenerateRandomAmount() float64 {
	// Generate amount between €5 and €50 (valid range)
	min := models.MinimumGamePrice
	max := models.MaximumGamePrice
	return min + rand.Float64()*(max-min)
}

// CreateTestOrganizerID generates a test Stripe Connect account ID
func (tu *TestUtilities) CreateTestOrganizerID() string {
	return fmt.Sprintf("acct_test_organizer_%s", tu.GenerateTestID())
}

// TestScenario represents a test scenario configuration
type TestScenario struct {
	Name          string
	Amount        float64
	ExpectedError string
	ShouldSucceed bool
	Description   string
}

// GetPaymentValidationScenarios returns common payment validation test scenarios
func (tu *TestUtilities) GetPaymentValidationScenarios() []TestScenario {
	return []TestScenario{
		{
			Name:          "valid_minimum_amount",
			Amount:        models.MinimumGamePrice,
			ExpectedError: "",
			ShouldSucceed: true,
			Description:   "Payment with minimum valid amount",
		},
		{
			Name:          "valid_maximum_amount",
			Amount:        models.MaximumGamePrice,
			ExpectedError: "",
			ShouldSucceed: true,
			Description:   "Payment with maximum valid amount",
		},
		{
			Name:          "valid_mid_range_amount",
			Amount:        25.0,
			ExpectedError: "",
			ShouldSucceed: true,
			Description:   "Payment with mid-range amount",
		},
		{
			Name:          "below_minimum_amount",
			Amount:        models.MinimumGamePrice - 1.0,
			ExpectedError: "below minimum",
			ShouldSucceed: false,
			Description:   "Payment below minimum amount",
		},
		{
			Name:          "above_maximum_amount",
			Amount:        models.MaximumGamePrice + 10.0,
			ExpectedError: "above maximum",
			ShouldSucceed: false,
			Description:   "Payment above maximum amount",
		},
		{
			Name:          "zero_amount",
			Amount:        0.0,
			ExpectedError: "must be greater than 0",
			ShouldSucceed: false,
			Description:   "Payment with zero amount",
		},
		{
			Name:          "negative_amount",
			Amount:        -5.0,
			ExpectedError: "must be greater than 0",
			ShouldSucceed: false,
			Description:   "Payment with negative amount",
		},
	}
}

// GetParameterValidationScenarios returns test scenarios for parameter validation
func (tu *TestUtilities) GetParameterValidationScenarios() []struct {
	Name          string
	UserID        string
	GameID        string
	ApplicationID string
	OrganizerID   string
	ExpectedError string
} {
	validID := tu.GenerateTestID()
	
	return []struct {
		Name          string
		UserID        string
		GameID        string
		ApplicationID string
		OrganizerID   string
		ExpectedError string
	}{
		{
			Name:          "empty_user_id",
			UserID:        "",
			GameID:        "game_" + validID,
			ApplicationID: "app_" + validID,
			OrganizerID:   "org_" + validID,
			ExpectedError: "user ID cannot be empty",
		},
		{
			Name:          "empty_game_id",
			UserID:        "user_" + validID,
			GameID:        "",
			ApplicationID: "app_" + validID,
			OrganizerID:   "org_" + validID,
			ExpectedError: "game ID cannot be empty",
		},
		{
			Name:          "empty_application_id",
			UserID:        "user_" + validID,
			GameID:        "game_" + validID,
			ApplicationID: "",
			OrganizerID:   "org_" + validID,
			ExpectedError: "application ID cannot be empty",
		},
		{
			Name:          "empty_organizer_id",
			UserID:        "user_" + validID,
			GameID:        "game_" + validID,
			ApplicationID: "app_" + validID,
			OrganizerID:   "",
			ExpectedError: "organizer ID cannot be empty",
		},
	}
}

// MockStripeResponses contains mock Stripe API responses for testing
type MockStripeResponses struct {
	SuccessfulPaymentIntent string
	FailedPaymentIntent     string
	RefundResponse          string
}

// GetMockStripeResponses returns mock Stripe API responses
func (tu *TestUtilities) GetMockStripeResponses() *MockStripeResponses {
	return &MockStripeResponses{
		SuccessfulPaymentIntent: `{
			"id": "pi_test_successful_payment",
			"status": "succeeded",
			"client_secret": "pi_test_successful_payment_secret_12345",
			"amount": 1550,
			"currency": "eur",
			"metadata": {
				"payment_id": "test_payment_123",
				"game_id": "test_game_123"
			}
		}`,
		FailedPaymentIntent: `{
			"id": "pi_test_failed_payment",
			"status": "requires_payment_method",
			"client_secret": "pi_test_failed_payment_secret_12345",
			"amount": 1550,
			"currency": "eur",
			"last_payment_error": {
				"code": "card_declined",
				"message": "Your card was declined."
			}
		}`,
		RefundResponse: `{
			"id": "re_test_refund_123",
			"amount": 775,
			"currency": "eur",
			"status": "succeeded",
			"payment_intent": "pi_test_successful_payment",
			"reason": "requested_by_customer"
		}`,
	}
}

// ValidatePaymentAmounts checks if payment amounts are calculated correctly
func (tu *TestUtilities) ValidatePaymentAmounts(payment *models.Payment, baseAmount float64) bool {
	expectedPlatformFee := baseAmount * models.PlatformFeePercentage / 100
	expectedNetAmount := baseAmount - expectedPlatformFee
	
	platformFeeValid := abs(payment.PlatformFee-expectedPlatformFee) < 0.01
	netAmountValid := abs(payment.NetAmount-expectedNetAmount) < 0.01
	paymentFeePositive := payment.PaymentFee > 0
	
	return platformFeeValid && netAmountValid && paymentFeePositive
}

// ValidateEscrowTransaction checks if escrow transaction is properly structured
func (tu *TestUtilities) ValidateEscrowTransaction(escrow *models.EscrowTransaction, payment *models.Payment) bool {
	if escrow == nil || payment == nil {
		return false
	}
	
	amountValid := abs(escrow.Amount-payment.NetAmount) < 0.01
	paymentIDValid := escrow.PaymentID == payment.ID
	statusValid := escrow.Status == models.EscrowStatusHeld
	eligibilityTimeValid := escrow.ReleaseEligibleAt.After(escrow.HeldAt)
	
	return amountValid && paymentIDValid && statusValid && eligibilityTimeValid
}

// Helper function for floating point comparison
func abs(x float64) float64 {
	if x < 0 {
		return -x
	}
	return x
}

// TestDataCleanup provides cleanup utilities for tests
type TestDataCleanup struct {
	testPaymentIDs []string
	testEscrowIDs  []string
}

// NewTestDataCleanup creates a new cleanup utility
func NewTestDataCleanup() *TestDataCleanup {
	return &TestDataCleanup{
		testPaymentIDs: make([]string, 0),
		testEscrowIDs:  make([]string, 0),
	}
}

// AddPaymentID adds a payment ID for cleanup
func (tdc *TestDataCleanup) AddPaymentID(paymentID string) {
	tdc.testPaymentIDs = append(tdc.testPaymentIDs, paymentID)
}

// AddEscrowID adds an escrow ID for cleanup
func (tdc *TestDataCleanup) AddEscrowID(escrowID string) {
	tdc.testEscrowIDs = append(tdc.testEscrowIDs, escrowID)
}

// GetPaymentIDs returns all registered payment IDs
func (tdc *TestDataCleanup) GetPaymentIDs() []string {
	return tdc.testPaymentIDs
}

// GetEscrowIDs returns all registered escrow IDs
func (tdc *TestDataCleanup) GetEscrowIDs() []string {
	return tdc.testEscrowIDs
}

// Clear removes all registered IDs
func (tdc *TestDataCleanup) Clear() {
	tdc.testPaymentIDs = make([]string, 0)
	tdc.testEscrowIDs = make([]string, 0)
}

// PerformanceTestConfig defines configuration for performance tests
type PerformanceTestConfig struct {
	ConcurrentUsers    int
	PaymentsPerUser    int
	TestDurationSeconds int
	MaxAcceptableLatencyMs int64
}

// GetDefaultPerformanceConfig returns default performance test configuration
func (tu *TestUtilities) GetDefaultPerformanceConfig() *PerformanceTestConfig {
	return &PerformanceTestConfig{
		ConcurrentUsers:        10,
		PaymentsPerUser:        5,
		TestDurationSeconds:    30,
		MaxAcceptableLatencyMs: 5000, // 5 seconds max
	}
}

// LoadTestResult represents the result of a load test
type LoadTestResult struct {
	TotalRequests     int
	SuccessfulRequests int
	FailedRequests    int
	AverageLatencyMs  int64
	MaxLatencyMs      int64
	MinLatencyMs      int64
	RequestsPerSecond float64
	ErrorRate         float64
}

// CalculateLoadTestMetrics calculates metrics from load test results
func (tu *TestUtilities) CalculateLoadTestMetrics(
	totalRequests int,
	successfulRequests int,
	latencies []time.Duration,
	testDuration time.Duration,
) *LoadTestResult {
	result := &LoadTestResult{
		TotalRequests:      totalRequests,
		SuccessfulRequests: successfulRequests,
		FailedRequests:     totalRequests - successfulRequests,
	}
	
	if len(latencies) > 0 {
		var totalLatency int64
		result.MinLatencyMs = latencies[0].Milliseconds()
		result.MaxLatencyMs = latencies[0].Milliseconds()
		
		for _, latency := range latencies {
			ms := latency.Milliseconds()
			totalLatency += ms
			
			if ms < result.MinLatencyMs {
				result.MinLatencyMs = ms
			}
			if ms > result.MaxLatencyMs {
				result.MaxLatencyMs = ms
			}
		}
		
		result.AverageLatencyMs = totalLatency / int64(len(latencies))
	}
	
	if testDuration.Seconds() > 0 {
		result.RequestsPerSecond = float64(totalRequests) / testDuration.Seconds()
	}
	
	if totalRequests > 0 {
		result.ErrorRate = float64(result.FailedRequests) / float64(totalRequests) * 100.0
	}
	
	return result
}