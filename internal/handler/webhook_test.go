package handler

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestHandleWebhook(t *testing.T) {
	// Создаем handler
	handler := NewWebhookHandler()

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
	handler := NewWebhookHandler()

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
