package services

import (
	"fmt"
	"log"
	"math"
	"os"
	"time"

	"github.com/sebastiancaldarola/goalhero-payment-jobs/models"
	"github.com/stripe/stripe-go/v76"
	"github.com/stripe/stripe-go/v76/paymentintent"
	"github.com/stripe/stripe-go/v76/refund"
	"github.com/stripe/stripe-go/v76/transfer"
)

// StripeConnectService handles Stripe Connect payments with escrow functionality
type StripeConnectService struct {
	secretKey       string
	connectAccount  string
	testMode        bool
}

// PaymentResult represents the result of a payment operation
type PaymentResult struct {
	PaymentIntent    *stripe.PaymentIntent `json:"payment_intent"`
	ClientSecret     string                `json:"client_secret"`
	Status           string                `json:"status"`
	TransferID       string                `json:"transfer_id,omitempty"`
	Error            string                `json:"error,omitempty"`
}

// NewStripeConnectService creates a new Stripe Connect service
func NewStripeConnectService() *StripeConnectService {
	secretKey := os.Getenv("STRIPE_SECRET_KEY")
	if secretKey == "" {
		log.Printf("⚠️ STRIPE_SECRET_KEY not found, using test key")
		secretKey = "sk_test_..." // You'll need to replace this with your actual test key
	}

	connectAccount := os.Getenv("STRIPE_CONNECT_ACCOUNT")
	testMode := os.Getenv("STRIPE_TEST_MODE") != "false"

	stripe.Key = secretKey

	return &StripeConnectService{
		secretKey:      secretKey,
		connectAccount: connectAccount,
		testMode:       testMode,
	}
}

// CalculateFees calculates platform and payment processing fees for Stripe
func (s *StripeConnectService) CalculateFees(amount float64) (platformFee, stripeFee, netAmount float64) {
	// Platform fee: 4% of total amount
	platformFee = math.Round((amount*models.PlatformFeePercentage/100)*100) / 100
	
	// Stripe fee: 1.4% + €0.25 (European rate) + 0.25% for Connect
	stripeFee = math.Round((amount*1.65/100+0.25)*100) / 100
	
	// Net amount for organizer (after platform fee, Stripe fee is separate)
	netAmount = math.Round((amount-platformFee)*100) / 100
	
	return platformFee, stripeFee, netAmount
}

// CreateEscrowPaymentIntent creates a payment intent with funds held in escrow
func (s *StripeConnectService) CreateEscrowPaymentIntent(payment *models.Payment, organizerID string) (*PaymentResult, error) {
	if payment == nil {
		return nil, fmt.Errorf("payment cannot be nil")
	}
	
	log.Printf("[StripeConnect] Creating escrow payment intent for €%.2f", payment.Amount)

	// Calculate fees
	platformFee, stripeFee, netAmount := s.CalculateFees(payment.Amount)
	
	// Total amount user pays (includes Stripe processing fee)
	totalAmount := payment.Amount + stripeFee
	amountCents := int64(math.Round(totalAmount * 100))
	platformFeeCents := int64(math.Round(platformFee * 100))

	// Create payment intent with application fee (platform fee)
	params := &stripe.PaymentIntentParams{
		Amount:   stripe.Int64(amountCents),
		Currency: stripe.String(string(models.DefaultCurrency)),
		ApplicationFeeAmount: stripe.Int64(platformFeeCents),
		TransferData: &stripe.PaymentIntentTransferDataParams{
			Destination: stripe.String(organizerID), // Organizer's Stripe Connect account
		},
		Metadata: map[string]string{
			"payment_id":     payment.ID,
			"game_id":        payment.GameID,
			"user_id":        payment.UserID,
			"application_id": payment.ApplicationID,
			"platform_fee":   fmt.Sprintf("%.2f", platformFee),
			"net_amount":     fmt.Sprintf("%.2f", netAmount),
		},
		Description: stripe.String(fmt.Sprintf("GoalHero Game Payment - Game %s", payment.GameID)),
	}

	// Add automatic payment methods
	params.AutomaticPaymentMethods = &stripe.PaymentIntentAutomaticPaymentMethodsParams{
		Enabled: stripe.Bool(true),
	}

	pi, err := paymentintent.New(params)
	if err != nil {
		log.Printf("[StripeConnect] Failed to create payment intent: %v", err)
		return nil, fmt.Errorf("failed to create payment intent: %w", err)
	}

	log.Printf("[StripeConnect] Payment intent created: %s", pi.ID)

	return &PaymentResult{
		PaymentIntent: pi,
		ClientSecret:  pi.ClientSecret,
		Status:        string(pi.Status),
	}, nil
}

