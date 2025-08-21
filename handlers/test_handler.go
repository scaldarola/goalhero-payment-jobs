package handlers

import (
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/sebastiancaldarola/goalhero-payment-jobs/services"
)

// TestHandler handles payment testing endpoints
type TestHandler struct {
	paymentService *services.PaymentService
}

// NewTestHandler creates a new test handler
func NewTestHandler() *TestHandler {
	return &TestHandler{
		paymentService: services.NewPaymentService(),
	}
}

// TestPaymentFlow represents a test payment scenario
type TestPaymentFlow struct {
	Name         string  `json:"name"`
	Description  string  `json:"description"`
	Amount       float64 `json:"amount"`
	ExpectedResult string `json:"expectedResult"`
	TestCard     string  `json:"testCard,omitempty"`
}

// GetTestScenarios handles GET /api/test/scenarios
func (h *TestHandler) GetTestScenarios(c *gin.Context) {
	stripeService := services.NewStripeConnectService()
	
	if !stripeService.IsTestMode() {
		c.JSON(http.StatusForbidden, gin.H{
			"success": false,
			"error":   "Test scenarios are only available in test mode",
		})
		return
	}

	scenarios := []TestPaymentFlow{
		{
			Name:           "successful_payment",
			Description:    "A successful €15 payment with automatic escrow",
			Amount:         15.0,
			ExpectedResult: "Payment succeeds, escrow created",
			TestCard:       "4242424242424242",
		},
		{
			Name:           "declined_card",
			Description:    "Payment declined due to card decline",
			Amount:         20.0,
			ExpectedResult: "Payment fails with decline error",
			TestCard:       "4000000000000002",
		},
		{
			Name:           "insufficient_funds",
			Description:    "Payment fails due to insufficient funds",
			Amount:         25.0,
			ExpectedResult: "Payment fails with insufficient funds error",
			TestCard:       "4000000000009995",
		},
		{
			Name:           "minimum_amount",
			Description:    "Test minimum payment amount (€5)",
			Amount:         5.0,
			ExpectedResult: "Payment succeeds with minimum amount",
			TestCard:       "4242424242424242",
		},
		{
			Name:           "maximum_amount", 
			Description:    "Test maximum payment amount (€50)",
			Amount:         50.0,
			ExpectedResult: "Payment succeeds with maximum amount",
			TestCard:       "4242424242424242",
		},
	}

	c.JSON(http.StatusOK, gin.H{
		"success":   true,
		"scenarios": scenarios,
		"test_cards": stripeService.GetTestCardTokens(),
	})
}

// RunTestScenario handles POST /api/test/scenarios/:scenario
func (h *TestHandler) RunTestScenario(c *gin.Context) {
	scenarioName := c.Param("scenario")
	if scenarioName == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Scenario name is required",
		})
		return
	}

	stripeService := services.NewStripeConnectService()
	if !stripeService.IsTestMode() {
		c.JSON(http.StatusForbidden, gin.H{
			"success": false,
			"error":   "Test scenarios are only available in test mode",
		})
		return
	}

	log.Printf("[TestHandler] Running test scenario: %s", scenarioName)

	// Generate test data
	testUserID := "test_user_" + uuid.New().String()[:8]
	testGameID := "test_game_" + uuid.New().String()[:8]
	testApplicationID := "test_app_" + uuid.New().String()[:8]
	testOrganizerID := "acct_test_organizer" // This would be a real Stripe Connect account ID

	var amount float64
	var expectedResult string

	switch scenarioName {
	case "successful_payment":
		amount = 15.0
		expectedResult = "Payment succeeds, escrow created"
	case "declined_card":
		amount = 20.0
		expectedResult = "Payment fails with decline error"
	case "insufficient_funds":
		amount = 25.0
		expectedResult = "Payment fails with insufficient funds error"
	case "minimum_amount":
		amount = 5.0
		expectedResult = "Payment succeeds with minimum amount"
	case "maximum_amount":
		amount = 50.0
		expectedResult = "Payment succeeds with maximum amount"
	case "below_minimum":
		amount = 3.0
		expectedResult = "Payment fails - below minimum amount"
	case "above_maximum":
		amount = 60.0
		expectedResult = "Payment fails - above maximum amount"
	default:
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Unknown scenario: " + scenarioName,
		})
		return
	}

	startTime := time.Now()
	result := gin.H{
		"success":      true,
		"scenario":     scenarioName,
		"test_data": gin.H{
			"user_id":        testUserID,
			"game_id":        testGameID,
			"application_id": testApplicationID,
			"organizer_id":   testOrganizerID,
			"amount":         amount,
		},
		"expected_result": expectedResult,
		"started_at":      startTime,
	}

	// Step 1: Create payment
	log.Printf("[TestHandler] Step 1: Creating payment")
	payment, paymentResult, err := h.paymentService.CreateGamePayment(
		testUserID,
		testGameID,
		testApplicationID,
		testOrganizerID,
		amount,
	)

	if err != nil {
		result["step1_create_payment"] = gin.H{
			"success": false,
			"error":   err.Error(),
			"note":    "This might be expected for validation error scenarios",
		}
		result["duration"] = time.Since(startTime).String()
		c.JSON(http.StatusOK, result)
		return
	}

	result["step1_create_payment"] = gin.H{
		"success":        true,
		"payment_id":     payment.ID,
		"client_secret":  paymentResult.ClientSecret,
		"payment_intent": paymentResult.PaymentIntent.ID,
		"amount_total":   payment.Amount + payment.PaymentFee,
		"platform_fee":   payment.PlatformFee,
		"payment_fee":    payment.PaymentFee,
		"net_amount":     payment.NetAmount,
	}

	// Step 2: Simulate payment confirmation
	log.Printf("[TestHandler] Step 2: Confirming payment")
	confirmedPayment, escrow, confirmErr := h.paymentService.ConfirmGamePayment(payment.ID)
	
	if confirmErr != nil {
		result["step2_confirm_payment"] = gin.H{
			"success": false,
			"error":   confirmErr.Error(),
			"note":    "Payment confirmation failed - this might be expected for decline scenarios",
		}
	} else {
		escrowResult := gin.H{
			"success":    true,
			"payment_id": confirmedPayment.ID,
			"status":     confirmedPayment.Status,
		}
		
		if escrow != nil {
			escrowResult["escrow_created"] = true
			escrowResult["escrow_id"] = escrow.ID
			escrowResult["escrow_amount"] = escrow.Amount
			escrowResult["escrow_status"] = escrow.Status
			escrowResult["release_eligible_at"] = escrow.ReleaseEligibleAt
		}
		
		result["step2_confirm_payment"] = escrowResult
	}

	result["duration"] = time.Since(startTime).String()
	result["completed_at"] = time.Now()

	c.JSON(http.StatusOK, result)
}

