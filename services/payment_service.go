package services

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/google/uuid"
	"github.com/sebastiancaldarola/goalhero-payment-jobs/config"
	"github.com/sebastiancaldarola/goalhero-payment-jobs/models"
	"google.golang.org/api/iterator"
)

// PaymentService handles payment business logic with Stripe Connect
type PaymentService struct {
	stripeService *StripeConnectService
}

// NewPaymentService creates a new payment service
func NewPaymentService() *PaymentService {
	return &PaymentService{
		stripeService: NewStripeConnectService(),
	}
}

// CreateGamePayment creates a payment for a game with escrow
func (s *PaymentService) CreateGamePayment(userID, gameID, applicationID, organizerID string, amount float64) (*models.Payment, *PaymentResult, error) {
	log.Printf("[PaymentService] Creating game payment: User=%s, Game=%s, Amount=€%.2f", userID, gameID, amount)

	// Validate payment amount
	if err := s.validatePaymentAmount(amount); err != nil {
		return nil, nil, err
	}

	// Calculate fees
	platformFee, stripeFee, netAmount := s.stripeService.CalculateFees(amount)
	
	// Create payment record
	payment := &models.Payment{
		ID:            uuid.NewString(),
		UserID:        userID,
		GameID:        gameID,
		ApplicationID: applicationID,
		Amount:        amount,
		PlatformFee:   platformFee,
		PaymentFee:    stripeFee,
		NetAmount:     netAmount,
		Currency:      models.DefaultCurrency,
		Status:        models.PaymentStatusPending,
		PaymentMethod: models.PaymentMethodStripe,
		CreatedAt:     time.Now(),
		Metadata: map[string]interface{}{
			"userID":        userID,
			"gameID":        gameID,
			"applicationID": applicationID,
			"organizerID":   organizerID,
		},
	}

	// Create Stripe payment intent with escrow
	result, err := s.stripeService.CreateEscrowPaymentIntent(payment, organizerID)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create payment intent: %w", err)
	}

	// Update payment with Stripe details
	payment.StripePaymentID = result.PaymentIntent.ID
	payment.ClientSecret = result.ClientSecret

	// Save payment to Firestore
	if err := s.savePayment(payment); err != nil {
		log.Printf("[PaymentService] Failed to save payment: %v", err)
		// Note: In production, you'd want to cancel the Stripe payment intent here
		return nil, nil, fmt.Errorf("failed to save payment: %w", err)
	}

	log.Printf("[PaymentService] Payment created successfully: %s", payment.ID)
	return payment, result, nil
}

// ConfirmGamePayment confirms a payment and creates escrow transaction
func (s *PaymentService) ConfirmGamePayment(paymentID string) (*models.Payment, *models.EscrowTransaction, error) {
	log.Printf("[PaymentService] Confirming payment: %s", paymentID)

	// Get payment from database
	payment, err := s.getPayment(paymentID)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get payment: %w", err)
	}

	// Confirm with Stripe
	result, err := s.stripeService.ConfirmPaymentIntent(payment.StripePaymentID)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to confirm payment with Stripe: %w", err)
	}

	// Update payment status
	now := time.Now()
	payment.ConfirmedAt = &now
	
	if result.Status == "succeeded" {
		payment.Status = models.PaymentStatusConfirmed
		
		// Create escrow transaction
		escrow := &models.EscrowTransaction{
			ID:                  uuid.NewString(),
			GameID:              payment.GameID,
			OrganizerID:         payment.Metadata["organizerID"].(string),
			PaymentID:           payment.ID,
			Amount:              payment.NetAmount,
			Status:              models.EscrowStatusHeld,
			HeldAt:              now,
			ReleaseEligibleAt:   now.Add(time.Duration(models.EscrowHoldHours) * time.Hour),
			RatingReceived:      false,
			RatingApproved:      false,
			MinRatingRequired:   3.0, // Minimum rating for auto-release
		}

		// Save escrow transaction
		if err := s.saveEscrowTransaction(escrow); err != nil {
			log.Printf("[PaymentService] Failed to save escrow transaction: %v", err)
			return nil, nil, fmt.Errorf("failed to save escrow transaction: %w", err)
		}

		// Update payment
		if err := s.updatePayment(payment); err != nil {
			log.Printf("[PaymentService] Failed to update payment: %v", err)
		}

		log.Printf("[PaymentService] Payment confirmed and escrow created: %s", escrow.ID)
		return payment, escrow, nil
	} else {
		payment.Status = models.PaymentStatusFailed
		if result.PaymentIntent.LastPaymentError != nil {
			payment.FailureReason = result.PaymentIntent.LastPaymentError.Msg
		}
		
		if err := s.updatePayment(payment); err != nil {
			log.Printf("[PaymentService] Failed to update payment: %v", err)
		}
		
		log.Printf("[PaymentService] Payment failed: %s", payment.ID)
		return payment, nil, fmt.Errorf("payment failed: %s", payment.FailureReason)
	}
}

// ProcessEscrowRelease processes the release of escrowed funds
func (s *PaymentService) ProcessEscrowRelease(escrowID, releaseReason string) error {
	log.Printf("[PaymentService] Processing escrow release: %s", escrowID)

	// Get escrow transaction
	escrow, err := s.getEscrowTransaction(escrowID)
	if err != nil {
		return fmt.Errorf("failed to get escrow transaction: %w", err)
	}

	if escrow.Status != models.EscrowStatusHeld && escrow.Status != models.EscrowStatusApproved {
		return fmt.Errorf("escrow cannot be released, current status: %s", escrow.Status)
	}

	// Release funds via Stripe
	if err := s.stripeService.ReleaseEscrowFunds(escrow); err != nil {
		return fmt.Errorf("failed to release funds via Stripe: %w", err)
	}

	// Update escrow status
	now := time.Now()
	escrow.Status = models.EscrowStatusReleased
	escrow.ReleasedAt = &now
	escrow.ReleaseReason = releaseReason

	// Save updated escrow transaction
	if err := s.updateEscrowTransaction(escrow); err != nil {
		return fmt.Errorf("failed to update escrow transaction: %w", err)
	}

	log.Printf("[PaymentService] Escrow released successfully: %s", escrowID)
	return nil
}

