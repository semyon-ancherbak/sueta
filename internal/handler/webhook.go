package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/semyon-ancherbak/sueta/internal/config"
	"github.com/semyon-ancherbak/sueta/internal/llm"
	"github.com/semyon-ancherbak/sueta/internal/models"
	"github.com/semyon-ancherbak/sueta/internal/repository"
	"github.com/semyon-ancherbak/sueta/internal/telegram"
)

type TelegramUpdate struct {
	UpdateID int      `json:"update_id"`
	Message  *Message `json:"message,omitempty"`
}

type Message struct {
	MessageID      int      `json:"message_id"`
	From           *User    `json:"from,omitempty"`
	Chat           *Chat    `json:"chat,omitempty"`
	Date           int64    `json:"date"`
	Text           string   `json:"text,omitempty"`
	ReplyToMessage *Message `json:"reply_to_message,omitempty"`
}

type User struct {
	ID        int64  `json:"id"`
	IsBot     bool   `json:"is_bot"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name,omitempty"`
	Username  string `json:"username,omitempty"`
}

type Chat struct {
	ID    int64  `json:"id"`
	Type  string `json:"type"`
	Title string `json:"title,omitempty"`
}

type WebhookHandler struct {
	repo      repository.Repository
	llmClient *llm.Client
	tgClient  *telegram.Client
	botName   string
	cfg       *config.Config
}

func NewWebhookHandler(
	repo repository.Repository,
	llmClient *llm.Client,
	tgClient *telegram.Client,
	botName string,
	config *config.Config,
) *WebhookHandler {
	return &WebhookHandler{
		repo:      repo,
		llmClient: llmClient,
		tgClient:  tgClient,
		botName:   botName,
		cfg:       config,
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
	r.Post("/webhook/{token}", h.HandleWebhook)

	return r
}

func (h *WebhookHandler) HandleWebhook(w http.ResponseWriter, r *http.Request) {
	token := chi.URLParam(r, "token")
	if token != h.cfg.TelegramToken {
		log.Printf("received invalid TG token")
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}

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

	h.processUpdate(&update)
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
}

func (h *WebhookHandler) processUpdate(update *TelegramUpdate) {
	if update.Message == nil {
		log.Printf("Получено обновление без сообщения: UpdateID=%d", update.UpdateID)
		return
	}

	msg := update.Message
	ctx := context.Background()

	if err := h.saveChat(ctx, msg.Chat, msg.From); err != nil {
		log.Printf("Ошибка сохранения чата: %v", err)
	}

	if err := h.saveMessage(ctx, update); err != nil {
		log.Printf("Ошибка сохранения сообщения: %v", err)
	}

	if h.isMessageForBot(msg) {
		log.Printf("Сообщение адресовано боту, обрабатываем через LLM")
		if err := h.handleBotMessage(ctx, msg); err != nil {
			log.Printf("Ошибка обработки сообщения через LLM: %v", err)
		}
	}

	h.printMessageInfo(update)
}

func (h *WebhookHandler) printMessageInfo(update *TelegramUpdate) {
	if update == nil || update.Message == nil {
		log.Printf("Получено обновление без сообщения")
		return
	}
	msg := update.Message

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
}

func (h *WebhookHandler) isMessageForBot(msg *Message) bool {
	if msg == nil {
		return false
	}
	if h.isReplyToBot(msg) {
		log.Printf("Сообщение является ответом на сообщение бота")
		return true
	}
	// Проверяем наличие склонений имени "жорик" (регистронезависимо)
	if h.containsBotName(msg.Text) {
		log.Printf("Сообщение содержит обращение к Жорику")
		return true
	}
	return false
}

func (h *WebhookHandler) isReplyToBot(msg *Message) bool {
	if msg.ReplyToMessage == nil || msg.ReplyToMessage.From == nil {
		return false
	}
	return msg.ReplyToMessage.From.IsBot
}

func (h *WebhookHandler) containsBotName(text string) bool {
	if text == "" {
		return false
	}
	lowerText := strings.ToLower(text)
	zhorikVariations := []string{
		"жорик",   // именительный падеж
		"жорика",  // родительный падеж
		"жорику",  // дательный падеж
		"жориком", // творительный падеж
		"жорике",  // предложный падеж
		"жора",    // сокращение
		"жорж",    // альтернативная форма
	}
	for _, variation := range zhorikVariations {
		if strings.Contains(lowerText, variation) {
			return true
		}
	}
	return false
}

func (h *WebhookHandler) handleBotMessage(ctx context.Context, msg *Message) error {
	// Получаем последние 100 сообщений из чата
	messages, err := h.repo.GetLastMessages(ctx, msg.Chat.ID, 100)
	if err != nil {
		return fmt.Errorf("ошибка получения сообщений: %w", err)
	}

	log.Printf("Найдено %d последних сообщений для контекста", len(messages))

	// Генерируем ответ с использованием простого контекста
	response, err := h.llmClient.GenerateResponse(ctx, messages, msg.Text)
	if err != nil {
		return fmt.Errorf("ошибка генерации ответа: %w", err)
	}

	log.Printf("LLM ответ: %s", response)

	if err := h.tgClient.SendMessage(ctx, msg.Chat.ID, response, msg.MessageID); err != nil {
		return fmt.Errorf("ошибка отправки сообщения в Telegram: %w", err)
	}
	log.Printf("Ответ отправлен в чат %d", msg.Chat.ID)

	return nil
}

func (h *WebhookHandler) saveChat(ctx context.Context, chat *Chat, user *User) error {
	if chat == nil {
		return nil
	}
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
