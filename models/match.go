package models

import (
	"time"
)

// MatchStatus represents the current status of a match
type MatchStatus string

const (
	MatchStatusPending   MatchStatus = "pending"
	MatchStatusConfirmed MatchStatus = "confirmed"
	MatchStatusCompleted MatchStatus = "completed"
	MatchStatusCancelled MatchStatus = "cancelled"
)

// Match represents a football match
type Match struct {
	ID               string            `json:"id" firestore:"id"`
	Title            string            `json:"title" firestore:"title"`
	Description      string            `json:"description" firestore:"description"`
	Location         string            `json:"location" firestore:"location"`
	ScheduledAt      time.Time         `json:"scheduledAt" firestore:"scheduledAt"`
	Duration         int               `json:"duration" firestore:"duration"` // in minutes
	CreatedBy        string            `json:"createdBy" firestore:"createdBy"`
	MaxPlayers       int               `json:"maxPlayers" firestore:"maxPlayers"`
	MinPlayers       int               `json:"minPlayers" firestore:"minPlayers"`
	CurrentPlayers   int               `json:"currentPlayers" firestore:"currentPlayers"`
	Status           MatchStatus       `json:"status" firestore:"status"`
	PlayersPresent   []string          `json:"playersPresent" firestore:"playersPresent"`
	PlayersAbsent    []string          `json:"playersAbsent" firestore:"playersAbsent"`
	Ratings          map[string]Rating `json:"ratings" firestore:"ratings"`
	PricePerPlayer   float64           `json:"pricePerPlayer" firestore:"pricePerPlayer"`
	TotalPrice       float64           `json:"totalPrice" firestore:"totalPrice"`
	CreatedAt        time.Time         `json:"createdAt" firestore:"createdAt"`
	UpdatedAt        time.Time         `json:"updatedAt" firestore:"updatedAt"`
	CompletedAt      *time.Time        `json:"completedAt,omitempty" firestore:"completedAt,omitempty"`
	PaymentStatus    PaymentStatus     `json:"paymentStatus" firestore:"paymentStatus"`
	EscrowAccountID  string            `json:"escrowAccountId,omitempty" firestore:"escrowAccountId,omitempty"`
}

// Rating represents a player's rating for a match
type Rating struct {
	PlayerID    string    `json:"playerId" firestore:"playerId"`
	RatedBy     string    `json:"ratedBy" firestore:"ratedBy"`
	Score       float64   `json:"score" firestore:"score"`       // 1-5 stars
	Comment     string    `json:"comment" firestore:"comment"`
	CreatedAt   time.Time `json:"createdAt" firestore:"createdAt"`
}

// PaymentStatus represents the payment status of a match
type PaymentStatus string

const (
	PaymentStatusPending    PaymentStatus = "pending"
	PaymentStatusEscrowed   PaymentStatus = "escrowed"
	PaymentStatusReleased   PaymentStatus = "released"
	PaymentStatusDisputed   PaymentStatus = "disputed"
	PaymentStatusRefunded   PaymentStatus = "refunded"
)