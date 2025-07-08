package telegram

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// Client представляет клиент для работы с Telegram Bot API
type Client struct {
	token      string
	baseURL    string
	httpClient *http.Client
}

// NewClient создает новый экземпляр Telegram клиента
func NewClient(token string) *Client {
	return &Client{
		token:   token,
		baseURL: "https://api.telegram.org/bot" + token,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
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

	// Конвертируем запрос в JSON
	jsonData, err := json.Marshal(request)
	if err != nil {
		return fmt.Errorf("ошибка кодирования JSON: %w", err)
	}

	// Создаем HTTP запрос
	url := c.baseURL + "/sendMessage"
	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("ошибка создания запроса: %w", err)
	}

	// Устанавливаем заголовки
	req.Header.Set("Content-Type", "application/json")

	// Выполняем запрос
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("ошибка выполнения HTTP запроса: %w", err)
	}
	defer resp.Body.Close()

	// Читаем ответ
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("ошибка чтения ответа: %w", err)
	}

	// Парсим ответ
	var response SendMessageResponse
	if err := json.Unmarshal(body, &response); err != nil {
		return fmt.Errorf("ошибка парсинга ответа: %w", err)
	}

	// Проверяем успешность отправки
	if !response.OK {
		return fmt.Errorf("Telegram API вернул ошибку %d: %s", response.ErrorCode, response.Description)
	}

	return nil
}

func (c *Client) GetToken() string {
	return c.token
}
