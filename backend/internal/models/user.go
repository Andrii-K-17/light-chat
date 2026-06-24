package models

import "time"

type User struct {
	ID           int       `db:"id"            json:"id"`
	Email        string    `db:"email"         json:"email"`
	Username     string    `db:"username"      json:"username"`
	DisplayName  string    `db:"display_name"  json:"display_name"`
	PasswordHash string    `db:"password_hash" json:"-"`
	Status       string    `db:"status"        json:"status"`
	CreatedAt    time.Time `db:"created_at"    json:"created_at"`
}
