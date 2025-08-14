package models

import "time"

type Conversation struct {
	ID            string     `json:"id"`
	MatchID       string     `json:"matchId"`
	ApplicationID string     `json:"applicationId,omitempty"`
	Participants  []string   `json:"participants"`
	UserA         string     `json:"userA,omitempty"` // UID del primer usuario
	UserB         string     `json:"userB,omitempty"` // UID del segundo usuario
	CreatedAt     time.Time  `json:"createdAt"`
	Messages      []*Message `json:"messages,omitempty"`
	LastMessage   string     `json:"lastMessage,omitempty"` // Mensajes de la conversación
}

type Message struct {
	ID             string    `json:"id"`
	ConversationID string    `json:"conversationId"`
	SenderID       string    `json:"senderId"`
	Content        string    `json:"content"`
	CreatedAt      time.Time `json:"createdAt"`
	Read           bool      `json:"read"`
}
