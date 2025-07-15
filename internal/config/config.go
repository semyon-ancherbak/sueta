package config

import (
	"fmt"
	"log"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

type Config struct {
	Port          string
	TelegramToken string
	WebhookURL    string
	MongoURL      string
	MongoDatabase string
	OpenRouterKey string
	// RAG настройки
	RAGMaxRelevantMessages int
	RAGRecentDaysExclude   int
}

func Load() (*Config, error) {
	// Пытаемся загрузить .env файл (если существует)
	if err := godotenv.Load(); err != nil {
		log.Println("Файл .env не найден, используем переменные окружения")
	}

	config := &Config{
		Port:          getEnvWithDefault("PORT", "8080"),
		TelegramToken: getEnv("TELEGRAM_TOKEN"),
		WebhookURL:    getEnv("WEBHOOK_URL"),
		MongoURL:      getEnvWithDefault("MONGO_URL", "mongodb://localhost:27017"),
		MongoDatabase: getEnvWithDefault("MONGO_DATABASE", "sueta"),
		OpenRouterKey: getEnv("OPENROUTER_API_KEY"),
		// RAG настройки
		RAGMaxRelevantMessages: getEnvWithDefaultInt("RAG_MAX_RELEVANT_MESSAGES", 5),
		RAGRecentDaysExclude:   getEnvWithDefaultInt("RAG_RECENT_DAYS_EXCLUDE", 3),
	}

	if err := validateConfig(config); err != nil {
		return nil, fmt.Errorf("ошибка в конфигурации: %w", err)
	}
	return config, nil
}

func validateConfig(cfg *Config) error {
	errors := []string{}
	if cfg.TelegramToken == "" {
		errors = append(errors, "TELEGRAM_TOKEN не установлен")
	}
	if cfg.OpenRouterKey == "" {
		errors = append(errors, "OPENROUTER_API_KEY не установлен")
	}
	if cfg.WebhookURL == "" {
		errors = append(errors, "WEBHOOK_URL не установлен")
	}
	if len(errors) > 0 {
		return fmt.Errorf("конфигурация содержит ошибки: %s", errors)
	}
	return nil
}

func getEnv(key string) string {
	return os.Getenv(key)
}

func getEnvWithDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvWithDefaultInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}
