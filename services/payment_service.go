package services

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
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
			ID:                uuid.NewString(),
			GameID:            payment.GameID,
			OrganizerID:       payment.Metadata["organizerID"].(string),
			PaymentID:         payment.ID,
			Amount:            payment.NetAmount,
			Status:            models.EscrowStatusHeld,
			HeldAt:            now,
			ReleaseEligibleAt: now.Add(time.Duration(models.EscrowHoldHours) * time.Hour),
			RatingReceived:    false,
			RatingApproved:    false,
			MinRatingRequired: 3.0, // Minimum rating for auto-release
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
	s.sendSlackSuccessNotification(escrowID, escrow.Amount, releaseReason)
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

// ProcessAutomaticReleases processes all eligible escrow releases automatically
func (s *PaymentService) ProcessAutomaticReleases() (int, int, []string, float64, error) {
	log.Printf("[PaymentService] Processing automatic escrow releases")

	// Get eligible escrow transactions
	escrows, err := s.GetEligibleEscrowReleases()
	if err != nil {
		return 0, 0, nil, 0, fmt.Errorf("failed to get eligible escrow releases: %w", err)
	}

	processed := 0
	failed := 0
	totalReleased := 0.0
	var errors []string

	for _, escrow := range escrows {
		// Check if escrow meets auto-release criteria
		if s.isEligibleForAutoRelease(escrow) {
			err := s.ProcessEscrowRelease(escrow.ID, "automatic_release")
			if err != nil {
				failed++
				errorMsg := fmt.Sprintf("Escrow %s: %v", escrow.ID, err)
				errors = append(errors, errorMsg)
				log.Printf("[PaymentService] Failed to auto-release escrow %s: %v", escrow.ID, err)
				s.sendSlackFailureNotification(escrow.ID, escrow.Amount, err.Error())
			} else {
				processed++
				totalReleased += escrow.Amount
				log.Printf("[PaymentService] Auto-released escrow: %s", escrow.ID)
				s.sendSlackSuccessNotification(escrow.ID, escrow.Amount, "automatic_release")
			}
		} else {
			// Update status to pending_rating if not eligible for auto-release
			escrow.Status = models.EscrowStatusPendingRating
			if err := s.updateEscrowTransaction(escrow); err != nil {
				log.Printf("[PaymentService] Failed to update escrow status: %v", err)
			}
		}
	}

	log.Printf("[PaymentService] Auto-release completed: %d processed, %d failed out of %d eligible",
		processed, failed, len(escrows))
	return processed, failed, errors, totalReleased, nil
}

// isEligibleForAutoRelease checks if an escrow transaction is eligible for automatic release
func (s *PaymentService) isEligibleForAutoRelease(escrow *models.EscrowTransaction) bool {
	// Must be past release eligible time
	if time.Now().Before(escrow.ReleaseEligibleAt) {
		return false
	}

	// Must not be disputed
	if escrow.Status == models.EscrowStatusDisputed {
		return false
	}

	// If rating received, check if it meets minimum threshold
	if escrow.RatingReceived {
		if escrow.ActualRating >= escrow.MinRatingRequired {
			escrow.RatingApproved = true
			return true
		} else {
			// Poor rating - requires manual review - this should send an alert to slack using an environment variable called SLACK_ESCROW_WEBHOOK_URL
			s.sendSlackAlert(escrow.ID, escrow.ActualRating, escrow.MinRatingRequired)
			return false
		}
	}

	// No rating after deadline - check business rules
	// For now, allow auto-release if no rating after 24h grace period
	graceDeadline := escrow.ReleaseEligibleAt.Add(24 * time.Hour)
	if time.Now().After(graceDeadline) {
		log.Printf("[PaymentService] Auto-releasing escrow %s due to no rating after grace period", escrow.ID)
		return true
	}

	return false
}

