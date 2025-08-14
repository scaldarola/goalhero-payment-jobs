package models

import "time"

// Payment represents a payment transaction in the system
type Payment struct {
	ID                string                 `json:"id" firestore:"id"`
	UserID            string                 `json:"userId" firestore:"userId"`
	GameID            string                 `json:"gameId" firestore:"gameId"`
	ApplicationID     string                 `json:"applicationId" firestore:"applicationId"`
	Amount            float64                `json:"amount" firestore:"amount"`                       // Amount in EUR
	PlatformFee       float64                `json:"platformFee" firestore:"platformFee"`             // 4% platform fee
	PaymentFee        float64                `json:"paymentFee" firestore:"paymentFee"`               // Stripe/PayPal fees
	NetAmount         float64                `json:"netAmount" firestore:"netAmount"`                 // Amount after fees
	Currency          string                 `json:"currency" firestore:"currency"`                   // EUR
	Status            string                 `json:"status" firestore:"status"`                       // pending, confirmed, failed, refunded
	PaymentMethod     string                 `json:"paymentMethod" firestore:"paymentMethod"`         // stripe, paypal
	StripePaymentID   string                 `json:"stripePaymentId,omitempty" firestore:"stripePaymentId,omitempty"`
	PayPalPaymentID   string                 `json:"paypalPaymentId,omitempty" firestore:"paypalPaymentId,omitempty"`
	ClientSecret      string                 `json:"clientSecret,omitempty" firestore:"clientSecret,omitempty"`
	FailureReason     string                 `json:"failureReason,omitempty" firestore:"failureReason,omitempty"`
	CreatedAt         time.Time              `json:"createdAt" firestore:"createdAt"`
	ConfirmedAt       *time.Time             `json:"confirmedAt,omitempty" firestore:"confirmedAt,omitempty"`
	Metadata          map[string]interface{} `json:"metadata,omitempty" firestore:"metadata,omitempty"`
}

// EscrowTransaction represents funds held in escrow
type EscrowTransaction struct {
	ID                  string     `json:"id" firestore:"id"`
	GameID              string     `json:"gameId" firestore:"gameId"`
	OrganizerID         string     `json:"organizerId" firestore:"organizerId"`
	PaymentID           string     `json:"paymentId" firestore:"paymentId"`
	Amount              float64    `json:"amount" firestore:"amount"`
	Status              string     `json:"status" firestore:"status"`         // held, pending_rating, approved, released, disputed, resolved, refunded
	HeldAt              time.Time  `json:"heldAt" firestore:"heldAt"`
	ReleasedAt          *time.Time `json:"releasedAt,omitempty" firestore:"releasedAt,omitempty"`
	ReleaseReason       string     `json:"releaseReason,omitempty" firestore:"releaseReason,omitempty"`
	DisputeID           string     `json:"disputeId,omitempty" firestore:"disputeId,omitempty"`
	ReleaseEligibleAt   time.Time  `json:"releaseEligibleAt" firestore:"releaseEligibleAt"`
	RatingReceived      bool       `json:"ratingReceived" firestore:"ratingReceived"`
	RatingApproved      bool       `json:"ratingApproved" firestore:"ratingApproved"`
	MinRatingRequired   float64    `json:"minRatingRequired" firestore:"minRatingRequired"`
	ActualRating        float64    `json:"actualRating,omitempty" firestore:"actualRating,omitempty"`
	ReviewedBy          string     `json:"reviewedBy,omitempty" firestore:"reviewedBy,omitempty"`
}

// UserPaymentMethod represents stored payment methods
type UserPaymentMethod struct {
	ID            string    `json:"id" firestore:"id"`
	UserID        string    `json:"userId" firestore:"userId"`
	Type          string    `json:"type" firestore:"type"`                   // card, sepa, paypal
	Provider      string    `json:"provider" firestore:"provider"`           // stripe, paypal
	ProviderToken string    `json:"providerToken" firestore:"providerToken"` // Tokenized payment method
	LastFour      string    `json:"lastFour,omitempty" firestore:"lastFour,omitempty"`
	Brand         string    `json:"brand,omitempty" firestore:"brand,omitempty"`     // visa, mastercard, etc.
	ExpiryMonth   int       `json:"expiryMonth,omitempty" firestore:"expiryMonth,omitempty"`
	ExpiryYear    int       `json:"expiryYear,omitempty" firestore:"expiryYear,omitempty"`
	IsDefault     bool      `json:"isDefault" firestore:"isDefault"`
	CreatedAt     time.Time `json:"createdAt" firestore:"createdAt"`
}

