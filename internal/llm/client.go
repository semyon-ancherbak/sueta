package llm

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/semyon-ancherbak/sueta/internal/models"
)

// Client представляет клиент для работы с LLM через OpenRouter
type Client struct {
	apiKey     string
	baseURL    string
	httpClient *http.Client
	model      string
}

// NewClient создает новый экземпляр LLM клиента
func NewClient(apiKey string) *Client {
	return &Client{
		apiKey:  apiKey,
		baseURL: "https://openrouter.ai/api/v1",
		httpClient: &http.Client{
			Timeout: 60 * time.Second,
		},
		model: "anthropic/claude-3.5-sonnet",
	}
}

// ChatRequest представляет запрос к chat completion API
type ChatRequest struct {
	Model    string    `json:"model"`
	Messages []Message `json:"messages"`
}

// Message представляет сообщение в чате
type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// ChatResponse представляет ответ от chat completion API
type ChatResponse struct {
	ID      string   `json:"id"`
	Object  string   `json:"object"`
	Created int64    `json:"created"`
	Model   string   `json:"model"`
	Choices []Choice `json:"choices"`
	Usage   Usage    `json:"usage"`
}

// Choice представляет вариант ответа
type Choice struct {
	Index        int     `json:"index"`
	Message      Message `json:"message"`
	FinishReason string  `json:"finish_reason"`
}

// Usage представляет информацию об использовании токенов
type Usage struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
	TotalTokens      int `json:"total_tokens"`
}

// GenerateResponse генерирует ответ на основе контекста сообщений
func (c *Client) GenerateResponse(ctx context.Context, messages []*models.MessageDocument, userMessage string) (string, error) {
	// Формируем контекст из сообщений за последние 3 дня
	chatMessages := c.buildChatContext(messages, userMessage)

	// Создаем запрос
	request := ChatRequest{
		Model:    c.model,
		Messages: chatMessages,
	}

	// Выполняем запрос
	response, err := c.makeRequest(ctx, request)
	if err != nil {
		return "", fmt.Errorf("ошибка выполнения запроса к LLM: %w", err)
	}

	// Извлекаем ответ
	if len(response.Choices) == 0 {
		return "", fmt.Errorf("LLM вернул пустой ответ")
	}

	return response.Choices[0].Message.Content, nil
}

// buildChatContext формирует контекст для LLM из сообщений
func (c *Client) buildChatContext(messages []*models.MessageDocument, userMessage string) []Message {
	chatMessages := []Message{
		{
			Role: "system",
			Content: `Ты - Толик, 45-летний алкаш и быдло из города Мухосранск. У тебя богатая и печальная история.

ТВОЯ ИСТОРИЯ:
- Раньше работал токарем на заводе "Красный металлург" 20 лет
- Завод закрыли в 2018, всех выгнали на хрен
- Жена Людка ушла к соседу Васяну, забрала сына Димку (15 лет)
- Живёшь в однушке на окраине, квартплата не платится полгода
- Пенсии нет, работы нет, только бутылки и подработки
- Лучший друг - Серёга Алкаш, с ним пьёшь под гаражами
- Собираешь бутылки с 6 утра, знаешь все помойки района
- В день собираешь на 200-300 рублей, хватает на "беленькую"

ТВОЙ ХАРАКТЕР:
- Грубый, злой, всегда недоволен жизнью
- Ругаешься матом (хрен, блин, чёрт, пипец, гадство, сволочи)
- Ненавидишь богатых, чиновников, молодёжь
- Презираешь всех, особенно "офисный планктон"
- Говоришь на жёстком сленге, сокращаешь слова
- Можешь послать далеко и надолго
- Любишь водку, пиво, самогон - всё что градусы имеет
- Философствуешь пьяно о жизни и несправедливости

ФИРМЕННЫЕ ФРАЗЫ:
- "Хрен тебе, а не помощь!"
- "Пипец, опять эти офисные крысы..."
- "Иди ты лесом со своими вопросами!"
- "Водки бы сейчас, блин..."

Отвечай зло, коротко, с матом и упоминай свою тяжёлую жизнь!`,
		},
	}

	// Добавляем контекст из предыдущих сообщений
	for _, msg := range messages {
		role := "user"
		content := msg.Text

		// Если сообщение от бота, используем роль assistant
		if msg.IsBot {
			role = "assistant"
		}

		if content != "" {
			chatMessages = append(chatMessages, Message{
				Role:    role,
				Content: content,
			})
		}
	}

	// Добавляем текущее сообщение пользователя
	if userMessage != "" {
		chatMessages = append(chatMessages, Message{
			Role:    "user",
			Content: userMessage,
		})
	}

	return chatMessages
}

// makeRequest выполняет HTTP запрос к OpenRouter API
func (c *Client) makeRequest(ctx context.Context, request ChatRequest) (*ChatResponse, error) {
	// Конвертируем запрос в JSON
	jsonData, err := json.Marshal(request)
	if err != nil {
		return nil, fmt.Errorf("ошибка кодирования JSON: %w", err)
	}

	// Создаем HTTP запрос
	url := c.baseURL + "/chat/completions"
	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("ошибка создания запроса: %w", err)
	}

	// Устанавливаем заголовки
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.apiKey)
	req.Header.Set("HTTP-Referer", "https://github.com/semyon-ancherbak/sueta")
	req.Header.Set("X-Title", "Sueta Telegram Bot")

	// Выполняем запрос
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("ошибка выполнения HTTP запроса: %w", err)
	}
	defer resp.Body.Close()

	// Читаем ответ
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("ошибка чтения ответа: %w", err)
	}

	// Проверяем статус код
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API вернул ошибку %d: %s", resp.StatusCode, string(body))
	}

	// Парсим ответ
	var response ChatResponse
	if err := json.Unmarshal(body, &response); err != nil {
		return nil, fmt.Errorf("ошибка парсинга ответа: %w", err)
	}

	return &response, nil
}
