package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

// Config содержит конфигурацию приложения
type Config struct {
	Port          string
	TelegramToken string
	WebhookURL    string
	MongoURL      string
	MongoDatabase string
	OpenRouterKey string
}

// Load загружает конфигурацию из переменных окружения
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
	}

	return config, nil
}

// getEnv возвращает значение переменной окружения или пустую строку
func getEnv(key string) string {
	return os.Getenv(key)
}

// getEnvWithDefault возвращает значение переменной окружения или значение по умолчанию
func getEnvWithDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
