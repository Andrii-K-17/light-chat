package models

import "time"

type Chat struct {
	ID        int       `db:"id"         json:"id"`
	Name      *string   `db:"name"       json:"name"`
	IsGroup   bool      `db:"is_group"   json:"is_group"`
	CreatedBy *int      `db:"created_by" json:"created_by"`
	CreatedAt time.Time `db:"created_at" json:"created_at"`
}

type ChatResponse struct {
	Chat
	Members     []ChatMember `json:"members"`
	LastMessage *Message     `json:"last_message"`
	UnreadCount int          `json:"unread_count"`
}

type ChatMember struct {
	ID          int    `db:"id"           json:"id"`
	Username    string `db:"username"     json:"username"`
	DisplayName string `db:"display_name" json:"display_name"`
	Status      string `db:"status"       json:"status"`
}
