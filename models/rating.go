package models

import "time"

// RatingValidation represents a player rating with validation workflow
type RatingValidation struct {
	ID                    string     `json:"id" firestore:"id"`
	GameID                string     `json:"gameId" firestore:"gameId"`
	RatedPlayerID         string     `json:"ratedPlayerId" firestore:"ratedPlayerId"`
	RaterID               string     `json:"raterId" firestore:"raterId"`
	Rating                float64    `json:"rating" firestore:"rating"`                           // 1.0 to 5.0
	Comment               string     `json:"comment,omitempty" firestore:"comment,omitempty"`
	ReleaseRecommendation bool       `json:"releaseRecommendation" firestore:"releaseRecommendation"` // Recommends escrow release
	Status                string     `json:"status" firestore:"status"`                           // pending, approved, disputed, resolved
	Approved              bool       `json:"approved" firestore:"approved"`
	DisputeReason         string     `json:"disputeReason,omitempty" firestore:"disputeReason,omitempty"`
	CreatedAt             time.Time  `json:"createdAt" firestore:"createdAt"`
	ReviewedAt            *time.Time `json:"reviewedAt,omitempty" firestore:"reviewedAt,omitempty"`
	ReviewedBy            string     `json:"reviewedBy,omitempty" firestore:"reviewedBy,omitempty"`
	EscrowImpact          string     `json:"escrowImpact,omitempty" firestore:"escrowImpact,omitempty"` // approved, disputed, no_impact
}

// EscrowDispute represents a dispute over escrow funds
type EscrowDispute struct {
	ID                string     `json:"id" firestore:"id"`
	EscrowID          string     `json:"escrowId" firestore:"escrowId"`
	GameID            string     `json:"gameId" firestore:"gameId"`
	DisputerID        string     `json:"disputerId" firestore:"disputerId"`        // Who filed the dispute
	DisputerRole      string     `json:"disputerRole" firestore:"disputerRole"`    // "player" or "organizer"
	DisputeReason     string     `json:"disputeReason" firestore:"disputeReason"`
	Evidence          string     `json:"evidence,omitempty" firestore:"evidence,omitempty"`
	RequestedAction   string     `json:"requestedAction" firestore:"requestedAction"` // "release", "refund", "partial_refund"
	Status            string     `json:"status" firestore:"status"`                   // pending, investigating, resolved, rejected
	AdminID           string     `json:"adminId,omitempty" firestore:"adminId,omitempty"`
	AdminDecision     string     `json:"adminDecision,omitempty" firestore:"adminDecision,omitempty"`
	AdminReasoning    string     `json:"adminReasoning,omitempty" firestore:"adminReasoning,omitempty"`
	ResolutionAmount  float64    `json:"resolutionAmount,omitempty" firestore:"resolutionAmount,omitempty"`
	CreatedAt         time.Time  `json:"createdAt" firestore:"createdAt"`
	ResolvedAt        *time.Time `json:"resolvedAt,omitempty" firestore:"resolvedAt,omitempty"`
	EscalatedAt       *time.Time `json:"escalatedAt,omitempty" firestore:"escalatedAt,omitempty"`
	NotifiedParties   []string   `json:"notifiedParties,omitempty" firestore:"notifiedParties,omitempty"`
}

// Rating validation constants
const (
	// Rating Status
	RatingStatusPending  = "pending"
	RatingStatusApproved = "approved"
	RatingStatusDisputed = "disputed"
	RatingStatusResolved = "resolved"

	// Rating Range
	MinRating = 1.0
	MaxRating = 5.0

	// Rating Thresholds
	AutoApprovalThreshold = 3.0 // Ratings >= 3.0 are auto-approved for escrow release

	// Escrow Impact Types
	EscrowImpactApproved   = "approved"   // Rating approved escrow for release
	EscrowImpactDisputed   = "disputed"   // Rating caused escrow dispute
	EscrowImpactNoImpact   = "no_impact"  // Rating had no escrow impact

	// Default rating requirements
	DefaultMinRatingRequired = 3.0

	// Note: Dispute status constants are defined in payment.go to avoid duplication

	// Dispute Actions
	DisputeActionRelease      = "release"       // Release funds to organizer
	DisputeActionRefund       = "refund"        // Refund to player
	DisputeActionPartialRefund = "partial_refund" // Partial refund with split

	// Dispute Roles
	DisputeRolePlayer    = "player"
	DisputeRoleOrganizer = "organizer"
)