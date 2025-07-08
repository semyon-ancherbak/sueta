package telegram

import (
	"context"
	"testing"
	"time"

	"github.com/semyon-ancherbak/sueta/internal/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockRepository для тестирования
type MockRepository struct {
	mock.Mock
}

func (m *MockRepository) SaveChat(ctx context.Context, chat *models.ChatDocument) error {
	args := m.Called(ctx, chat)
	return args.Error(0)
}

func (m *MockRepository) SaveMessage(ctx context.Context, message *models.MessageDocument) error {
	args := m.Called(ctx, message)
	return args.Error(0)
}

func (m *MockRepository) ChatExists(ctx context.Context, chatID int64) (bool, error) {
	args := m.Called(ctx, chatID)
	return args.Bool(0), args.Error(1)
}

func (m *MockRepository) GetRecentMessages(ctx context.Context, chatID int64, days int) ([]*models.MessageDocument, error) {
	args := m.Called(ctx, chatID, days)
	return args.Get(0).([]*models.MessageDocument), args.Error(1)
}

func (m *MockRepository) Close(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

func TestClient_saveBotMessage(t *testing.T) {
	mockRepo := new(MockRepository)
	client := &Client{
		token:   "test_token",
		baseURL: "https://api.telegram.org/bottest_token",
		repo:    mockRepo,
	}

	// Подготавливаем тестовые данные
	msg := &Message{
		MessageID: 123,
		From: &User{
			ID:        987654321,
			IsBot:     true,
			FirstName: "TestBot",
			Username:  "testbot",
		},
		Chat: &Chat{
			ID:   123456789,
			Type: "private",
		},
		Date: time.Now().Unix(),
	}

	text := "Тестовое сообщение от бота"

	// Настраиваем ожидания для мока
	mockRepo.On("SaveMessage", mock.Anything, mock.MatchedBy(func(msgDoc *models.MessageDocument) bool {
		return msgDoc.MessageID == 123 &&
			msgDoc.ChatID == 123456789 &&
			msgDoc.Text == text &&
			msgDoc.IsBot == true &&
			msgDoc.UserID == 987654321
	})).Return(nil)

	// Вызываем тестируемый метод
	ctx := context.Background()
	err := client.saveBotMessage(ctx, msg, text)

	// Проверяем результат
	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)
}

func TestClient_saveBotMessage_NilMessage(t *testing.T) {
	mockRepo := new(MockRepository)
	client := &Client{
		token:   "test_token",
		baseURL: "https://api.telegram.org/bottest_token",
		repo:    mockRepo,
	}

	// Тестируем с nil сообщением
	ctx := context.Background()
	err := client.saveBotMessage(ctx, nil, "текст")

	// Должно пройти без ошибок, но ничего не сохранить
	assert.NoError(t, err)
	mockRepo.AssertNotCalled(t, "SaveMessage")
}

func TestClient_saveBotMessage_NilRepo(t *testing.T) {
	client := &Client{
		token:   "test_token",
		baseURL: "https://api.telegram.org/bottest_token",
		repo:    nil, // nil репозиторий
	}

	msg := &Message{
		MessageID: 123,
		Chat:      &Chat{ID: 123456789},
		Date:      time.Now().Unix(),
	}

	// Тестируем с nil репозиторием
	ctx := context.Background()
	err := client.saveBotMessage(ctx, msg, "текст")

	// Должно пройти без ошибок
	assert.NoError(t, err)
}
