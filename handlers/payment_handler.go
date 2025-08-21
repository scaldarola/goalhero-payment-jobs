package handlers

import (
	"fmt"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/sebastiancaldarola/goalhero-payment-jobs/services"
)

// PaymentHandler handles payment-related endpoints
type PaymentHandler struct {
	paymentService *services.PaymentService
}

// NewPaymentHandler creates a new payment handler
func NewPaymentHandler() *PaymentHandler {
	return &PaymentHandler{
		paymentService: services.NewPaymentService(),
	}
}

// CreateGamePaymentRequest represents the request to create a game payment
type CreateGamePaymentRequest struct {
	UserID        string  `json:"userId" binding:"required"`
	GameID        string  `json:"gameId" binding:"required"`
	ApplicationID string  `json:"applicationId" binding:"required"`
	OrganizerID   string  `json:"organizerId" binding:"required"`
	Amount        float64 `json:"amount" binding:"required,min=5,max=50"`
}

// CreateGamePayment handles POST /api/payments/games
func (h *PaymentHandler) CreateGamePayment(c *gin.Context) {
	var req CreateGamePaymentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		log.Printf("[PaymentHandler] Invalid request: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Invalid request format",
			"details": err.Error(),
		})
		return
	}

	log.Printf("[PaymentHandler] Creating game payment for user %s, game %s", req.UserID, req.GameID)

	payment, result, err := h.paymentService.CreateGamePayment(
		req.UserID,
		req.GameID,
		req.ApplicationID,
		req.OrganizerID,
		req.Amount,
	)

	if err != nil {
		log.Printf("[PaymentHandler] Failed to create payment: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Failed to create payment",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success":        true,
		"payment":        payment,
		"client_secret":  result.ClientSecret,
		"payment_intent": result.PaymentIntent.ID,
	})
}

// ConfirmPaymentRequest represents the request to confirm a payment
type ConfirmPaymentRequest struct {
	PaymentID string `json:"paymentId" binding:"required"`
}

// ConfirmPayment handles POST /api/payments/confirm
func (h *PaymentHandler) ConfirmPayment(c *gin.Context) {
	var req ConfirmPaymentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		log.Printf("[PaymentHandler] Invalid request: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Invalid request format",
			"details": err.Error(),
		})
		return
	}

	log.Printf("[PaymentHandler] Confirming payment: %s", req.PaymentID)

	payment, escrow, err := h.paymentService.ConfirmGamePayment(req.PaymentID)
	if err != nil {
		log.Printf("[PaymentHandler] Failed to confirm payment: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Failed to confirm payment",
			"details": err.Error(),
		})
		return
	}

	response := gin.H{
		"success": true,
		"payment": payment,
	}

	if escrow != nil {
		response["escrow"] = escrow
	}

	c.JSON(http.StatusOK, response)
}

// ReleaseEscrowRequest represents the request to release escrow funds
type ReleaseEscrowRequest struct {
	EscrowID      string `json:"escrowId" binding:"required"`
	ReleaseReason string `json:"releaseReason" binding:"required"`
}

// ReleaseEscrow handles POST /api/payments/escrow/release
func (h *PaymentHandler) ReleaseEscrow(c *gin.Context) {
	var req ReleaseEscrowRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		log.Printf("[PaymentHandler] Invalid request: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Invalid request format",
			"details": err.Error(),
		})
		return
	}

	log.Printf("[PaymentHandler] Releasing escrow: %s", req.EscrowID)

	err := h.paymentService.ProcessEscrowRelease(req.EscrowID, req.ReleaseReason)
	if err != nil {
		log.Printf("[PaymentHandler] Failed to release escrow: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Failed to release escrow",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Escrow released successfully",
		"escrowId": req.EscrowID,
	})
}

// RefundPaymentRequest represents the request to refund a payment
type RefundPaymentRequest struct {
	PaymentID string  `json:"paymentId" binding:"required"`
	Amount    float64 `json:"amount" binding:"required,min=0"`
	Reason    string  `json:"reason" binding:"required"`
}

// RefundPayment handles POST /api/payments/refund
func (h *PaymentHandler) RefundPayment(c *gin.Context) {
	var req RefundPaymentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		log.Printf("[PaymentHandler] Invalid request: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Invalid request format",
			"details": err.Error(),
		})
		return
	}

	log.Printf("[PaymentHandler] Refunding payment: %s, Amount: €%.2f", req.PaymentID, req.Amount)

	err := h.paymentService.ProcessRefund(req.PaymentID, req.Amount, req.Reason)
	if err != nil {
		log.Printf("[PaymentHandler] Failed to process refund: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Failed to process refund",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success":   true,
		"message":   "Refund processed successfully",
		"paymentId": req.PaymentID,
		"amount":    req.Amount,
	})
}

// GetEligibleEscrowReleases handles GET /api/payments/escrow/eligible
func (h *PaymentHandler) GetEligibleEscrowReleases(c *gin.Context) {
	log.Printf("[PaymentHandler] Getting eligible escrow releases")

	escrows, err := h.paymentService.GetEligibleEscrowReleases()
	if err != nil {
		log.Printf("[PaymentHandler] Failed to get eligible escrow releases: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Failed to get eligible escrow releases",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"escrows": escrows,
		"count":   len(escrows),
	})
}

// ProcessEligibleReleases handles POST /api/payments/escrow/process-eligible
func (h *PaymentHandler) ProcessEligibleReleases(c *gin.Context) {
	log.Printf("[PaymentHandler] Processing eligible escrow releases")

	escrows, err := h.paymentService.GetEligibleEscrowReleases()
	if err != nil {
		log.Printf("[PaymentHandler] Failed to get eligible escrow releases: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Failed to get eligible escrow releases",
			"details": err.Error(),
		})
		return
	}

	processed := 0
	failed := 0
	var errors []string

	for _, escrow := range escrows {
		err := h.paymentService.ProcessEscrowRelease(escrow.ID, "automatic_release_job")
		if err != nil {
			log.Printf("[PaymentHandler] Failed to process escrow %s: %v", escrow.ID, err)
			failed++
			errors = append(errors, fmt.Sprintf("Escrow %s: %v", escrow.ID, err))
		} else {
			processed++
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"success":           true,
		"total_eligible":    len(escrows),
		"processed":         processed,
		"failed":            failed,
		"errors":            errors,
	})
}

// GetPaymentStatus handles GET /api/payments/:id/status
func (h *PaymentHandler) GetPaymentStatus(c *gin.Context) {
	paymentID := c.Param("id")
	if paymentID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Payment ID is required",
		})
		return
	}

	log.Printf("[PaymentHandler] Getting payment status: %s", paymentID)

	// This would require implementing a method to get payment by ID
	c.JSON(http.StatusNotImplemented, gin.H{
		"success": false,
		"error":   "Not implemented yet",
	})
}

// GetTestCards handles GET /api/payments/test-cards
func (h *PaymentHandler) GetTestCards(c *gin.Context) {
	stripeService := services.NewStripeConnectService()
	
	if !stripeService.IsTestMode() {
		c.JSON(http.StatusForbidden, gin.H{
			"success": false,
			"error":   "Test cards are only available in test mode",
		})
		return
	}

	testCards := stripeService.GetTestCardTokens()

	c.JSON(http.StatusOK, gin.H{
		"success":    true,
		"test_cards": testCards,
		"note":       "These are test card numbers for Stripe testing",
	})
}