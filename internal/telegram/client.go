package telegram

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/semyon-ancherbak/sueta/internal/models"
	"github.com/semyon-ancherbak/sueta/internal/repository"
)

type Client struct {
	token      string
	baseURL    string
	httpClient *http.Client
	repo       repository.Repository // Добавляем репозиторий для сохранения сообщений
}

func NewClient(token string, repo repository.Repository) *Client {
	return &Client{
		token:   token,
		baseURL: "https://api.telegram.org/bot" + token,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
		repo: repo,
	}
}

// SendMessageRequest представляет запрос для отправки сообщения
type SendMessageRequest struct {
	ChatID           int64  `json:"chat_id"`
	Text             string `json:"text"`
	ParseMode        string `json:"parse_mode,omitempty"`
	ReplyToMessageID int    `json:"reply_to_message_id,omitempty"`
}

// SendMessageResponse представляет ответ на отправку сообщения
type SendMessageResponse struct {
	OK          bool     `json:"ok"`
	Result      *Message `json:"result,omitempty"`
	Description string   `json:"description,omitempty"`
	ErrorCode   int      `json:"error_code,omitempty"`
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

// SendMessage отправляет сообщение в указанный чат
func (c *Client) SendMessage(ctx context.Context, chatID int64, text string, replyToMessageID int) error {
	request := SendMessageRequest{
		ChatID: chatID,
		Text:   text,
	}

	if replyToMessageID > 0 {
		request.ReplyToMessageID = replyToMessageID
	}

	jsonData, err := json.Marshal(request)
	if err != nil {
		return fmt.Errorf("ошибка кодирования JSON: %w", err)
	}

	url := c.baseURL + "/sendMessage"
	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("ошибка создания запроса: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("ошибка выполнения HTTP запроса: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("ошибка чтения ответа: %w", err)
	}

	var response SendMessageResponse
	if err := json.Unmarshal(body, &response); err != nil {
		return fmt.Errorf("ошибка парсинга ответа: %w", err)
	}

	if !response.OK {
		return fmt.Errorf("telegram API вернул ошибку %d: %s", response.ErrorCode, response.Description)
	}

	if response.Result != nil {
		if err := c.saveBotMessage(ctx, response.Result, text); err != nil {
			// Логируем ошибку, но не возвращаем её, так как сообщение уже отправлено
			fmt.Printf("Ошибка сохранения сообщения бота: %v\n", err)
		}
	}

	return nil
}

func (c *Client) saveBotMessage(ctx context.Context, msg *Message, text string) error {
	if msg == nil || c.repo == nil {
		return nil
	}

	messageDoc := &models.MessageDocument{
		MessageID: msg.MessageID,
		ChatID:    msg.Chat.ID,
		Text:      text,
		Date:      time.Unix(msg.Date, 0),
		UpdateID:  0,    // Для сообщений бота UpdateID = 0
		IsBot:     true, // Помечаем как сообщение от бота
		CreatedAt: time.Now(),
	}

	// Добавляем информацию о боте как пользователе
	if msg.From != nil {
		messageDoc.UserID = msg.From.ID
		messageDoc.Username = msg.From.Username
		messageDoc.FirstName = msg.From.FirstName
		messageDoc.LastName = msg.From.LastName
	}

	if err := c.repo.SaveMessage(ctx, messageDoc); err != nil {
		return fmt.Errorf("ошибка сохранения сообщения бота: %w", err)
	}

	fmt.Printf("Сохранено сообщение бота: ID=%d, ChatID=%d, Text=%s\n",
		msg.MessageID, msg.Chat.ID, text)
	return nil
}
