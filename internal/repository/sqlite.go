package repository

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	_ "github.com/mattn/go-sqlite3"
	"github.com/semyon-ancherbak/sueta/internal/models"
)

type Repository interface {
	SaveChat(ctx context.Context, chat *models.ChatDocument) error
	SaveMessage(ctx context.Context, message *models.MessageDocument) error
	ChatExists(ctx context.Context, chatID int64) (bool, error)
	UpdateExists(ctx context.Context, updateID int) (bool, error)
	GetRecentMessages(ctx context.Context, chatID int64, days int) ([]*models.MessageDocument, error)
	GetLastMessages(ctx context.Context, chatID int64, limit int) ([]*models.MessageDocument, error)
	Close(ctx context.Context) error
}

// SQLiteRepository реализация Repository для SQLite
type SQLiteRepository struct {
	db *sql.DB
}

func NewSQLiteRepository(dbPath string) (*SQLiteRepository, error) {
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, fmt.Errorf("ошибка открытия базы данных: %w", err)
	}

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("ошибка подключения к базе данных: %w", err)
	}

	repo := &SQLiteRepository{db: db}

	if err := repo.createTables(); err != nil {
		return nil, fmt.Errorf("ошибка создания таблиц: %w", err)
	}

	return repo, nil
}

// createTables создает необходимые таблицы
func (r *SQLiteRepository) createTables() error {
	// Создаем таблицу chats
	chatTableSQL := `
	CREATE TABLE IF NOT EXISTS chats (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		chat_id INTEGER UNIQUE NOT NULL,
		type TEXT NOT NULL,
		title TEXT,
		username TEXT,
		first_name TEXT,
		last_name TEXT,
		created_at DATETIME NOT NULL,
		updated_at DATETIME NOT NULL
	);`

	if _, err := r.db.Exec(chatTableSQL); err != nil {
		return fmt.Errorf("ошибка создания таблицы chats: %w", err)
	}

	// Создаем таблицу messages
	messageTableSQL := `
	CREATE TABLE IF NOT EXISTS messages (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		message_id INTEGER NOT NULL,
		chat_id INTEGER NOT NULL,
		user_id INTEGER,
		username TEXT,
		first_name TEXT,
		last_name TEXT,
		text TEXT,
		date DATETIME NOT NULL,
		update_id INTEGER UNIQUE NOT NULL,
		is_bot BOOLEAN NOT NULL DEFAULT 0,
		is_addressed_to_bot BOOLEAN NOT NULL DEFAULT 0,
		created_at DATETIME NOT NULL,
		UNIQUE(message_id, chat_id)
	);`

	if _, err := r.db.Exec(messageTableSQL); err != nil {
		return fmt.Errorf("ошибка создания таблицы messages: %w", err)
	}

	// Создаем индексы
	indexes := []string{
		"CREATE INDEX IF NOT EXISTS idx_chats_chat_id ON chats(chat_id);",
		"CREATE INDEX IF NOT EXISTS idx_messages_chat_id ON messages(chat_id);",
		"CREATE INDEX IF NOT EXISTS idx_messages_update_id ON messages(update_id);",
		"CREATE INDEX IF NOT EXISTS idx_messages_date ON messages(date);",
		"CREATE INDEX IF NOT EXISTS idx_messages_text ON messages(text);",
	}

	for _, indexSQL := range indexes {
		if _, err := r.db.Exec(indexSQL); err != nil {
			return fmt.Errorf("ошибка создания индекса: %w", err)
		}
	}

	return nil
}

func (r *SQLiteRepository) SaveChat(ctx context.Context, chat *models.ChatDocument) error {
	now := time.Now()

	// Используем INSERT OR REPLACE для избежания дублирования
	query := `
	INSERT OR REPLACE INTO chats (
		chat_id, type, title, username, first_name, last_name, 
		created_at, updated_at
	) VALUES (?, ?, ?, ?, ?, ?, 
		COALESCE((SELECT created_at FROM chats WHERE chat_id = ?), ?), ?)`

	_, err := r.db.ExecContext(ctx, query,
		chat.ChatID, chat.Type, chat.Title, chat.Username, chat.FirstName, chat.LastName,
		chat.ChatID, now, now)

	return err
}

