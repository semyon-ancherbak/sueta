package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/semyon-ancherbak/sueta/internal/models"
	"github.com/semyon-ancherbak/sueta/internal/repository"
)

// TelegramUpdate представляет обновление от Telegram API
type TelegramUpdate struct {
	UpdateID int      `json:"update_id"`
	Message  *Message `json:"message,omitempty"`
}

// Message представляет сообщение в Telegram
type Message struct {
	MessageID int    `json:"message_id"`
	From      *User  `json:"from,omitempty"`
	Chat      *Chat  `json:"chat,omitempty"`
	Date      int64  `json:"date"`
	Text      string `json:"text,omitempty"`
}

// User представляет пользователя Telegram
type User struct {
	ID        int64  `json:"id"`
	IsBot     bool   `json:"is_bot"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name,omitempty"`
	Username  string `json:"username,omitempty"`
}

// Chat представляет чат в Telegram
type Chat struct {
	ID    int64  `json:"id"`
	Type  string `json:"type"`
	Title string `json:"title,omitempty"`
}

// WebhookHandler обрабатывает входящие webhook от Telegram
type WebhookHandler struct {
	repo repository.Repository
}

// NewWebhookHandler создает новый экземпляр WebhookHandler
func NewWebhookHandler(repo repository.Repository) *WebhookHandler {
	return &WebhookHandler{
		repo: repo,
	}
}

// SetupRouter настраивает маршруты для webhook
func (h *WebhookHandler) SetupRouter() *chi.Mux {
	r := chi.NewRouter()

	// Middleware
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.Timeout(60 * time.Second))

	// Webhook endpoint
	r.Post("/webhook", h.HandleWebhook)

	return r
}

// HandleWebhook обрабатывает входящие обновления от Telegram
func (h *WebhookHandler) HandleWebhook(w http.ResponseWriter, r *http.Request) {
	// Читаем тело запроса
	body, err := io.ReadAll(r.Body)
	if err != nil {
		log.Printf("Ошибка чтения тела запроса: %v", err)
		http.Error(w, "Ошибка чтения запроса", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	// Парсим JSON
	var update TelegramUpdate
	if err := json.Unmarshal(body, &update); err != nil {
		log.Printf("Ошибка парсинга JSON: %v", err)
		http.Error(w, "Неверный формат JSON", http.StatusBadRequest)
		return
	}

	// Выводим информацию о сообщении в консоль
	h.processUpdate(&update)

	// Возвращаем успешный ответ
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
}

// processUpdate обрабатывает полученное обновление
func (h *WebhookHandler) processUpdate(update *TelegramUpdate) {
	if update.Message == nil {
		log.Printf("Получено обновление без сообщения: UpdateID=%d", update.UpdateID)
		return
	}

	msg := update.Message
	ctx := context.Background()

	// Сохраняем чат (если он еще не существует)
	if err := h.saveChat(ctx, msg.Chat, msg.From); err != nil {
		log.Printf("Ошибка сохранения чата: %v", err)
	}

	// Сохраняем сообщение
	if err := h.saveMessage(ctx, update); err != nil {
		log.Printf("Ошибка сохранения сообщения: %v", err)
	}

	// Форматированный вывод информации о сообщении
	fmt.Printf("\n=== НОВОЕ СООБЩЕНИЕ ===\n")
	fmt.Printf("Update ID: %d\n", update.UpdateID)
	fmt.Printf("Message ID: %d\n", msg.MessageID)
	fmt.Printf("Дата: %s\n", time.Unix(msg.Date, 0).Format("2006-01-02 15:04:05"))

	if msg.From != nil {
		fmt.Printf("От пользователя: %s", msg.From.FirstName)
		if msg.From.LastName != "" {
			fmt.Printf(" %s", msg.From.LastName)
		}
		if msg.From.Username != "" {
			fmt.Printf(" (@%s)", msg.From.Username)
		}
		fmt.Printf(" [ID: %d]\n", msg.From.ID)
	}

	if msg.Chat != nil {
		fmt.Printf("Чат: %s [ID: %d, Type: %s]\n", msg.Chat.Title, msg.Chat.ID, msg.Chat.Type)
	}

	if msg.Text != "" {
		fmt.Printf("Текст сообщения: %s\n", msg.Text)
	}

	fmt.Printf("========================\n\n")

	// Также логируем в стандартный лог
	log.Printf("Получено сообщение от %s: %s",
		getUserName(msg.From), msg.Text)
}

// saveChat сохраняет информацию о чате
func (h *WebhookHandler) saveChat(ctx context.Context, chat *Chat, user *User) error {
	if chat == nil {
		return nil
	}

	// Проверяем, существует ли чат
	exists, err := h.repo.ChatExists(ctx, chat.ID)
	if err != nil {
		return fmt.Errorf("ошибка проверки существования чата: %w", err)
	}

	if exists {
		log.Printf("Чат %d уже существует, пропускаем сохранение", chat.ID)
		return nil
	}

	chatDoc := &models.ChatDocument{
		ChatID: chat.ID,
		Type:   chat.Type,
		Title:  chat.Title,
	}

	// Для приватных чатов добавляем информацию о пользователе
	if chat.Type == "private" && user != nil {
		chatDoc.Username = user.Username
		chatDoc.FirstName = user.FirstName
		chatDoc.LastName = user.LastName
	}

	if err := h.repo.SaveChat(ctx, chatDoc); err != nil {
		return fmt.Errorf("ошибка сохранения чата: %w", err)
	}

	log.Printf("Сохранен новый чат: ID=%d, Type=%s", chat.ID, chat.Type)
	return nil
}

// saveMessage сохраняет сообщение
func (h *WebhookHandler) saveMessage(ctx context.Context, update *TelegramUpdate) error {
	msg := update.Message
	if msg == nil {
		return nil
	}

	messageDoc := &models.MessageDocument{
		MessageID: msg.MessageID,
		ChatID:    msg.Chat.ID,
		Text:      msg.Text,
		Date:      time.Unix(msg.Date, 0),
		UpdateID:  update.UpdateID,
	}

	// Добавляем информацию о пользователе
	if msg.From != nil {
		messageDoc.UserID = msg.From.ID
		messageDoc.Username = msg.From.Username
		messageDoc.FirstName = msg.From.FirstName
		messageDoc.LastName = msg.From.LastName
		messageDoc.IsBot = msg.From.IsBot
	}

	if err := h.repo.SaveMessage(ctx, messageDoc); err != nil {
		return fmt.Errorf("ошибка сохранения сообщения: %w", err)
	}

	log.Printf("Сохранено сообщение: ID=%d, ChatID=%d", msg.MessageID, msg.Chat.ID)
	return nil
}

// getUserName возвращает отформатированное имя пользователя
func getUserName(user *User) string {
	if user == nil {
		return "Unknown"
	}

	name := user.FirstName
	if user.LastName != "" {
		name += " " + user.LastName
	}
	if user.Username != "" {
		name += " (@" + user.Username + ")"
	}

	return name
}
