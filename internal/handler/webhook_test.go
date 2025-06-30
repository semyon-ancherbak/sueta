package handler

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/semyon-ancherbak/sueta/internal/models"
)

// MockRepository реализует интерфейс Repository для тестов
type MockRepository struct {
	chats    map[int64]bool
	messages []models.MessageDocument
}

func NewMockRepository() *MockRepository {
	return &MockRepository{
		chats:    make(map[int64]bool),
		messages: make([]models.MessageDocument, 0),
	}
}

func (m *MockRepository) SaveChat(ctx context.Context, chat *models.ChatDocument) error {
	m.chats[chat.ChatID] = true
	return nil
}

func (m *MockRepository) SaveMessage(ctx context.Context, message *models.MessageDocument) error {
	m.messages = append(m.messages, *message)
	return nil
}

func (m *MockRepository) ChatExists(ctx context.Context, chatID int64) (bool, error) {
	return m.chats[chatID], nil
}

func (m *MockRepository) Close(ctx context.Context) error {
	return nil
}

func TestHandleWebhook(t *testing.T) {
	// Создаем mock repository
	mockRepo := NewMockRepository()

	// Создаем handler
	handler := NewWebhookHandler(mockRepo)

	// Подготавливаем тестовые данные
	testUpdate := `{
		"update_id": 123456789,
		"message": {
			"message_id": 1,
			"from": {
				"id": 123456,
				"is_bot": false,
				"first_name": "Test",
				"last_name": "User",
				"username": "testuser"
			},
			"chat": {
				"id": 123456,
				"type": "private"
			},
			"date": 1640995200,
			"text": "Привет, бот!"
		}
	}`

	// Создаем HTTP запрос
	req, err := http.NewRequest("POST", "/webhook", bytes.NewBufferString(testUpdate))
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Content-Type", "application/json")

	// Создаем recorder для захвата ответа
	rr := httptest.NewRecorder()

	// Вызываем handler
	handler.HandleWebhook(rr, req)

	// Проверяем статус код
	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, http.StatusOK)
	}

	// Проверяем тело ответа
	expected := "OK"
	if rr.Body.String() != expected {
		t.Errorf("handler returned unexpected body: got %v want %v",
			rr.Body.String(), expected)
	}
}

func TestHandleWebhookInvalidJSON(t *testing.T) {
	// Создаем mock repository
	mockRepo := NewMockRepository()

	handler := NewWebhookHandler(mockRepo)

	// Невалидный JSON
	invalidJSON := `{"update_id": 123, "message": {invalid}`

	req, err := http.NewRequest("POST", "/webhook", bytes.NewBufferString(invalidJSON))
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Content-Type", "application/json")

	rr := httptest.NewRecorder()
	handler.HandleWebhook(rr, req)

	// Ожидаем ошибку 400
	if status := rr.Code; status != http.StatusBadRequest {
		t.Errorf("handler returned wrong status code for invalid JSON: got %v want %v",
			status, http.StatusBadRequest)
	}
}

func TestChatNotDuplicatedOnMultipleMessages(t *testing.T) {
	// Создаем mock repository
	mockRepo := NewMockRepository()
	handler := NewWebhookHandler(mockRepo)

	// Первое сообщение
	testUpdate1 := `{
		"update_id": 123456789,
		"message": {
			"message_id": 1,
			"from": {
				"id": 123456,
				"is_bot": false,
				"first_name": "Test",
				"username": "testuser"
			},
			"chat": {
				"id": 123456,
				"type": "private"
			},
			"date": 1640995200,
			"text": "Первое сообщение"
		}
	}`

	// Второе сообщение в том же чате
	testUpdate2 := `{
		"update_id": 123456790,
		"message": {
			"message_id": 2,
			"from": {
				"id": 123456,
				"is_bot": false,
				"first_name": "Test",
				"username": "testuser"
			},
			"chat": {
				"id": 123456,
				"type": "private"
			},
			"date": 1640995201,
			"text": "Второе сообщение"
		}
	}`

	// Отправляем первое сообщение
	req1, _ := http.NewRequest("POST", "/webhook", bytes.NewBufferString(testUpdate1))
	req1.Header.Set("Content-Type", "application/json")
	rr1 := httptest.NewRecorder()
	handler.HandleWebhook(rr1, req1)

	// Отправляем второе сообщение
	req2, _ := http.NewRequest("POST", "/webhook", bytes.NewBufferString(testUpdate2))
	req2.Header.Set("Content-Type", "application/json")
	rr2 := httptest.NewRecorder()
	handler.HandleWebhook(rr2, req2)

	// Проверяем, что оба запроса успешны
	if rr1.Code != http.StatusOK || rr2.Code != http.StatusOK {
		t.Errorf("Expected status OK for both requests, got %d and %d", rr1.Code, rr2.Code)
	}

	// Проверяем, что в репозитории есть только одна запись о чате
	if len(mockRepo.chats) != 1 {
		t.Errorf("Expected 1 chat in repository, got %d", len(mockRepo.chats))
	}

	// Проверяем, что сохранены 2 сообщения
	if len(mockRepo.messages) != 2 {
		t.Errorf("Expected 2 messages in repository, got %d", len(mockRepo.messages))
	}
}