func (r *SQLiteRepository) SaveMessage(ctx context.Context, message *models.MessageDocument) error {
	now := time.Now()

	// Используем INSERT OR IGNORE для избежания дублирования сообщений
	query := `
	INSERT OR IGNORE INTO messages (
		message_id, chat_id, user_id, username, first_name, last_name,
		text, date, update_id, is_bot, is_addressed_to_bot, created_at
	) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`

	_, err := r.db.ExecContext(ctx, query,
		message.MessageID, message.ChatID, message.UserID, message.Username,
		message.FirstName, message.LastName, message.Text, message.Date,
		message.UpdateID, message.IsBot, message.IsAddressedToBot, now)

	return err
}

func (r *SQLiteRepository) ChatExists(ctx context.Context, chatID int64) (bool, error) {
	query := "SELECT COUNT(*) FROM chats WHERE chat_id = ?"
	var count int
	err := r.db.QueryRowContext(ctx, query, chatID).Scan(&count)
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

func (r *SQLiteRepository) UpdateExists(ctx context.Context, updateID int) (bool, error) {
	query := "SELECT COUNT(*) FROM messages WHERE update_id = ?"
	var count int
	err := r.db.QueryRowContext(ctx, query, updateID).Scan(&count)
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

func (r *SQLiteRepository) GetRecentMessages(
	ctx context.Context,
	chatID int64,
	days int,
) ([]*models.MessageDocument, error) {
	since := time.Now().AddDate(0, 0, -days)

	query := `
	SELECT id, message_id, chat_id, user_id, username, first_name, last_name,
		   text, date, update_id, is_bot, is_addressed_to_bot, created_at
	FROM messages 
	WHERE chat_id = ? AND date >= ?
	ORDER BY date ASC`

	rows, err := r.db.QueryContext(ctx, query, chatID, since)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var messages []*models.MessageDocument
	for rows.Next() {
		msg := &models.MessageDocument{}
		err := rows.Scan(
			&msg.ID, &msg.MessageID, &msg.ChatID, &msg.UserID, &msg.Username,
			&msg.FirstName, &msg.LastName, &msg.Text, &msg.Date, &msg.UpdateID,
			&msg.IsBot, &msg.IsAddressedToBot, &msg.CreatedAt,
		)
		if err != nil {
			return nil, err
		}
		messages = append(messages, msg)
	}

	return messages, rows.Err()
}

func (r *SQLiteRepository) GetLastMessages(
	ctx context.Context,
	chatID int64,
	limit int,
) ([]*models.MessageDocument, error) {
	query := `
	SELECT id, message_id, chat_id, user_id, username, first_name, last_name,
		   text, date, update_id, is_bot, is_addressed_to_bot, created_at
	FROM messages 
	WHERE chat_id = ?
	ORDER BY date DESC
	LIMIT ?`

	rows, err := r.db.QueryContext(ctx, query, chatID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var messages []*models.MessageDocument
	for rows.Next() {
		msg := &models.MessageDocument{}
		err := rows.Scan(
			&msg.ID, &msg.MessageID, &msg.ChatID, &msg.UserID, &msg.Username,
			&msg.FirstName, &msg.LastName, &msg.Text, &msg.Date, &msg.UpdateID,
			&msg.IsBot, &msg.IsAddressedToBot, &msg.CreatedAt,
		)
		if err != nil {
			return nil, err
		}
		messages = append(messages, msg)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	// Поскольку мы получили сообщения в убывающем порядке,
	// нужно развернуть их обратно для правильной хронологии
	for i, j := 0, len(messages)-1; i < j; i, j = i+1, j-1 {
		messages[i], messages[j] = messages[j], messages[i]
	}

	return messages, nil
}

func (r *SQLiteRepository) Close(ctx context.Context) error {
	return r.db.Close()
}
