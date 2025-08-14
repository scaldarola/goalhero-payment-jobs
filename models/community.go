package models

import "time"

// CommunityStats represents overall platform statistics
type CommunityStats struct {
	TotalPlayers        int     `json:"totalPlayers"`
	TotalMatches        int     `json:"totalMatches"`
	AverageRating       float64 `json:"averageRating"`
	ActivePlayersToday  int     `json:"activePlayersToday"`
}

// NewsItem represents a community news/update item
type NewsItem struct {
	ID          string       `json:"id"`
	Type        string       `json:"type"`        // promo, matches, announcement, maintenance
	Title       string       `json:"title"`
	Description string       `json:"description"`
	Icon        string       `json:"icon"`        // Material Icon name
	Color       string       `json:"color"`       // UI color theme
	Priority    int          `json:"priority"`    // Display order (1 = highest)
	ExpiresAt   time.Time    `json:"expiresAt"`
	GeoLocation *GeoLocation `json:"geoLocation,omitempty"` // Optional location for filtering
}

// GeoLocation represents geographical coordinates for filtering
type GeoLocation struct {
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
	Radius    float64 `json:"radius"` // in kilometers
}

// NewsResponse represents the response for community news
type NewsResponse struct {
	News []NewsItem `json:"news"`
}

// EventParticipants represents event participation details
type EventParticipants struct {
	Current int `json:"current"`
	Maximum int `json:"maximum"`
}

// CommunityEvent represents a special event or tournament
type CommunityEvent struct {
	ID           string            `json:"id"`
	Title        string            `json:"title"`
	Description  string            `json:"description"`
	StartDate    time.Time         `json:"startDate"`
	EndDate      time.Time         `json:"endDate"`
	Participants EventParticipants `json:"participants"`
	Color        string            `json:"color"`   // UI theme color
	Status       string            `json:"status"`  // open, full, closed, cancelled, completed
	Location     string            `json:"location"`
	EntryFee     float64           `json:"entryFee"` // in euros
	Prize        string            `json:"prize"`    // Description of prize
}

// EventsResponse represents the response for community events
type EventsResponse struct {
	Events []CommunityEvent `json:"events"`
}

// CommunityAchievement represents a community milestone or player accomplishment
type CommunityAchievement struct {
	ID           string    `json:"id"`
	Type         string    `json:"type"`         // milestone, rating, streak, tournament
	Title        string    `json:"title"`
	Description  string    `json:"description"`
	PlayerName   string    `json:"playerName"`   // Display name (abbreviated)
	PlayerAvatar string    `json:"playerAvatar,omitempty"` // Profile picture URL
	Icon         string    `json:"icon"`         // Material Icon name
	Color        string    `json:"color"`        // UI color theme
	AchievedAt   time.Time `json:"achievedAt"`
	Value        float64   `json:"value"`        // Numeric value of achievement
	Unit         string    `json:"unit"`         // Unit of measurement
}

// AchievementsResponse represents the response for community achievements
type AchievementsResponse struct {
	Achievements []CommunityAchievement `json:"achievements"`
}