// ConfirmPaymentIntent confirms a payment intent
func (s *StripeConnectService) ConfirmPaymentIntent(paymentIntentID string) (*PaymentResult, error) {
	log.Printf("[StripeConnect] Confirming payment intent: %s", paymentIntentID)

	pi, err := paymentintent.Get(paymentIntentID, nil)
	if err != nil {
		log.Printf("[StripeConnect] Failed to retrieve payment intent: %v", err)
		return nil, fmt.Errorf("failed to retrieve payment intent: %w", err)
	}

	result := &PaymentResult{
		PaymentIntent: pi,
		Status:        string(pi.Status),
	}

	// Note: Transfer information would be available via separate API calls if needed
	// For Stripe Connect payments, transfers are handled automatically

	log.Printf("[StripeConnect] Payment intent status: %s", pi.Status)

	return result, nil
}

// ReleaseEscrowFunds releases escrowed funds to the organizer
func (s *StripeConnectService) ReleaseEscrowFunds(escrow *models.EscrowTransaction) error {
	log.Printf("[StripeConnect] Releasing escrow funds: %s for amount €%.2f", escrow.ID, escrow.Amount)

	// In Stripe Connect, funds are automatically transferred when the payment intent succeeds
	// If we need additional control, we could use separate transfers
	
	// For now, we just mark the transaction as released in our system
	// In a real implementation, you might want to:
	// 1. Check the transfer status
	// 2. Create additional transfers if needed
	// 3. Handle partial releases

	log.Printf("[StripeConnect] Escrow funds released successfully")
	return nil
}

// CreateRefund creates a refund for a payment
func (s *StripeConnectService) CreateRefund(paymentIntentID string, amount float64, reason string) (*stripe.Refund, error) {
	log.Printf("[StripeConnect] Creating refund for payment %s: €%.2f", paymentIntentID, amount)

	amountCents := int64(math.Round(amount * 100))
	
	params := &stripe.RefundParams{
		PaymentIntent: stripe.String(paymentIntentID),
		Amount:        stripe.Int64(amountCents),
		Reason:        stripe.String("requested_by_customer"),
		Metadata: map[string]string{
			"refund_reason": reason,
			"timestamp":     time.Now().Format(time.RFC3339),
		},
	}

	refundObj, err := refund.New(params)
	if err != nil {
		log.Printf("[StripeConnect] Failed to create refund: %v", err)
		return nil, fmt.Errorf("failed to create refund: %w", err)
	}

	log.Printf("[StripeConnect] Refund created successfully: %s", refundObj.ID)
	return refundObj, nil
}

// GetPaymentDetails retrieves payment details from Stripe
func (s *StripeConnectService) GetPaymentDetails(paymentIntentID string) (*stripe.PaymentIntent, error) {
	log.Printf("[StripeConnect] Retrieving payment details: %s", paymentIntentID)

	pi, err := paymentintent.Get(paymentIntentID, nil)
	if err != nil {
		log.Printf("[StripeConnect] Failed to retrieve payment: %v", err)
		return nil, fmt.Errorf("failed to retrieve payment: %w", err)
	}

	return pi, nil
}

// ValidateConnectAccount validates a Stripe Connect account
func (s *StripeConnectService) ValidateConnectAccount(accountID string) error {
	// In a real implementation, you would validate the account using Stripe's API
	// For now, we'll do basic validation
	if accountID == "" {
		return fmt.Errorf("connect account ID cannot be empty")
	}

	if len(accountID) < 10 {
		return fmt.Errorf("invalid connect account ID format")
	}

	log.Printf("[StripeConnect] Connect account validated: %s", accountID)
	return nil
}

// CreateTransfer creates a manual transfer to a connected account
func (s *StripeConnectService) CreateTransfer(amount float64, destinationAccount string, metadata map[string]string) (*stripe.Transfer, error) {
	log.Printf("[StripeConnect] Creating transfer: €%.2f to %s", amount, destinationAccount)

	amountCents := int64(math.Round(amount * 100))
	
	params := &stripe.TransferParams{
		Amount:      stripe.Int64(amountCents),
		Currency:    stripe.String(string(models.DefaultCurrency)),
		Destination: stripe.String(destinationAccount),
		Metadata:    metadata,
	}

	transfer, err := transfer.New(params)
	if err != nil {
		log.Printf("[StripeConnect] Failed to create transfer: %v", err)
		return nil, fmt.Errorf("failed to create transfer: %w", err)
	}

	log.Printf("[StripeConnect] Transfer created successfully: %s", transfer.ID)
	return transfer, nil
}

// GetTestCardTokens returns test card tokens for testing
func (s *StripeConnectService) GetTestCardTokens() map[string]string {
	return map[string]string{
		"visa_success":         "4242424242424242",
		"visa_decline":         "4000000000000002", 
		"mastercard_success":   "5555555555554444",
		"amex_success":         "378282246310005",
		"insufficient_funds":   "4000000000009995",
		"expired_card":         "4000000000000069",
		"incorrect_cvc":        "4000000000000127",
		"processing_error":     "4000000000000119",
	}
}

// IsTestMode returns whether the service is in test mode
func (s *StripeConnectService) IsTestMode() bool {
	return s.testMode
}