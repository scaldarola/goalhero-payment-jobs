package models

import "time"

type Rating struct {
	Value     int       `json:"value"`     // 1 a 5
	GivenBy   string    `json:"givenBy"`   // UID del equipo que calificó
	CreatedAt time.Time `json:"createdAt"` // Fecha de calificación
}

type User struct {
	UID            string    `json:"uid"`
	Email          string    `json:"email"`
	Name           string    `json:"name"`
	ProfilePicture string    `json:"profilePicture"`
	Credits        float64   `json:"credits"` // Créditos del usuario
	CreatedAt      time.Time `json:"createdAt"`
	Ratings        []Rating  `json:"ratings"`
	AverageRating  float64   `json:"averageRating"`
	GamesPlayed    int       `json:"gamesPlayed"`
	Comment        string    `json:"comment"` // Comentario del usuario
	IsPremium      bool      `json:"isPremium"` // Usuario premium
	PremiumExpiry  *time.Time `json:"premiumExpiry,omitempty"` // Fecha de expiración premium
	FCMToken       string    `json:"fcmToken,omitempty"` // Firebase Cloud Messaging token
}

func (u *User) CalculateAverage() {
	total := 0
	for _, r := range u.Ratings {
		total += r.Value
	}
	if len(u.Ratings) > 0 {
		u.AverageRating = float64(total) / float64(len(u.Ratings))
	} else {
		u.AverageRating = 0.0
	}
}

func (u *User) IsActivePremium() bool {
	if !u.IsPremium {
		return false
	}
	if u.PremiumExpiry == nil {
		return true // Lifetime premium
	}
	return time.Now().Before(*u.PremiumExpiry)
}
