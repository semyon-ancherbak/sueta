package repository

import (
	"context"
	"time"

	"github.com/semyon-ancherbak/sueta/internal/models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// Repository интерфейс для работы с данными
type Repository interface {
	SaveChat(ctx context.Context, chat *models.ChatDocument) error
	SaveMessage(ctx context.Context, message *models.MessageDocument) error
	ChatExists(ctx context.Context, chatID int64) (bool, error)
	GetRecentMessages(ctx context.Context, chatID int64, days int) ([]*models.MessageDocument, error)
	Close(ctx context.Context) error
}

// MongoRepository реализация Repository для MongoDB
type MongoRepository struct {
	client   *mongo.Client
	database *mongo.Database
	chats    *mongo.Collection
	messages *mongo.Collection
}

// NewMongoRepository создает новый экземпляр MongoRepository
func NewMongoRepository(ctx context.Context, mongoURL, dbName string) (*MongoRepository, error) {
	client, err := mongo.Connect(ctx, options.Client().ApplyURI(mongoURL))
	if err != nil {
		return nil, err
	}

	// Проверяем соединение
	if err := client.Ping(ctx, nil); err != nil {
		return nil, err
	}

	database := client.Database(dbName)
	chats := database.Collection("chats")
	messages := database.Collection("messages")

	repo := &MongoRepository{
		client:   client,
		database: database,
		chats:    chats,
		messages: messages,
	}

	// Создаем индексы
	if err := repo.createIndexes(ctx); err != nil {
		return nil, err
	}

	return repo, nil
}

// createIndexes создает необходимые индексы
func (r *MongoRepository) createIndexes(ctx context.Context) error {
	// Уникальный индекс для chat_id в коллекции chats
	chatIndex := mongo.IndexModel{
		Keys:    bson.D{{Key: "chat_id", Value: 1}},
		Options: options.Index().SetUnique(true),
	}

	if _, err := r.chats.Indexes().CreateOne(ctx, chatIndex); err != nil {
		return err
	}

	// Составной индекс для message_id и chat_id в коллекции messages
	messageIndex := mongo.IndexModel{
		Keys: bson.D{
			{Key: "message_id", Value: 1},
			{Key: "chat_id", Value: 1},
		},
		Options: options.Index().SetUnique(true),
	}

	if _, err := r.messages.Indexes().CreateOne(ctx, messageIndex); err != nil {
		return err
	}

	// Индекс для поиска по chat_id в сообщениях
	chatIDIndex := mongo.IndexModel{
		Keys: bson.D{{Key: "chat_id", Value: 1}},
	}

	if _, err := r.messages.Indexes().CreateOne(ctx, chatIDIndex); err != nil {
		return err
	}
	return nil
}

// SaveChat сохраняет чат в MongoDB
func (r *MongoRepository) SaveChat(ctx context.Context, chat *models.ChatDocument) error {
	now := time.Now()
	chat.CreatedAt = now
	chat.UpdatedAt = now

	// Используем upsert для избежания дублирования
	filter := bson.M{"chat_id": chat.ChatID}
	update := bson.M{
		"$setOnInsert": bson.M{
			"chat_id":    chat.ChatID,
			"type":       chat.Type,
			"title":      chat.Title,
			"username":   chat.Username,
			"first_name": chat.FirstName,
			"last_name":  chat.LastName,
			"created_at": now,
		},
		"$set": bson.M{
			"updated_at": now,
		},
	}

	opts := options.Update().SetUpsert(true)
	_, err := r.chats.UpdateOne(ctx, filter, update, opts)
	return err
}

// SaveMessage сохраняет сообщение в MongoDB
func (r *MongoRepository) SaveMessage(ctx context.Context, message *models.MessageDocument) error {
	now := time.Now()
	message.CreatedAt = now

	_, err := r.messages.InsertOne(ctx, message)
	return err
}

// ChatExists проверяет, существует ли чат
func (r *MongoRepository) ChatExists(ctx context.Context, chatID int64) (bool, error) {
	filter := bson.M{"chat_id": chatID}
	count, err := r.chats.CountDocuments(ctx, filter)
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

// GetRecentMessages получает сообщения за последние N дней
func (r *MongoRepository) GetRecentMessages(ctx context.Context, chatID int64, days int) ([]*models.MessageDocument, error) {
	// Вычисляем дату начала периода
	since := time.Now().AddDate(0, 0, -days)

	// Создаем фильтр для поиска
	filter := bson.M{
		"chat_id": chatID,
		"date": bson.M{
			"$gte": since,
		},
	}

	// Сортируем по дате (старые сообщения первыми)
	opts := options.Find().SetSort(bson.D{{Key: "date", Value: 1}})

	cursor, err := r.messages.Find(ctx, filter, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var messages []*models.MessageDocument
	if err := cursor.All(ctx, &messages); err != nil {
		return nil, err
	}

	return messages, nil
}

// Close закрывает соединение с MongoDB
func (r *MongoRepository) Close(ctx context.Context) error {
	return r.client.Disconnect(ctx)
}
