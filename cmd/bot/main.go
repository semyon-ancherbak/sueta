package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/semyon-ancherbak/sueta/internal/config"
	"github.com/semyon-ancherbak/sueta/internal/handler"
	"github.com/semyon-ancherbak/sueta/internal/llm"
	"github.com/semyon-ancherbak/sueta/internal/repository"
	"github.com/semyon-ancherbak/sueta/internal/telegram"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Ошибка загрузки конфигурации: %v", err)
	}
	log.Printf("Конфигурация загружена")

	ctx := context.Background()
	repo, err := repository.NewMongoRepository(ctx, cfg.MongoURL, cfg.MongoDatabase)
	if err != nil {
		log.Fatalf("Ошибка подключения к MongoDB: %v", err)
	}
	defer func() {
		if err := repo.Close(ctx); err != nil {
			log.Printf("Ошибка закрытия соединения с MongoDB: %v", err)
		}
	}()
	log.Println("Подключение к MongoDB установлено")

	llmClient := llm.NewClient(cfg.OpenRouterKey)
	log.Println("LLM клиент инициализирован")

	tgClient := telegram.NewClient(cfg.TelegramToken, repo)
	log.Println("Telegram бот клиент инициализирован")

	botName := "Жорик" // Имя бота
	webhookHandler := handler.NewWebhookHandler(repo, llmClient, tgClient, botName, cfg)

	router := webhookHandler.SetupRouter()
	server := &http.Server{
		Addr:         ":" + cfg.Port,
		Handler:      router,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	go func() {
		log.Printf("Сервер запущен на порту %s", cfg.Port)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Ошибка запуска сервера: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Получен сигнал завершения, выключение сервера...")

	// Graceful shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Printf("Ошибка при завершении работы сервера: %v", err)
	}

	log.Println("Сервер остановлен")
}
