package config

import (
	"fmt"
	"log"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	Port          string
	TelegramToken string
	WebhookURL    string
	DatabasePath  string
	OpenRouterKey string
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
		DatabasePath:  getEnvWithDefault("DATABASE_PATH", "./data/sueta.db"),
		OpenRouterKey: getEnv("OPENROUTER_API_KEY"),
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
