package models

import (
	"time"
)

type Location struct {
	PlaceId   string  `json:"placeId"`
	Name      string  `json:"name"`
	Address   string  `json:"address"`
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
}

// AcceptedPlayer represents a player who has been accepted to a match
type AcceptedPlayer struct {
	PlayerID      string    `json:"playerId"`
	ApplicationID string    `json:"applicationId"`
	AcceptedAt    time.Time `json:"acceptedAt"`
}

type Match struct {
	ID              string            `json:"id"`
	DateTime        time.Time         `json:"dateTime"`
	Zone            string            `json:"zone"`
	Level           string            `json:"level"`
	Comment         string            `json:"comment,omitempty"`
	CreatedAt       time.Time         `json:"createdAt"`
	UpdatedAt       time.Time         `json:"updatedAt,omitempty"`
	Format          string            `json:"format"`                   // 8v8, 5v5, etc.
	MaxSpots        int               `json:"maxSpots"`                 // Máximo de jugadores
	SpotsTaken      int               `json:"spotsTaken"`               // Jugadores que ya se unieron
	Recorded        bool              `json:"recorded"`                 // Si el partido fue grabado
	CreatedBy       string            `json:"createdBy"`                // UID del creador del partido
	Status          string            `json:"status"`                   // open, player_selected, full, completed, cancelled
	Location        *Location         `json:"location,omitempty"`       // Ubicación del partido
	Applications    []*Application    `json:"applications,omitempty"`   // Aplicaciones de jugadores
	AcceptedPlayers []AcceptedPlayer  `json:"acceptedPlayers,omitempty"` // Jugadores aceptados
	CompletedAt     *time.Time        `json:"completedAt,omitempty" firestore:"completedAt,omitempty"` // When match was completed
	CompletionNotes string            `json:"completionNotes,omitempty" firestore:"completionNotes,omitempty"` // Notes from completion
	PlayersPresent  []string          `json:"playersPresent,omitempty" firestore:"playersPresent,omitempty"` // Players who attended
}

// Match status constants
const (
	MatchStatusOpen           = "open"
	MatchStatusPlayerSelected = "player_selected"
	MatchStatusFull           = "full"
	MatchStatusCompleted      = "completed"
	MatchStatusCancelled      = "cancelled"
)
