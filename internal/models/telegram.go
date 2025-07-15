package models

import (
	"time"
)

// ChatDocument представляет запись чата в SQLite
type ChatDocument struct {
	ID        int64     `db:"id" json:"id"`
	ChatID    int64     `db:"chat_id" json:"chat_id"`
	Type      string    `db:"type" json:"type"`
	Title     string    `db:"title" json:"title"`
	Username  string    `db:"username" json:"username"`
	FirstName string    `db:"first_name" json:"first_name"`
	LastName  string    `db:"last_name" json:"last_name"`
	CreatedAt time.Time `db:"created_at" json:"created_at"`
	UpdatedAt time.Time `db:"updated_at" json:"updated_at"`
}

// MessageDocument представляет запись сообщения в SQLite
type MessageDocument struct {
	ID               int64     `db:"id" json:"id"`
	MessageID        int       `db:"message_id" json:"message_id"`
	ChatID           int64     `db:"chat_id" json:"chat_id"`
	UserID           int64     `db:"user_id" json:"user_id"`
	Username         string    `db:"username" json:"username"`
	FirstName        string    `db:"first_name" json:"first_name"`
	LastName         string    `db:"last_name" json:"last_name"`
	Text             string    `db:"text" json:"text"`
	Date             time.Time `db:"date" json:"date"`
	UpdateID         int       `db:"update_id" json:"update_id"`
	IsBot            bool      `db:"is_bot" json:"is_bot"`
	IsAddressedToBot bool      `db:"is_addressed_to_bot" json:"is_addressed_to_bot"`
	CreatedAt        time.Time `db:"created_at" json:"created_at"`
}
