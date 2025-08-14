package models

import (
	"strings"
	"time"
)

// UserProfile represents the complete user profile stored in Firestore
type UserProfile struct {
	ID               string    `firestore:"-" json:"id"` // Document ID = Firebase UID
	Name             string    `firestore:"name" json:"name"`
	Age              int       `firestore:"age" json:"age"`
	ProfilePhotoURL  string    `firestore:"profilePhotoUrl" json:"profilePhoto"`
	Bio              string    `firestore:"bio" json:"bio"`
	IsVerified       bool      `firestore:"isVerified" json:"isVerified"`
	Position         string    `firestore:"position" json:"position"`
	Level            string    `firestore:"level" json:"level"`
	PlayingStyle     string    `firestore:"playingStyle" json:"playingStyle"`
	PreferredFormats []string  `firestore:"preferredFormats" json:"preferredFormats"`
	Availability     string    `firestore:"availability" json:"availability"`
	CreatedAt        time.Time `firestore:"createdAt" json:"joinedDate"`
	UpdatedAt        time.Time `firestore:"updatedAt" json:"-"`
	LastActive       time.Time `firestore:"lastActive" json:"lastActive"`
}

// UserStatistics represents pre-computed user statistics
type UserStatistics struct {
	UserID                  string    `firestore:"-" json:"userId"` // Document ID
	GamesPlayed            int       `firestore:"gamesPlayed" json:"gamesPlayed"`
	GamesCompleted         int       `firestore:"gamesCompleted" json:"gamesCompleted"`
	GamesCancelled         int       `firestore:"gamesCancelled" json:"gamesCancelled"`
	GamesThisMonth         int       `firestore:"gamesThisMonth" json:"gamesThisMonth"`
	TotalRatingPoints      float64   `firestore:"totalRatingPoints" json:"-"`
	TotalReviews           int       `firestore:"totalReviews" json:"totalReviews"`
	OverallRating          float64   `firestore:"overallRating" json:"overallRating"`
	Last10Average          float64   `firestore:"last10Average" json:"last10Average"`
	AvgResponseTimeMinutes int       `firestore:"avgResponseTimeMinutes" json:"responseTimeMinutes"`
	PunctualArrivals       int       `firestore:"punctualArrivals" json:"-"`
	TotalArrivals          int       `firestore:"totalArrivals" json:"-"`
	PunctualityRate        int       `firestore:"punctualityRate" json:"punctualityRate"`
	CompletionRate         int       `firestore:"completionRate" json:"completionRate"`
	CancelRate             int       `firestore:"cancelRate" json:"cancelRate"`
	ReliabilityScore       float64   `firestore:"reliabilityScore" json:"reliabilityScore"`
	LastCalculated         time.Time `firestore:"lastCalculated" json:"-"`
}

// GameReview represents a review given to a user after a game
type GameReview struct {
	ID             string    `firestore:"-" json:"id"`
	GameID         string    `firestore:"gameId" json:"gameId"`
	ReviewerID     string    `firestore:"reviewerId" json:"reviewerId"`
	ReviewedUserID string    `firestore:"reviewedUserId" json:"reviewedUserId"`
	Rating         float64   `firestore:"rating" json:"rating"`
	Comment        string    `firestore:"comment" json:"comment"`
	GameLocation   string    `firestore:"gameLocation" json:"gameLocation"`
	ReviewerName   string    `firestore:"reviewerName" json:"reviewerName"`
	CreatedAt      time.Time `firestore:"createdAt" json:"date"`
}

// UserPricingStats represents pricing statistics for a user
type UserPricingStats struct {
	UserID       string    `firestore:"-" json:"userId"`
	AveragePrice float64   `firestore:"averagePrice" json:"averagePrice"`
	MinPrice     float64   `firestore:"minPrice" json:"minPrice"`
	MaxPrice     float64   `firestore:"maxPrice" json:"maxPrice"`
	TotalGames   int       `firestore:"totalGames" json:"-"`
	UpdatedAt    time.Time `firestore:"updatedAt" json:"-"`
}

// Response structures for API endpoints

// ProfileResponse is the main response structure for user profiles
type ProfileResponse struct {
	ID            string          `json:"id"`
	Name          string          `json:"name"`
	Age           int             `json:"age"`
	ProfilePhoto  string          `json:"profilePhoto"`
	IsVerified    bool            `json:"isVerified"`
	Ratings       RatingInfo      `json:"ratings"`
	Statistics    StatisticsInfo  `json:"statistics"`
	Performance   PerformanceInfo `json:"performance"`
	Profile       ProfileInfo     `json:"profile"`
	RecentReviews []ReviewInfo    `json:"recentReviews"`
}

// RatingInfo contains rating-related information
type RatingInfo struct {
	Overall       float64 `json:"overall"`
	TotalReviews  int     `json:"totalReviews"`
	Last10Average float64 `json:"last10Average"`
}

// StatisticsInfo contains user statistics
type StatisticsInfo struct {
	GamesPlayed         int     `json:"gamesPlayed"`
	CompletionRate      int     `json:"completionRate"`
	ResponseTimeMinutes int     `json:"responseTimeMinutes"`
	ReliabilityScore    float64 `json:"reliabilityScore"`
	GamesThisMonth      int     `json:"gamesThisMonth"`
	CancelRate          int     `json:"cancelRate"`
	PunctualityRate     int     `json:"punctualityRate"`
}

// PerformanceInfo contains performance-related information
type PerformanceInfo struct {
	AveragePrice     float64    `json:"averagePrice"`
	PriceRange       PriceRange `json:"priceRange"`
	Position         string     `json:"position"`
	Level            string     `json:"level"`
	PlayingStyle     string     `json:"playingStyle"`
	PreferredFormats []string   `json:"preferredFormats"`
}

