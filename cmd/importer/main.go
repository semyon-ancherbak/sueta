package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/semyon-ancherbak/sueta/internal/config"
	"github.com/semyon-ancherbak/sueta/internal/models"
	"github.com/semyon-ancherbak/sueta/internal/repository"
	"go.mongodb.org/mongo-driver/mongo"
)

func main() {
	// Флаги командной строки
	var (
		filePath = flag.String("file", "", "Путь к JSON файлу с экспортом Telegram")
		dryRun   = flag.Bool("dry-run", false, "Только показать статистику, не сохранять в БД")
		verbose  = flag.Bool("v", false, "Подробный вывод")
		noDB     = flag.Bool("no-db", false, "Не подключаться к БД (только для тестирования парсинга)")
	)
	flag.Parse()

	if *filePath == "" {
		log.Fatal("Укажите путь к файлу экспорта через флаг -file")
	}

	// Загружаем конфигурацию
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Ошибка загрузки конфигурации: %v", err)
	}

	// Подключение к MongoDB (если не указан флаг --no-db)
	ctx := context.Background()
	var repo repository.Repository
	if !*noDB {
		repo, err = repository.NewMongoRepository(ctx, cfg.MongoURL, cfg.MongoDatabase)
		if err != nil {
			log.Fatalf("Ошибка подключения к MongoDB: %v", err)
		}
		defer func() {
			if err := repo.Close(ctx); err != nil {
				log.Printf("Ошибка закрытия соединения с MongoDB: %v", err)
			}
		}()
	}

	// Парсинг и импорт
	importer := &TelegramImporter{
		repo:    repo,
		dryRun:  *dryRun,
		verbose: *verbose,
	}

	if err := importer.ImportFromFile(ctx, *filePath); err != nil {
		log.Fatalf("Ошибка импорта: %v", err)
	}
}

// TelegramImporter отвечает за импорт данных из экспорта Telegram
type TelegramImporter struct {
	repo    repository.Repository
	dryRun  bool
	verbose bool
}

// ImportFromFile импортирует данные из JSON файла
func (ti *TelegramImporter) ImportFromFile(ctx context.Context, filePath string) error {
	log.Printf("Начинаем импорт из файла: %s", filePath)

	// Читаем файл
	data, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("ошибка чтения файла: %w", err)
	}

	// Парсим JSON
	var export models.TelegramExport
	if err := json.Unmarshal(data, &export); err != nil {
		return fmt.Errorf("ошибка парсинга JSON: %w", err)
	}

	log.Printf("Найдено сообщений в экспорте: %d", len(export.Messages))
	log.Printf("Название чата: %s", export.Name)
	log.Printf("Тип чата: %s", export.Type)
	log.Printf("ID чата: %d", export.ID)

	// Сохраняем информацию о чате
	if !ti.dryRun {
		chatDoc := &models.ChatDocument{
			ChatID:    export.ID,
			Type:      export.Type,
			Title:     export.Name,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}

		if err := ti.repo.SaveChat(ctx, chatDoc); err != nil {
			// Игнорируем ошибку дублирования чата
			if ti.verbose {
				log.Printf("Предупреждение при сохранении чата: %v", err)
			}
		}
	}

	// Обрабатываем сообщения
	stats := &ImportStats{}
	for i, msg := range export.Messages {
		if ti.verbose && i%1000 == 0 {
			log.Printf("Обработано сообщений: %d/%d", i, len(export.Messages))
		}

		parsedMsg, err := ti.parseMessage(&msg, export.ID)
		if err != nil {
			stats.Errors++
			if ti.verbose {
				log.Printf("Ошибка парсинга сообщения %d: %v", msg.ID, err)
			}
			continue
		}

		if parsedMsg == nil {
			stats.Skipped++
			continue
		}

		// Сохраняем сообщение
		if !ti.dryRun {
			msgDoc := ti.convertToMessageDocument(parsedMsg)
			if err := ti.repo.SaveMessage(ctx, msgDoc); err != nil {
				// Проверяем, является ли ошибка дублированием
				if mongo.IsDuplicateKeyError(err) {
					stats.Duplicates++
					if ti.verbose {
						log.Printf("Сообщение %d уже существует в БД", msg.ID)
					}
				} else {
					stats.Errors++
					if ti.verbose {
						log.Printf("Ошибка сохранения сообщения %d: %v", msg.ID, err)
					}
				}
				continue
			}
		}

		if parsedMsg.IsService {
			stats.ServiceMessages++
		} else {
			stats.UserMessages++
		}
		stats.Total++
	}

	// Выводим статистику
	ti.printStats(stats)
	return nil
}

// parseMessage парсит сообщение из экспорта
func (ti *TelegramImporter) parseMessage(msg *models.TelegramExportMessage, chatID int64) (*models.ParsedMessage, error) {
	// Парсим дату
	date, err := time.Parse("2006-01-02T15:04:05", msg.Date)
	if err != nil {
		return nil, fmt.Errorf("ошибка парсинга даты: %w", err)
	}

	// Извлекаем текст
	text := msg.ExtractTextFromMessage()

	// Пропускаем пустые сообщения (кроме служебных)
	if text == "" && msg.Type != "service" {
		return nil, nil // Сообщение будет пропущено
	}

	// Определяем отправителя
	var username, firstName string
	var userID int64

	if msg.From != "" {
		username = msg.From
		firstName = msg.From
	} else if msg.Actor != "" {
		username = msg.Actor
		firstName = msg.Actor
	}

	userID = msg.ExtractUserIDFromActorID()

	// Создаем парсеную структуру
	parsed := &models.ParsedMessage{
		OriginalID:  msg.ID,
		ChatID:      chatID,
		UserID:      userID,
		Username:    username,
		FirstName:   firstName,
		Text:        text,
		Date:        date,
		MessageType: msg.Type,
		IsService:   msg.Type == "service",
	}

	return parsed, nil
}

// convertToMessageDocument конвертирует ParsedMessage в MessageDocument для MongoDB
func (ti *TelegramImporter) convertToMessageDocument(parsed *models.ParsedMessage) *models.MessageDocument {
	return &models.MessageDocument{
		MessageID: int(parsed.OriginalID),
		ChatID:    parsed.ChatID,
		UserID:    parsed.UserID,
		Username:  parsed.Username,
		FirstName: parsed.FirstName,
		Text:      parsed.Text,
		Date:      parsed.Date,
		UpdateID:  0, // Для импортированных сообщений
		IsBot:     false,
		CreatedAt: time.Now(),
	}
}

// ImportStats содержит статистику импорта
type ImportStats struct {
	Total           int
	UserMessages    int
	ServiceMessages int
	Skipped         int
	Errors          int
	Duplicates      int
}

// printStats выводит статистику импорта
func (ti *TelegramImporter) printStats(stats *ImportStats) {
	log.Println("=== Статистика импорта ===")
	log.Printf("Всего обработано: %d", stats.Total)
	log.Printf("Пользовательских сообщений: %d", stats.UserMessages)
	log.Printf("Служебных сообщений: %d", stats.ServiceMessages)
	log.Printf("Пропущено: %d", stats.Skipped)
	log.Printf("Дубликатов: %d", stats.Duplicates)
	log.Printf("Ошибок: %d", stats.Errors)

	if ti.dryRun {
		log.Println("*** РЕЖИМ ПРОБНОГО ЗАПУСКА - ДАННЫЕ НЕ СОХРАНЕНЫ ***")
	} else {
		log.Println("*** ИМПОРТ ЗАВЕРШЕН ***")
	}
}
