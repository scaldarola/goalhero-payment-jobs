package models

import "time"

type Application struct {
	ID        string    `json:"id"`
	PlayerID  string    `json:"playerId"`          // UID del arquero
	Status    string    `json:"status"`            // pending, open, accepted, rejected, not_selected, withdrawn
	Comment   string    `json:"comment,omitempty"` // comentario opcional
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt,omitempty"`
	Rating    float64   `json:"rating,omitempty"`
	Price     float64   `json:"price"` // Precio de la postulacion
	UserName  string    `json:"userName,omitempty"`
	Rated     bool      `json:"rated,omitempty"`   // Indica si la postulacion fue valorada
	MatchID   string    `json:"matchId,omitempty"` // ID del partido al que se postuló
}

// Application status constants
const (
	ApplicationStatusPending     = "pending"
	ApplicationStatusOpen        = "open"
	ApplicationStatusAccepted    = "accepted"
	ApplicationStatusRejected    = "rejected"
	ApplicationStatusNotSelected = "not_selected"
	ApplicationStatusWithdrawn   = "withdrawn"
)