// PriceRange represents min/max pricing
type PriceRange struct {
	Min float64 `json:"min"`
	Max float64 `json:"max"`
}

// ProfileInfo contains profile information
type ProfileInfo struct {
	Bio          string    `json:"bio"`
	Availability string    `json:"availability"`
	JoinedDate   time.Time `json:"joinedDate"`
	LastActive   time.Time `json:"lastActive"`
}

// ReviewInfo contains review information for the profile response
type ReviewInfo struct {
	ID           string    `json:"id"`
	Rating       float64   `json:"rating"`
	Comment      string    `json:"comment"`
	ReviewerName string    `json:"reviewerName"`
	GameLocation string    `json:"gameLocation"`
	Date         time.Time `json:"date"`
}

// UserGameHistory represents a user's game participation history
type UserGameHistory struct {
	ID           string    `json:"id"`
	GameID       string    `json:"gameId"`
	GameTitle    string    `json:"gameTitle"`
	Zone         string    `json:"zone"`
	DateTime     time.Time `json:"dateTime"`
	Status       string    `json:"status"` // completed, cancelled, no_show
	Role         string    `json:"role"`   // player, organizer
	Rating       float64   `json:"rating,omitempty"`
	Price        float64   `json:"price,omitempty"`
}

// Default values constants
const (
	DefaultUserName        = "Usuario"
	DefaultAge             = 25
	DefaultBio             = "Jugador apasionado por el fútbol"
	DefaultPosition        = "Centrocampista"
	DefaultLevel           = "Intermedio"
	DefaultPlayingStyle    = "Versátil"
	DefaultAvailability    = "Fines de semana"
	DefaultOverallRating   = 4.0
	DefaultResponseTime    = 15
	DefaultPunctualityRate = 95
	DefaultCompletionRate  = 90
	DefaultCancelRate      = 5
	DefaultReliabilityScore = 8.0
	DefaultAveragePrice    = 12.0
	DefaultMinPrice        = 10.0
	DefaultMaxPrice        = 15.0
)

// Spanish position mappings
var SpanishPositions = map[string]string{
	"goalkeeper":     "Portero",
	"defender":       "Defensa",
	"midfielder":     "Centrocampista",
	"forward":        "Delantero",
	"winger":         "Extremo",
	"center_back":    "Central",
	"fullback":       "Lateral",
	"attacking_mid":  "Mediapunta",
	"defensive_mid":  "Pivote",
}

// Spanish level mappings
var SpanishLevels = map[string]string{
	"beginner":     "Principiante",
	"intermediate": "Intermedio",
	"advanced":     "Avanzado",
	"expert":       "Experto",
}

// Helper functions for setting defaults

// SetUserProfileDefaults sets default values for missing user profile data
func SetUserProfileDefaults(profile *UserProfile) {
	if profile.Name == "" {
		profile.Name = DefaultUserName
	}
	if profile.Age == 0 {
		profile.Age = DefaultAge
	}
	if profile.Bio == "" {
		profile.Bio = DefaultBio
	}
	if profile.Position == "" {
		profile.Position = DefaultPosition
	}
	if profile.Level == "" {
		profile.Level = DefaultLevel
	}
	if profile.PlayingStyle == "" {
		profile.PlayingStyle = DefaultPlayingStyle
	}
	if len(profile.PreferredFormats) == 0 {
		profile.PreferredFormats = []string{"5v5"}
	}
	if profile.Availability == "" {
		profile.Availability = DefaultAvailability
	}
	if profile.ProfilePhotoURL == "" {
		profile.ProfilePhotoURL = ""
	}
}

// SetStatisticsDefaults sets default values for missing statistics
func SetStatisticsDefaults(stats *UserStatistics) {
	if stats.OverallRating == 0 {
		stats.OverallRating = DefaultOverallRating
	}
	if stats.Last10Average == 0 {
		stats.Last10Average = stats.OverallRating
	}
	if stats.AvgResponseTimeMinutes == 0 {
		stats.AvgResponseTimeMinutes = DefaultResponseTime
	}
	if stats.PunctualityRate == 0 {
		stats.PunctualityRate = DefaultPunctualityRate
	}
	if stats.CompletionRate == 0 {
		stats.CompletionRate = DefaultCompletionRate
	}
	if stats.CancelRate == 0 {
		stats.CancelRate = DefaultCancelRate
	}
	if stats.ReliabilityScore == 0 {
		stats.ReliabilityScore = DefaultReliabilityScore
	}
}

// SetPricingDefaults sets default values for missing pricing data
func SetPricingDefaults(pricing *UserPricingStats) {
	if pricing.AveragePrice == 0 {
		pricing.AveragePrice = DefaultAveragePrice
	}
	if pricing.MinPrice == 0 {
		pricing.MinPrice = DefaultMinPrice
	}
	if pricing.MaxPrice == 0 {
		pricing.MaxPrice = DefaultMaxPrice
	}
}

// AnonymizeReviewerName anonymizes reviewer names for privacy
func AnonymizeReviewerName(fullName string) string {
	if len(fullName) == 0 {
		return "Usuario Anónimo"
	}
	
	names := strings.Split(fullName, " ")
	if len(names) == 1 {
		if len(names[0]) > 1 {
			return string(names[0][0]) + "."
		}
		return "Usuario Anónimo"
	}
	
	// Return first name + last name initial
	firstName := names[0]
	lastInitial := string(names[len(names)-1][0])
	return firstName + " " + lastInitial + "."
}