// UpdateEscrowRating updates the rating for an escrow transaction
func (s *PaymentService) UpdateEscrowRating(escrowID string, rating float64, reviewerID string) error {
	log.Printf("[PaymentService] Updating escrow rating: %s, Rating: %.1f", escrowID, rating)

	escrow, err := s.getEscrowTransaction(escrowID)
	if err != nil {
		return fmt.Errorf("failed to get escrow transaction: %w", err)
	}

	escrow.RatingReceived = true
	escrow.ActualRating = rating
	escrow.ReviewedBy = reviewerID

	// Determine if rating meets minimum threshold
	if rating >= escrow.MinRatingRequired {
		escrow.RatingApproved = true
		escrow.Status = models.EscrowStatusApproved
	} else {
		escrow.RatingApproved = false
		// Poor rating - keep in held status for manual review
	}

	if err := s.updateEscrowTransaction(escrow); err != nil {
		return fmt.Errorf("failed to update escrow transaction: %w", err)
	}

	log.Printf("[PaymentService] Escrow rating updated: %s (Approved: %v)", escrowID, escrow.RatingApproved)
	return nil
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

// SlackMessage represents a Slack webhook message
type SlackMessage struct {
	Text string `json:"text"`
}

// sendSlackAlert sends an alert to Slack for manual review
func (s *PaymentService) sendSlackAlert(escrowID string, rating float64, minRating float64) {
	webhookURL := os.Getenv("SLACK_ESCROW_WEBHOOK_URL")
	if webhookURL == "" {
		log.Printf("[PaymentService] SLACK_ESCROW_WEBHOOK_URL not configured, skipping Slack alert")
		return
	}

	message := SlackMessage{
		Text: fmt.Sprintf("🚨 *Escrow Manual Review Required*\n\nEscrow ID: %s\nActual Rating: %.1f\nMinimum Required: %.1f\n\nThis escrow requires manual review due to poor rating.",
			escrowID, rating, minRating),
	}

	jsonData, err := json.Marshal(message)
	if err != nil {
		log.Printf("[PaymentService] Failed to marshal Slack message: %v", err)
		return
	}

	resp, err := http.Post(webhookURL, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		log.Printf("[PaymentService] Failed to send Slack alert: %v", err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		log.Printf("[PaymentService] Slack alert failed with status: %d", resp.StatusCode)
		return
	}

	log.Printf("[PaymentService] Slack alert sent for escrow %s", escrowID)
}

// sendSlackSuccessNotification sends a success notification to Slack for processed escrow payments
func (s *PaymentService) sendSlackSuccessNotification(escrowID string, amount float64, releaseReason string) {
	webhookURL := os.Getenv("SLACK_ESCROW_WEBHOOK_URL")
	if webhookURL == "" {
		return
	}

	message := SlackMessage{
		Text: fmt.Sprintf("✅ *Escrow Payment Processed Successfully*\n\nEscrow ID: %s\nAmount: €%.2f\nReason: %s\nStatus: Released",
			escrowID, amount, releaseReason),
	}

	s.sendSlackMessage(message, webhookURL)
}

// sendSlackFailureNotification sends a failure notification to Slack for failed escrow payments
func (s *PaymentService) sendSlackFailureNotification(escrowID string, amount float64, errorMsg string) {
	webhookURL := os.Getenv("SLACK_ESCROW_WEBHOOK_URL")
	if webhookURL == "" {
		return
	}

	message := SlackMessage{
		Text: fmt.Sprintf("❌ *Escrow Payment Processing Failed*\n\nEscrow ID: %s\nAmount: €%.2f\nError: %s\nStatus: Failed",
			escrowID, amount, errorMsg),
	}

	s.sendSlackMessage(message, webhookURL)
}

// SendSlackJobSummaryNotification sends a summary notification for payment job execution
func (s *PaymentService) SendSlackJobSummaryNotification(validated, processed, failed int, totalReleased float64, runtime time.Duration) {
	log.Printf("[PaymentService] Sending job summary notification: validated=%d, processed=%d, failed=%d, totalReleased=€%.2f", validated, processed, failed, totalReleased)
	
	webhookURL := os.Getenv("SLACK_ESCROW_WEBHOOK_URL")
	if webhookURL == "" {
		log.Printf("[PaymentService] SLACK_ESCROW_WEBHOOK_URL not configured, skipping job summary notification")
		return
	}
	
	log.Printf("[PaymentService] Using Slack webhook: %s...%s", webhookURL[:20], webhookURL[len(webhookURL)-10:])

	var statusIcon, statusText string
	if failed > 0 {
		statusIcon = "⚠️"
		statusText = "Completed with Issues"
	} else if processed > 0 {
		statusIcon = "✅"
		statusText = "Completed Successfully"
	} else {
		statusIcon = "ℹ️"
		statusText = "No Payments to Process"
	}

	var releaseText string
	if totalReleased > 0 {
		releaseText = fmt.Sprintf("\n💰 **Total Released:** €%.2f", totalReleased)
	} else {
		releaseText = "\n💰 **Money Released:** No payments released"
	}

	message := SlackMessage{
		Text: fmt.Sprintf("%s *Payment Processing Job %s*\n\n📊 **Validation Summary:**\n• Payments Validated: %d\n• Successfully Processed: %d\n• Failed: %d%s\n\n⏱️ **Runtime:** %v\n📅 **Completed:** %s",
			statusIcon, statusText, validated, processed, failed, releaseText, runtime.Round(time.Second), time.Now().Format("2006-01-02 15:04:05 MST")),
	}

	log.Printf("[PaymentService] 📤 Sending job summary to Slack: %s", statusText)
	s.sendSlackMessage(message, webhookURL)
	log.Printf("[PaymentService] ✅ Job summary Slack notification sent successfully!")
}

// sendSlackMessage sends a message to Slack webhook
func (s *PaymentService) sendSlackMessage(message SlackMessage, webhookURL string) {
	log.Printf("[PaymentService] 🌐 Posting to Slack webhook...")
	
	jsonData, err := json.Marshal(message)
	if err != nil {
		log.Printf("[PaymentService] ❌ Failed to marshal Slack message: %v", err)
		return
	}

	resp, err := http.Post(webhookURL, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		log.Printf("[PaymentService] ❌ Failed to send Slack message (network error): %v", err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		log.Printf("[PaymentService] ❌ Slack message failed with HTTP status: %d", resp.StatusCode)
		// Try to read response body for more details
		body := make([]byte, 512)
		if n, err := resp.Body.Read(body); err == nil && n > 0 {
			log.Printf("[PaymentService] Slack error response: %s", string(body[:n]))
		}
		return
	}

	log.Printf("[PaymentService] ✅ Slack message delivered successfully (HTTP %d)", resp.StatusCode)
}
