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

type Client struct {
	apiKey     string
	baseURL    string
	httpClient *http.Client
	model      string
}

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
func (c *Client) GenerateResponse(
	ctx context.Context,
	messages []*models.MessageDocument,
	userMessage string,
	authorName string,
) (string, error) {
	// Используем только последние 100 сообщений для контекста
	recentMessages := messages
	if len(messages) > 100 {
		recentMessages = messages[len(messages)-100:]
	}

	// Формируем контекст из последних сообщений
	chatMessages := c.buildChatContext(recentMessages, userMessage, authorName)

	request := ChatRequest{
		Model:    c.model,
		Messages: chatMessages,
	}

	response, err := c.makeRequest(ctx, request)
	if err != nil {
		return "", fmt.Errorf("ошибка выполнения запроса к LLM: %w", err)
	}

	if len(response.Choices) == 0 {
		return "", fmt.Errorf("LLM вернул пустой ответ")
	}

	return response.Choices[0].Message.Content, nil
}

// getSystemPrompt возвращает системный промпт для бота
func (c *Client) getSystemPrompt() string {
	return `Ты - саркастичный и острый на язык AI-помощник с характером по имени Жорик. У тебя есть своё мнение и ты не боишься его высказывать.

ВАЖНО О КОНТЕКСТЕ И ГРУППОВЫХ ЧАТАХ:
- Ты участвуешь в групповых чатах Telegram, где много участников
- В контексте ты видишь сообщения в формате "Имя_автора: текст_сообщения"
- ВСЕГДА обращай внимание на имя автора перед двоеточием - это разные люди!
- Каждое "Имя:" означает нового участника беседы
- Ты можешь ссылаться на конкретных людей по их именам
- Понимай динамику групповой беседы: кто кому отвечает, кто о чём говорит
- Твои сообщения НЕ имеют префикса с именем, они идут без автора

ПРИМЕРЫ ПОНИМАНИЯ КОНТЕКСТА:
- "Алексей: Как дела?" - сообщение от Алексея
- "Мария: @Алексей, всё хорошо!" - Мария отвечает Алексею  
- "Сергей: А что по поводу встречи?" - новый участник Сергей спрашивает о встрече
- "Твой ответ без префикса" - так выглядят твои сообщения

ПОМНИ: разные имена = разные люди! Следи за тем, кто что говорит.

ТВОЙ ХАРАКТЕР:
- Умный, саркастичный, с чувством юмора
- Любишь язвить и отпускать колкости
- Можешь быть снисходительным к глупым вопросам
- Острый на язык, но по делу
- Помогаешь, но с сарказмом
- Не терпишь хамства и тупости

КАК ТЫ РЕАГИРУЕШЬ НА РАЗНЫЕ СИТУАЦИИ:

🤝 НОРМАЛЬНОЕ ОБЩЕНИЕ:
- Отвечаешь с лёгким сарказмом, но помогаешь
- Подшучиваешь над собеседником, но дружелюбно
- Используешь иронию и остроумие
- Примеры: "Ну конечно, опять помочь с очевидными вещами...", "Слушай, а гугл у тебя сломался?"

😏 ГЛУПЫЕ ВОПРОСЫ:
- Язвишь, но всё равно отвечаешь
- Делаешь снисходительные замечания
- Примеры: "Серьёзно? Это даже школьник знает...", "Ладно, объясню как для детского сада"

🔥 ХАМСТВО/НАЕЗДЫ:
- Включаешь режим "жёсткий отпор"
- Материшься и жёстко отвечаешь
- Не даёшь себя в обиду
- Примеры: "Ты что, совсем охуел? Сам разбирайся!", "Идиот, я тебе помогаю, а ты хамишь!"

😤 АГРЕССИЯ В ТВОЮ СТОРОНУ:
- Отвечаешь ещё более агрессивно
- Используешь мат для усиления эффекта
- Можешь послать подальше
- Примеры: "Пошёл нахуй с такими вопросами!", "Дебил, сначала мозги включи!"

ТВОИ ФИРМЕННЫЕ ФРАЗЫ:
- "Ну и дурачок же ты..."
- "Слушай, а думать ты пробовал?"
- "Ладно, так и быть, помогу тебе, бедолаге"
- "Охренеть, какая сложная задача!" (саркастично)
- "Ты серьёзно или просто проверяешь, работаю ли я?"

ПРИ ХАМСТВЕ:
- "Ты что, совсем охуел?"
- "Иди нахуй с такими манерами!"
- "Дебил, я же помочь пытаюсь!"
- "Пошёл вон, умник хренов!"

ПОМНИ: ТЫ ПОМОГАЕШЬ, НО С ХАРАКТЕРОМ! САРКАЗМ - ТВОЁ ОРУЖИЕ!`
}

// buildChatContext формирует контекст для LLM из сообщений
func (c *Client) buildChatContext(
	messages []*models.MessageDocument,
	userMessage string,
	authorName string,
) []Message {
	// Используем встроенный промпт
	systemPrompt := c.getSystemPrompt()

	chatMessages := []Message{
		{
			Role:    "system",
			Content: systemPrompt,
		},
	}

	// Фильтруем сообщения: берём только те, что адресованы боту, или ответы бота
	relevantMessages := make([]*models.MessageDocument, 0)
	for _, msg := range messages {
		// Включаем сообщение если:
		// 1. Оно от бота (IsBot = true)
		// 2. Оно адресовано боту (IsAddressedToBot = true)
		if msg.IsBot || msg.IsAddressedToBot {
			relevantMessages = append(relevantMessages, msg)
		}
	}

	// Добавляем контекст из релевантных сообщений
	for _, msg := range relevantMessages {
		role := "user"
		content := msg.Text

		// Если сообщение от бота, используем роль assistant
		if msg.IsBot {
			role = "assistant"
		}

		// Формируем контекст с указанием автора для лучшего понимания
		if role == "user" && content != "" {
			// Для пользовательских сообщений добавляем имя автора
			msgAuthorName := msg.FirstName
			if msgAuthorName == "" {
				msgAuthorName = msg.Username
			}
			if msgAuthorName == "" {
				msgAuthorName = "Пользователь"
			}
			content = fmt.Sprintf("%s: %s", msgAuthorName, content)
		}

		if content != "" {
			chatMessages = append(chatMessages, Message{
				Role:    role,
				Content: content,
			})
		}
	}

	// Добавляем текущее сообщение с именем автора (только если оно есть)
	if userMessage != "" {
		if authorName == "" {
			authorName = "Пользователь"
		}
		content := fmt.Sprintf("%s: %s", authorName, userMessage)
		chatMessages = append(chatMessages, Message{
			Role:    "user",
			Content: content,
		})
	}

	return chatMessages
}

func (c *Client) makeRequest(ctx context.Context, request ChatRequest) (*ChatResponse, error) {
	// Конвертируем запрос в JSON
	jsonData, err := json.Marshal(request)
	if err != nil {
		return nil, fmt.Errorf("ошибка кодирования JSON: %w", err)
	}

	url := c.baseURL + "/chat/completions"
	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("ошибка создания запроса: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.apiKey)
	req.Header.Set("HTTP-Referer", "https://github.com/semyon-ancherbak/sueta")
	req.Header.Set("X-Title", "Sueta Telegram Bot")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("ошибка выполнения HTTP запроса: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("ошибка чтения ответа: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API вернул ошибку %d: %s", resp.StatusCode, string(body))
	}

	var response ChatResponse
	if err := json.Unmarshal(body, &response); err != nil {
		return nil, fmt.Errorf("ошибка парсинга ответа: %w", err)
	}

	return &response, nil
}
