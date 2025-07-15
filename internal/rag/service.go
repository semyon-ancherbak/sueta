package rag

import (
	"context"
	"fmt"
	"log"
	"regexp"
	"strings"

	"github.com/semyon-ancherbak/sueta/internal/models"
	"github.com/semyon-ancherbak/sueta/internal/repository"
)

// Service предоставляет функциональность RAG (Retrieval-Augmented Generation)
type Service struct {
	repo repository.Repository
}

// NewService создает новый экземпляр RAG сервиса
func NewService(repo repository.Repository) *Service {
	return &Service{
		repo: repo,
	}
}

// RetrieveRelevantContext получает релевантный контекст для генерации ответа
func (s *Service) RetrieveRelevantContext(
	ctx context.Context,
	chatID int64,
	userMessage string,
	maxRelevantMessages int,
	excludeDays int,
) ([]*models.MessageDocument, []*models.MessageDocument, error) {
	// Получаем последние сообщения (как и раньше)
	recentMessages, err := s.repo.GetRecentMessages(ctx, chatID, excludeDays)
	if err != nil {
		return nil, nil, fmt.Errorf("ошибка получения недавних сообщений: %w", err)
	}

	// Извлекаем ключевые слова из сообщения пользователя для поиска
	searchQuery := s.extractSearchKeywords(userMessage)
	if searchQuery == "" {
		log.Printf("Не удалось извлечь ключевые слова из сообщения: %s", userMessage)
		return recentMessages, []*models.MessageDocument{}, nil
	}

	log.Printf("Поиск релевантных сообщений по запросу: %s", searchQuery)

	// Ищем релевантные старые сообщения
	relevantMessages, err := s.repo.SearchRelevantMessages(ctx, chatID, searchQuery, maxRelevantMessages, excludeDays)
	if err != nil {
		log.Printf("Ошибка поиска релевантных сообщений: %v", err)
		// Не возвращаем ошибку, просто используем только недавние сообщения
		return recentMessages, []*models.MessageDocument{}, nil
	}

	log.Printf("Найдено %d релевантных сообщений", len(relevantMessages))

	return recentMessages, relevantMessages, nil
}

// extractSearchKeywords извлекает ключевые слова из сообщения пользователя для поиска
func (s *Service) extractSearchKeywords(message string) string {
	if message == "" {
		return ""
	}

	// Убираем обращения к боту
	message = s.removeBotMentions(message)

	// Убираем знаки препинания и лишние пробелы
	re := regexp.MustCompile(`[^\p{L}\p{N}\s]+`)
	cleaned := re.ReplaceAllString(message, " ")
	cleaned = regexp.MustCompile(`\s+`).ReplaceAllString(cleaned, " ")
	cleaned = strings.TrimSpace(cleaned)

	// Фильтруем стоп-слова и короткие слова
	words := strings.Fields(strings.ToLower(cleaned))
	var keywords []string

	stopWords := map[string]bool{
		"что": true, "как": true, "где": true, "когда": true, "почему": true,
		"кто": true, "какой": true, "какая": true, "какое": true, "какие": true,
		"это": true, "тот": true, "тех": true, "том": true, "той": true,
		"мне": true, "меня": true, "мной": true, "тебе": true, "тебя": true,
		"его": true, "её": true, "них": true, "ним": true, "нам": true,
		"для": true, "про": true, "без": true, "при": true, "над": true,
		"под": true, "через": true, "между": true, "перед": true, "после": true,
		"или": true, "ну": true, "да": true, "нет": true, "не": true,
		"был": true, "была": true, "было": true, "были": true, "буду": true,
		"будет": true, "будем": true, "есть": true, "быть": true,
	}

	for _, word := range words {
		// Пропускаем короткие слова и стоп-слова
		if len(word) < 3 || stopWords[word] {
			continue
		}
		keywords = append(keywords, word)
	}

	// Если осталось мало ключевых слов, используем исходное сообщение
	if len(keywords) < 2 {
		return cleaned
	}

	return strings.Join(keywords, " ")
}

// removeBotMentions удаляет упоминания бота из сообщения
func (s *Service) removeBotMentions(message string) string {
	lowerMessage := strings.ToLower(message)

	botMentions := []string{
		"жорик", "жорика", "жорику", "жориком", "жорике",
		"жора", "жорж",
	}

	for _, mention := range botMentions {
		// Удаляем упоминания с учетом границ слов
		re := regexp.MustCompile(`\b` + regexp.QuoteMeta(mention) + `\b`)
		lowerMessage = re.ReplaceAllString(lowerMessage, "")
	}

	// Очищаем лишние пробелы
	return regexp.MustCompile(`\s+`).ReplaceAllString(lowerMessage, " ")
}
