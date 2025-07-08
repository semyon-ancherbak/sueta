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
	"github.com/semyon-ancherbak/sueta/internal/llm"
	"github.com/semyon-ancherbak/sueta/internal/models"
	"github.com/semyon-ancherbak/sueta/internal/repository"
	"github.com/semyon-ancherbak/sueta/internal/telegram"
)

// TelegramUpdate представляет обновление от Telegram API
type TelegramUpdate struct {
	UpdateID int      `json:"update_id"`
	Message  *Message `json:"message,omitempty"`
}

// Message представляет сообщение в Telegram
type Message struct {
	MessageID      int      `json:"message_id"`
	From           *User    `json:"from,omitempty"`
	Chat           *Chat    `json:"chat,omitempty"`
	Date           int64    `json:"date"`
	Text           string   `json:"text,omitempty"`
	ReplyToMessage *Message `json:"reply_to_message,omitempty"`
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
	repo        repository.Repository
	llmClient   *llm.Client
	telegramBot *telegram.Client
	botName     string
}

// NewWebhookHandler создает новый экземпляр WebhookHandler
func NewWebhookHandler(repo repository.Repository, llmClient *llm.Client, telegramBot *telegram.Client, botName string) *WebhookHandler {
	return &WebhookHandler{
		repo:        repo,
		llmClient:   llmClient,
		telegramBot: telegramBot,
		botName:     botName,
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

// HandleWebhook обрабатывает входящие обновления от Telegram
func (h *WebhookHandler) HandleWebhook(w http.ResponseWriter, r *http.Request) {
	// Проверяем токен в URL только если есть Telegram клиент
	if h.telegramBot != nil {
		token := chi.URLParam(r, "token")

		// Извлекаем часть токена после двоеточия
		fullToken := h.telegramBot.GetToken()
		parts := strings.Split(fullToken, ":")
		var expectedToken string
		if len(parts) == 2 {
			expectedToken = parts[1] // Берем часть после двоеточия
		} else {
			expectedToken = fullToken // Если нет двоеточия, используем весь токен
		}

		if token != expectedToken {
			log.Printf("Неверный токен в webhook URL: получен '%s', ожидался '%s'", token, expectedToken)
			log.Printf("Полный токен бота: '%s'", fullToken)
			http.Error(w, "Forbidden", http.StatusForbidden)
			return
		}

		log.Printf("Токен проверен успешно: '%s'", token)
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

	// Проверяем, адресовано ли сообщение боту
	if h.isMessageForBot(msg) {
		log.Printf("Сообщение адресовано боту, обрабатываем через LLM")
		if err := h.handleBotMessage(ctx, msg); err != nil {
			log.Printf("Ошибка обработки сообщения через LLM: %v", err)
		}
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

// isMessageForBot проверяет, адресовано ли сообщение боту
func (h *WebhookHandler) isMessageForBot(msg *Message) bool {
	if msg == nil {
		return false
	}

	// Проверяем, является ли это ответом на сообщение бота
	if h.isReplyToBot(msg) {
		log.Printf("Сообщение является ответом на сообщение бота")
		return true
	}

	// Проверяем наличие склонений имени "толик" (регистронезависимо)
	if h.containsTolikVariation(msg.Text) {
		log.Printf("Сообщение содержит обращение к Толику")
		return true
	}

	return false
}

// isReplyToBot проверяет, является ли сообщение ответом на сообщение бота
func (h *WebhookHandler) isReplyToBot(msg *Message) bool {
	if msg.ReplyToMessage == nil || msg.ReplyToMessage.From == nil {
		return false
	}

	// Проверяем, что отвечаем на сообщение бота
	return msg.ReplyToMessage.From.IsBot
}

// containsTolikVariation проверяет содержит ли текст любое склонение имени "толик"
func (h *WebhookHandler) containsTolikVariation(text string) bool {
	if text == "" {
		return false
	}

	lowerText := strings.ToLower(text)

	// Список склонений имени "толик"
	tolikVariations := []string{
		"толик",   // именительный падеж
		"толика",  // родительный падеж
		"толику",  // дательный падеж
		"толиком", // творительный падеж
		"толике",  // предложный падеж
		"толь",    // сокращение
		"толя",    // альтернативная форма
	}

	for _, variation := range tolikVariations {
		if strings.Contains(lowerText, variation) {
			return true
		}
	}

	return false
}

// handleBotMessage обрабатывает сообщение, адресованное боту
func (h *WebhookHandler) handleBotMessage(ctx context.Context, msg *Message) error {
	// Получаем сообщения за последние 3 дня
	recentMessages, err := h.repo.GetRecentMessages(ctx, msg.Chat.ID, 3)
	if err != nil {
		return fmt.Errorf("ошибка получения истории сообщений: %w", err)
	}

	log.Printf("Найдено %d сообщений за последние 3 дня", len(recentMessages))

	// Генерируем ответ через LLM
	response, err := h.llmClient.GenerateResponse(ctx, recentMessages, msg.Text)
	if err != nil {
		return fmt.Errorf("ошибка генерации ответа LLM: %w", err)
	}

	log.Printf("LLM ответ: %s", response)

	// Отправляем ответ в Telegram
	if h.telegramBot != nil {
		if err := h.telegramBot.SendMessage(ctx, msg.Chat.ID, response, msg.MessageID); err != nil {
			return fmt.Errorf("ошибка отправки сообщения в Telegram: %w", err)
		}
		log.Printf("Ответ отправлен в чат %d", msg.Chat.ID)
	} else {
		log.Printf("Telegram клиент не инициализирован, ответ не отправлен")
	}

	return nil
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