// Payout represents payments to organizers
type Payout struct {
	ID              string     `json:"id" firestore:"id"`
	OrganizerID     string     `json:"organizerId" firestore:"organizerId"`
	Amount          float64    `json:"amount" firestore:"amount"`
	Currency        string     `json:"currency" firestore:"currency"`
	Status          string     `json:"status" firestore:"status"`         // pending, processing, completed, failed
	PayoutMethod    string     `json:"payoutMethod" firestore:"payoutMethod"` // bank_transfer, paypal
	BankAccount     string     `json:"bankAccount,omitempty" firestore:"bankAccount,omitempty"`
	PayPalEmail     string     `json:"paypalEmail,omitempty" firestore:"paypalEmail,omitempty"`
	StripePayoutID  string     `json:"stripePayoutId,omitempty" firestore:"stripePayoutId,omitempty"`
	PayPalPayoutID  string     `json:"paypalPayoutId,omitempty" firestore:"paypalPayoutId,omitempty"`
	FailureReason   string     `json:"failureReason,omitempty" firestore:"failureReason,omitempty"`
	RequestedAt     time.Time  `json:"requestedAt" firestore:"requestedAt"`
	CompletedAt     *time.Time `json:"completedAt,omitempty" firestore:"completedAt,omitempty"`
	EscrowIDs       []string   `json:"escrowIds" firestore:"escrowIds"` // IDs of escrow transactions being paid out
}

// PaymentDispute represents disputes and refunds
type PaymentDispute struct {
	ID          string     `json:"id" firestore:"id"`
	PaymentID   string     `json:"paymentId" firestore:"paymentId"`
	GameID      string     `json:"gameId" firestore:"gameId"`
	UserID      string     `json:"userId" firestore:"userId"`
	OrganizerID string     `json:"organizerId" firestore:"organizerId"`
	Type        string     `json:"type" firestore:"type"`         // cancellation, no_show, fraud, other
	Reason      string     `json:"reason" firestore:"reason"`
	Status      string     `json:"status" firestore:"status"`     // open, investigating, resolved, rejected
	Resolution  string     `json:"resolution,omitempty" firestore:"resolution,omitempty"` // full_refund, partial_refund, no_refund
	RefundAmount float64   `json:"refundAmount,omitempty" firestore:"refundAmount,omitempty"`
	CreatedAt   time.Time  `json:"createdAt" firestore:"createdAt"`
	ResolvedAt  *time.Time `json:"resolvedAt,omitempty" firestore:"resolvedAt,omitempty"`
	AdminNotes  string     `json:"adminNotes,omitempty" firestore:"adminNotes,omitempty"`
}

// UserWallet represents user's wallet balance
type UserWallet struct {
	UserID          string    `json:"userId" firestore:"userId"`
	Balance         float64   `json:"balance" firestore:"balance"`           // Available balance in EUR
	PendingBalance  float64   `json:"pendingBalance" firestore:"pendingBalance"` // Funds in escrow
	TotalEarned     float64   `json:"totalEarned" firestore:"totalEarned"`
	TotalSpent      float64   `json:"totalSpent" firestore:"totalSpent"`
	LastUpdated     time.Time `json:"lastUpdated" firestore:"lastUpdated"`
}

// PaymentConstants for business logic
const (
	// Payment Status
	PaymentStatusPending   = "pending"
	PaymentStatusConfirmed = "confirmed"
	PaymentStatusFailed    = "failed"
	PaymentStatusRefunded  = "refunded"

	// Escrow Status
	EscrowStatusHeld          = "held"
	EscrowStatusPendingRating = "pending_rating"
	EscrowStatusApproved      = "approved"
	EscrowStatusReleased      = "released"
	EscrowStatusDisputed      = "disputed"
	EscrowStatusResolved      = "resolved"
	EscrowStatusRefunded      = "refunded"

	// Payment Methods
	PaymentMethodStripe = "stripe"
	PaymentMethodPayPal = "paypal"

	// Payout Status
	PayoutStatusPending    = "pending"
	PayoutStatusProcessing = "processing"
	PayoutStatusCompleted  = "completed"
	PayoutStatusFailed     = "failed"

	// Dispute Status
	DisputeStatusPending       = "pending"
	DisputeStatusOpen          = "open"
	DisputeStatusInvestigating = "investigating"
	DisputeStatusResolved      = "resolved"
	DisputeStatusRejected      = "rejected"

	// Business Rules
	PlatformFeePercentage = 4.0    // 4%
	MinimumGamePrice     = 5.0     // €5
	MaximumGamePrice     = 50.0    // €50
	EscrowHoldHours      = 24      // 24 hours after game ends
	
	// Currency
	DefaultCurrency = "EUR"
)