// SimulateEscrowRelease handles POST /api/test/escrow/release
func (h *TestHandler) SimulateEscrowRelease(c *gin.Context) {
	stripeService := services.NewStripeConnectService()
	if !stripeService.IsTestMode() {
		c.JSON(http.StatusForbidden, gin.H{
			"success": false,
			"error":   "Escrow simulation is only available in test mode",
		})
		return
	}

	log.Printf("[TestHandler] Simulating escrow release")

	// Get all eligible escrow releases
	escrows, err := h.paymentService.GetEligibleEscrowReleases()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Failed to get eligible escrows: " + err.Error(),
		})
		return
	}

	if len(escrows) == 0 {
		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"message": "No eligible escrows found for release",
			"count":   0,
		})
		return
	}

	// Release first eligible escrow
	escrow := escrows[0]
	err = h.paymentService.ProcessEscrowRelease(escrow.ID, "test_simulation_release")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Failed to release escrow: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success":             true,
		"message":             "Escrow released successfully",
		"escrow_id":           escrow.ID,
		"amount_released":     escrow.Amount,
		"total_eligible":      len(escrows),
	})
}

// FullPaymentFlow handles POST /api/test/full-flow
func (h *TestHandler) FullPaymentFlow(c *gin.Context) {
	stripeService := services.NewStripeConnectService()
	if !stripeService.IsTestMode() {
		c.JSON(http.StatusForbidden, gin.H{
			"success": false,
			"error":   "Full payment flow testing is only available in test mode",
		})
		return
	}

	log.Printf("[TestHandler] Running full payment flow test")
	
	// Generate test data
	testUserID := "test_user_" + uuid.New().String()[:8]
	testGameID := "test_game_" + uuid.New().String()[:8]
	testApplicationID := "test_app_" + uuid.New().String()[:8]
	testOrganizerID := "acct_test_organizer"
	amount := 15.0

	startTime := time.Now()
	flowSteps := []gin.H{}

	// Step 1: Create payment
	step := gin.H{"step": 1, "name": "create_payment"}
	payment, paymentResult, err := h.paymentService.CreateGamePayment(
		testUserID, testGameID, testApplicationID, testOrganizerID, amount,
	)
	if err != nil {
		step["success"] = false
		step["error"] = err.Error()
		flowSteps = append(flowSteps, step)
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"flow_steps": flowSteps,
			"duration": time.Since(startTime).String(),
		})
		return
	}
	step["success"] = true
	step["payment_id"] = payment.ID
	step["client_secret"] = paymentResult.ClientSecret
	flowSteps = append(flowSteps, step)

	// Step 2: Confirm payment
	step = gin.H{"step": 2, "name": "confirm_payment"}
	confirmedPayment, escrow, err := h.paymentService.ConfirmGamePayment(payment.ID)
	if err != nil {
		step["success"] = false
		step["error"] = err.Error()
		flowSteps = append(flowSteps, step)
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"flow_steps": flowSteps,
			"duration": time.Since(startTime).String(),
		})
		return
	}
	step["success"] = true
	step["payment_status"] = confirmedPayment.Status
	if escrow != nil {
		step["escrow_id"] = escrow.ID
		step["escrow_status"] = escrow.Status
	}
	flowSteps = append(flowSteps, step)

	// Step 3: Release escrow (if created)
	if escrow != nil {
		step = gin.H{"step": 3, "name": "release_escrow"}
		err = h.paymentService.ProcessEscrowRelease(escrow.ID, "full_flow_test")
		if err != nil {
			step["success"] = false
			step["error"] = err.Error()
		} else {
			step["success"] = true
			step["amount_released"] = escrow.Amount
		}
		flowSteps = append(flowSteps, step)
	}

	c.JSON(http.StatusOK, gin.H{
		"success":    true,
		"message":    "Full payment flow completed successfully",
		"flow_steps": flowSteps,
		"duration":   time.Since(startTime).String(),
		"test_data": gin.H{
			"user_id":        testUserID,
			"game_id":        testGameID,
			"payment_id":     payment.ID,
			"escrow_id":      func() string { if escrow != nil { return escrow.ID }; return "" }(),
		},
	})
}