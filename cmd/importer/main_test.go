package main

import (
	"testing"
	"time"

	"github.com/semyon-ancherbak/sueta/internal/models"
	"github.com/stretchr/testify/assert"
)

func TestTelegramImporter_parseMessage(t *testing.T) {
	importer := &TelegramImporter{
		verbose: true,
	}

	tests := []struct {
		name        string
		input       *models.TelegramExportMessage
		chatID      int64
		expected    *models.ParsedMessage
		expectError bool
	}{
		{
			name: "Обычное сообщение пользователя",
			input: &models.TelegramExportMessage{
				ID:           123,
				Type:         "message",
				Date:         "2021-05-09T12:07:03",
				From:         "Тестовый пользователь",
				FromID:       "user262343524",
				Text:         "Привет, как дела?",
				TextEntities: []models.TextEntity{},
			},
			chatID: 1698320716,
			expected: &models.ParsedMessage{
				OriginalID:  123,
				ChatID:      1698320716,
				UserID:      262343524,
				Username:    "Тестовый пользователь",
				FirstName:   "Тестовый пользователь",
				Text:        "Привет, как дела?",
				MessageType: "message",
				IsService:   false,
			},
			expectError: false,
		},
		{
			name: "Служебное сообщение",
			input: &models.TelegramExportMessage{
				ID:           124,
				Type:         "service",
				Date:         "2021-05-09T12:07:04",
				Actor:        "Админ",
				ActorID:      "user123456789",
				Action:       "create_group",
				Text:         "",
				TextEntities: []models.TextEntity{},
			},
			chatID: 1698320716,
			expected: &models.ParsedMessage{
				OriginalID:  124,
				ChatID:      1698320716,
				UserID:      123456789,
				Username:    "Админ",
				FirstName:   "Админ",
				Text:        "",
				MessageType: "service",
				IsService:   true,
			},
			expectError: false,
		},
		{
			name: "Сообщение с массивом текста",
			input: &models.TelegramExportMessage{
				ID:     125,
				Type:   "message",
				Date:   "2021-05-09T12:07:05",
				From:   "Пользователь",
				FromID: "user987654321",
				Text: []interface{}{
					map[string]interface{}{"type": "plain", "text": "Часть 1 "},
					map[string]interface{}{"type": "bold", "text": "жирный текст"},
					map[string]interface{}{"type": "plain", "text": " часть 3"},
				},
				TextEntities: []models.TextEntity{},
			},
			chatID: 1698320716,
			expected: &models.ParsedMessage{
				OriginalID:  125,
				ChatID:      1698320716,
				UserID:      987654321,
				Username:    "Пользователь",
				FirstName:   "Пользователь",
				Text:        "Часть 1 жирный текст часть 3",
				MessageType: "message",
				IsService:   false,
			},
			expectError: false,
		},
		{
			name: "Ошибка парсинга даты",
			input: &models.TelegramExportMessage{
				ID:           126,
				Type:         "message",
				Date:         "invalid-date",
				From:         "Пользователь",
				Text:         "Текст",
				TextEntities: []models.TextEntity{},
			},
			chatID:      1698320716,
			expected:    nil,
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := importer.parseMessage(tt.input, tt.chatID)

			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)

				if tt.expected != nil {
					assert.Equal(t, tt.expected.OriginalID, result.OriginalID)
					assert.Equal(t, tt.expected.ChatID, result.ChatID)
					assert.Equal(t, tt.expected.UserID, result.UserID)
					assert.Equal(t, tt.expected.Username, result.Username)
					assert.Equal(t, tt.expected.FirstName, result.FirstName)
					assert.Equal(t, tt.expected.Text, result.Text)
					assert.Equal(t, tt.expected.MessageType, result.MessageType)
					assert.Equal(t, tt.expected.IsService, result.IsService)

					// Проверяем, что дата корректно распарсена
					expectedDate, _ := time.Parse("2006-01-02T15:04:05", tt.input.Date)
					assert.Equal(t, expectedDate, result.Date)
				}
			}
		})
	}
}

func TestExtractTextFromMessage(t *testing.T) {
	tests := []struct {
		name     string
		input    interface{}
		expected string
	}{
		{
			name:     "Простая строка",
			input:    "Простой текст сообщения",
			expected: "Простой текст сообщения",
		},
		{
			name: "Массив текстовых объектов",
			input: []interface{}{
				map[string]interface{}{"type": "plain", "text": "Начало "},
				map[string]interface{}{"type": "bold", "text": "жирный"},
				map[string]interface{}{"type": "plain", "text": " конец"},
			},
			expected: "Начало жирный конец",
		},
		{
			name:     "Пустая строка",
			input:    "",
			expected: "",
		},
		{
			name:     "Неожиданный тип",
			input:    123,
			expected: "",
		},
		{
			name:     "Пустой массив",
			input:    []interface{}{},
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			msg := &models.TelegramExportMessage{
				Text: tt.input,
			}
			result := msg.ExtractTextFromMessage()
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestExtractUserIDFromActorID(t *testing.T) {
	tests := []struct {
		name     string
		actorID  string
		fromID   string
		expected int64
	}{
		{
			name:     "ActorID с префиксом user",
			actorID:  "user262343524",
			fromID:   "",
			expected: 262343524,
		},
		{
			name:     "FromID с префиксом user",
			actorID:  "",
			fromID:   "user123456789",
			expected: 123456789,
		},
		{
			name:     "ActorID без префикса",
			actorID:  "987654321",
			fromID:   "",
			expected: 987654321,
		},
		{
			name:     "Пустые значения",
			actorID:  "",
			fromID:   "",
			expected: 0,
		},
		{
			name:     "Некорректный ID",
			actorID:  "invalid",
			fromID:   "",
			expected: 0,
		},
		{
			name:     "Приоритет ActorID",
			actorID:  "user111",
			fromID:   "user222",
			expected: 111,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			msg := &models.TelegramExportMessage{
				ActorID: tt.actorID,
				FromID:  tt.fromID,
			}
			result := msg.ExtractUserIDFromActorID()
			assert.Equal(t, tt.expected, result)
		})
	}
}
