package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// ChatDocument представляет документ чата в MongoDB
type ChatDocument struct {
	ID        primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	ChatID    int64              `bson:"chat_id" json:"chat_id"`
	Type      string             `bson:"type" json:"type"`
	Title     string             `bson:"title,omitempty" json:"title,omitempty"`
	Username  string             `bson:"username,omitempty" json:"username,omitempty"`
	FirstName string             `bson:"first_name,omitempty" json:"first_name,omitempty"`
	LastName  string             `bson:"last_name,omitempty" json:"last_name,omitempty"`
	CreatedAt time.Time          `bson:"created_at" json:"created_at"`
	UpdatedAt time.Time          `bson:"updated_at" json:"updated_at"`
}

// MessageDocument представляет документ сообщения в MongoDB
type MessageDocument struct {
	ID        primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	MessageID int                `bson:"message_id" json:"message_id"`
	ChatID    int64              `bson:"chat_id" json:"chat_id"`
	UserID    int64              `bson:"user_id,omitempty" json:"user_id,omitempty"`
	Username  string             `bson:"username,omitempty" json:"username,omitempty"`
	FirstName string             `bson:"first_name,omitempty" json:"first_name,omitempty"`
	LastName  string             `bson:"last_name,omitempty" json:"last_name,omitempty"`
	Text      string             `bson:"text,omitempty" json:"text,omitempty"`
	Date      time.Time          `bson:"date" json:"date"`
	UpdateID  int                `bson:"update_id" json:"update_id"`
	IsBot     bool               `bson:"is_bot" json:"is_bot"`
	CreatedAt time.Time          `bson:"created_at" json:"created_at"`
}
