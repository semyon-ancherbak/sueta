package models

import (
	"encoding/json"
	"fmt"
	"strconv"
	"time"
)

// TelegramExport представляет структуру экспорта Telegram
type TelegramExport struct {
	Name     string                  `json:"name"`
	Type     string                  `json:"type"`
	ID       int64                   `json:"id"`
	Messages []TelegramExportMessage `json:"messages"`
}

// TelegramExportMessage представляет сообщение из экспорта Telegram
type TelegramExportMessage struct {
	ID               int64        `json:"id"`
	Type             string       `json:"type"`
	Date             string       `json:"date"`
	DateUnixtime     string       `json:"date_unixtime"`
	From             string       `json:"from,omitempty"`
	FromID           string       `json:"from_id,omitempty"`
	Actor            string       `json:"actor,omitempty"`
	ActorID          string       `json:"actor_id,omitempty"`
	Action           string       `json:"action,omitempty"`
	Title            string       `json:"title,omitempty"`
	Members          []string     `json:"members,omitempty"`
	Text             interface{}  `json:"text"` // Может быть строкой или массивом объектов
	TextEntities     []TextEntity `json:"text_entities"`
	Photo            string       `json:"photo,omitempty"`
	ReplyToMessageID int64        `json:"reply_to_message_id,omitempty"`
	ForwardedFrom    string       `json:"forwarded_from,omitempty"`
	File             string       `json:"file,omitempty"`
	MediaType        string       `json:"media_type,omitempty"`
	MimeType         string       `json:"mime_type,omitempty"`
	Duration         int          `json:"duration,omitempty"`
	Width            int          `json:"width,omitempty"`
	Height           int          `json:"height,omitempty"`
}

// FlexibleUserID может быть строкой или числом в JSON
type FlexibleUserID string

// UnmarshalJSON кастомный unmarshaler для обработки строк и чисел
func (f *FlexibleUserID) UnmarshalJSON(data []byte) error {
	// Пробуем как строку
	var s string
	if err := json.Unmarshal(data, &s); err == nil {
		*f = FlexibleUserID(s)
		return nil
	}

	// Пробуем как число
	var n int64
	if err := json.Unmarshal(data, &n); err == nil {
		*f = FlexibleUserID(strconv.FormatInt(n, 10))
		return nil
	}

	return fmt.Errorf("user_id должен быть строкой или числом")
}

// String возвращает строковое представление
func (f FlexibleUserID) String() string {
	return string(f)
}

// TextEntity представляет сущность текста (ссылки, упоминания и т.д.)
type TextEntity struct {
	Type   string         `json:"type"`
	Text   string         `json:"text"`
	Href   string         `json:"href,omitempty"`
	UserID FlexibleUserID `json:"user_id,omitempty"`
}

// ParsedMessage содержит обработанную информацию из экспорта
type ParsedMessage struct {
	OriginalID  int64
	ChatID      int64
	UserID      int64
	Username    string
	FirstName   string
	Text        string
	Date        time.Time
	MessageType string
	IsService   bool
}

// ExtractTextFromMessage извлекает текст из поля text (строка или массив объектов)
func (msg *TelegramExportMessage) ExtractTextFromMessage() string {
	switch v := msg.Text.(type) {
	case string:
		return v
	case []interface{}:
		var result string
		for _, item := range v {
			if textObj, ok := item.(map[string]interface{}); ok {
				if text, exists := textObj["text"]; exists {
					if textStr, ok := text.(string); ok {
						result += textStr
					}
				}
			}
		}
		return result
	default:
		return ""
	}
}

// ExtractUserIDFromActorID извлекает числовой ID пользователя из строки actor_id
func (msg *TelegramExportMessage) ExtractUserIDFromActorID() int64 {
	if msg.ActorID == "" && msg.FromID == "" {
		return 0
	}

	idStr := msg.ActorID
	if idStr == "" {
		idStr = msg.FromID
	}

	// Удаляем префикс "user" если есть
	if len(idStr) > 4 && idStr[:4] == "user" {
		idStr = idStr[4:]
	}

	// Пробуем преобразовать в int64
	var userID int64
	if _, err := fmt.Sscanf(idStr, "%d", &userID); err != nil {
		return 0
	}

	return userID
}
