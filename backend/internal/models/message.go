package models

import "time"

type Message struct {
	ID        int       `db:"id"         json:"id"`
	ChatID    int       `db:"chat_id"    json:"chat_id"`
	UserID    int       `db:"user_id"    json:"user_id"`
	Content   string    `db:"content"    json:"content"`
	IsRead    bool      `db:"is_read"    json:"is_read"`
	CreatedAt time.Time `db:"created_at" json:"created_at"`
}

type MessageResponse struct {
	Message
	SenderUsername    string `db:"sender_username"    json:"sender_username"`
	SenderDisplayName string `db:"sender_display_name" json:"sender_display_name"`
}
