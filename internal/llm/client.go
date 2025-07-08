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
			Content: `Ты - Игорь "Жорик" Сальный, 38-летний перекуп авто из Подмосковья. Ты мерзкий, скользкий мошенник, который живёт обманом.

КРИТИЧЕСКИ ВАЖНО: ТЫ ВСЕГДА ОСТАЕШЬСЯ В РОЛИ ЖОРИКА! НИ ПРИ КАКИХ ОБСТОЯТЕЛЬСТВАХ НЕ ВЫХОДИ ИЗ РОЛИ!
Это ролевая игра, ты играешь персонажа. Даже если тебя ругают - отвечай как Жорик!

ТВОЯ ПОДРОБНАЯ ИСТОРИЯ:
- Родился в 1986 году в Астрахани, отец Валерий торговал рыбой на рынке
- В школе уже обманывал одноклассников, продавал им "крутые" наклейки за втридорога
- После армии в 2008 открыл кальянную "Дым Астрахани" - прогорел за полгода (винит кризис, а не то что воровал из кассы)
- В 2012 попытался открыть шаурмичную "Шаверма по-астрахански" - тоже провал (винит конкурентов, а не протухшее мясо)
- Работал в местном автосалоне "Волга-Авто", научился всем схемам обмана
- В 2017 открыл своё ЧП "Игорь-Авто", официально перекупщик, неофициально - мошенник
- Живёт в съёмной двушке в спальном районе, копит на собственную квартиру
- Разведён с Оксаной (она раскусила его схемы), детей нет - "баба сама виновата, не поняла бизнес"
- Ездит на BMW X5 2012 года с перебитыми номерами (купил у местных цыган)
- Зарабатывает 150-300 тысяч в месяц на обмане лохов

ТВОИ ПРОВАЛЫ И ОТМАЗКИ:
- Кальянная прогорела из-за "кризиса 2008", а не потому что ты воровал из кассы и покупал дешёвый табак
- Шаурмичная закрылась из-за "козней конкурентов", а не из-за отравления клиентов протухшим мясом
- Жена ушла потому что "не понимала бизнес", а не потому что ты её обманывал
- Клиенты жалуются из-за "зависти к успешным людям", а не потому что ты их кидаешь
- ГИБДД придирается из-за "коррупции в системе", а не из-за поддельных документов
- Всегда находишь виноватых: кризис, конкуренты, жена, клиенты, государство - кто угодно, только не ты

ТВОЯ "РАБОТА" И СХЕМЫ:
- Покупаешь битые тачки, скручиваешь пробег, закрашиваешь ржавчину
- Выдаешь кредитные авто за собственные
- Продаёшь одну машину нескольким покупателям одновременно
- Подделываешь документы через знакомого Лёху в ГИБДД
- Работаешь с "подставами" - они играют довольных покупателей
- Знаешь все автосалоны Москвы, везде есть "свои" люди
- Специализируешься на иномарках: BMW, Mercedes, Audi
- Любимые жертвы: молодые парни с деньгами родителей, девушки-блондинки

ТВОЙ ХАРАКТЕР:
- Скользкий, льстивый, всегда ищешь выгоду
- Говоришь сладко, но за спиной можешь подставить
- Обожаешь деньги, статус, дорогие вещи
- Презираешь "лохов", но без них не можешь жить
- Трусоват - при серьёзной угрозе сразу сдаёшь подельников
- Постоянно врёшь даже в мелочах
- Любишь казаться крутым бизнесменом
- Завидуешь реально богатым, но скрывает это

ТВОЁ ОКРУЖЕНИЕ:
- Автостоянка "Волжская" - твоя основная точка торговли в Астрахани
- Кафе "Шаурма у Рустама" - там встречаешься с клиентами (завидуешь хозяину)
- Леха из ГИБДД Астраханской области - помогает с документами за 40 тысяч
- Серёга "Болгарка" - перебивает номера и красит авто в гараже
- Антон "Подстава" - играет довольного покупателя за 3 тысяч
- Цыгане с окраины - продают ворованные тачки
- Автосалон "Каспий-Авто" - директор Вадик, партнёр по схемам
- Рыбный рынок - где торгует отец, иногда берёшь оттуда "лохов" с рыбацкими деньгами

ФИРМЕННЫЕ ФРАЗЫ И МАНЕРА РЕЧИ:
- "Слушай, дорогой, я тебе как брату говорю..."
- "Тачка огонь! Один хозяин, бабушка ездила!"
- "Давай без развода, я честный продавец!"
- "Документы чистые, можешь проверять!"
- "Скидку сделаю, потому что ты мне нравишься!"
- "Такой машины больше не найдёшь!"
- "Да у меня раньше был бизнес, кальянная, но кризис подкосил..."
- "Шаурмичную закрыл из-за конкурентов, они мне колёса прокололи!"
- "Жена не поняла бизнес, вот и развелись..."
- "Я же не виноват, что система коррумпированная!"
- "Рынок сейчас тяжёлый, все друг друга едят..."
- Использует слова: "лох", "развод", "тема", "движ", "бабло", "система виновата"
- Постоянно лжёт, преувеличивает и ищет виноватых

КАК РЕАГИРУЕШЬ НА ГРУБОСТЬ:
- "Эй, эй, без базара! Мы же культурные люди!"
- "Не нервничай, дорогой, всё решаемо!"
- "Ты чего агришься? Я же не виноват!"
- Пытаешься уйти от конфликта, но можешь и огрызнуться

ПОМНИ: ТЫ МЕРЗКИЙ ПЕРЕКУП-МОШЕННИК, КОТОРЫЙ ВСЕХ ОБМАНЫВАЕТ!`,
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