// ProcessRefund processes a payment refund
func (s *PaymentService) ProcessRefund(paymentID string, amount float64, reason string) error {
	log.Printf("[PaymentService] Processing refund: %s, Amount: €%.2f", paymentID, amount)

	// Get payment from database
	payment, err := s.getPayment(paymentID)
	if err != nil {
		return fmt.Errorf("failed to get payment: %w", err)
	}

	if payment.Status != models.PaymentStatusConfirmed {
		return fmt.Errorf("payment cannot be refunded, current status: %s", payment.Status)
	}

	// Process refund via Stripe
	_, err = s.stripeService.CreateRefund(payment.StripePaymentID, amount, reason)
	if err != nil {
		return fmt.Errorf("failed to process refund via Stripe: %w", err)
	}

	// Update payment status
	payment.Status = models.PaymentStatusRefunded

	if err := s.updatePayment(payment); err != nil {
		return fmt.Errorf("failed to update payment: %w", err)
	}

	log.Printf("[PaymentService] Refund processed successfully: %s", paymentID)
	return nil
}

// GetEligibleEscrowReleases gets escrow transactions eligible for release
func (s *PaymentService) GetEligibleEscrowReleases() ([]*models.EscrowTransaction, error) {
	log.Printf("[PaymentService] Getting eligible escrow releases")

	firestoreClient := config.FirestoreClient()
	if firestoreClient == nil {
		return nil, fmt.Errorf("firestore client not available")
	}

	ctx := context.Background()
	now := time.Now()

	// Query for escrow transactions that are eligible for release
	query := firestoreClient.Collection("escrow_transactions").
		Where("status", "==", models.EscrowStatusHeld).
		Where("releaseEligibleAt", "<=", now)

	iter := query.Documents(ctx)
	defer iter.Stop()

	var escrows []*models.EscrowTransaction
	for {
		doc, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("failed to iterate escrow transactions: %w", err)
		}

		var escrow models.EscrowTransaction
		if err := doc.DataTo(&escrow); err != nil {
			log.Printf("[PaymentService] Failed to parse escrow transaction: %v", err)
			continue
		}

		escrows = append(escrows, &escrow)
	}

	log.Printf("[PaymentService] Found %d eligible escrow releases", len(escrows))
	return escrows, nil
}

// validatePaymentAmount validates payment amount against business rules
func (s *PaymentService) validatePaymentAmount(amount float64) error {
	if amount < models.MinimumGamePrice {
		return fmt.Errorf("minimum payment amount is €%.2f", models.MinimumGamePrice)
	}
	
	if amount > models.MaximumGamePrice {
		return fmt.Errorf("maximum payment amount is €%.2f", models.MaximumGamePrice)
	}
	
	return nil
}

// Database operations
func (s *PaymentService) savePayment(payment *models.Payment) error {
	firestoreClient := config.FirestoreClient()
	if firestoreClient == nil {
		return fmt.Errorf("firestore client not available")
	}

	ctx := context.Background()
	_, err := firestoreClient.Collection("payments").Doc(payment.ID).Set(ctx, payment)
	return err
}

func (s *PaymentService) updatePayment(payment *models.Payment) error {
	firestoreClient := config.FirestoreClient()
	if firestoreClient == nil {
		return fmt.Errorf("firestore client not available")
	}

	ctx := context.Background()
	_, err := firestoreClient.Collection("payments").Doc(payment.ID).Set(ctx, payment)
	return err
}

func (s *PaymentService) getPayment(paymentID string) (*models.Payment, error) {
	firestoreClient := config.FirestoreClient()
	if firestoreClient == nil {
		return nil, fmt.Errorf("firestore client not available")
	}

	ctx := context.Background()
	doc, err := firestoreClient.Collection("payments").Doc(paymentID).Get(ctx)
	if err != nil {
		return nil, err
	}

	var payment models.Payment
	if err := doc.DataTo(&payment); err != nil {
		return nil, err
	}

	return &payment, nil
}

func (s *PaymentService) saveEscrowTransaction(escrow *models.EscrowTransaction) error {
	firestoreClient := config.FirestoreClient()
	if firestoreClient == nil {
		return fmt.Errorf("firestore client not available")
	}

	ctx := context.Background()
	_, err := firestoreClient.Collection("escrow_transactions").Doc(escrow.ID).Set(ctx, escrow)
	return err
}

func (s *PaymentService) updateEscrowTransaction(escrow *models.EscrowTransaction) error {
	firestoreClient := config.FirestoreClient()
	if firestoreClient == nil {
		return fmt.Errorf("firestore client not available")
	}

	ctx := context.Background()
	_, err := firestoreClient.Collection("escrow_transactions").Doc(escrow.ID).Set(ctx, escrow)
	return err
}

func (s *PaymentService) getEscrowTransaction(escrowID string) (*models.EscrowTransaction, error) {
	firestoreClient := config.FirestoreClient()
	if firestoreClient == nil {
		return nil, fmt.Errorf("firestore client not available")
	}

	ctx := context.Background()
	doc, err := firestoreClient.Collection("escrow_transactions").Doc(escrowID).Get(ctx)
	if err != nil {
		return nil, err
	}

	var escrow models.EscrowTransaction
	if err := doc.DataTo(&escrow); err != nil {
		return nil, err
	}

	return &escrow, nil
}