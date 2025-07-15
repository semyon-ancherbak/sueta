package repository

import (
	"context"
	"time"

	"github.com/semyon-ancherbak/sueta/internal/models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type Repository interface {
	SaveChat(ctx context.Context, chat *models.ChatDocument) error
	SaveMessage(ctx context.Context, message *models.MessageDocument) error
	ChatExists(ctx context.Context, chatID int64) (bool, error)
	GetRecentMessages(ctx context.Context, chatID int64, days int) ([]*models.MessageDocument, error)
	GetLastMessages(ctx context.Context, chatID int64, limit int) ([]*models.MessageDocument, error)
	Close(ctx context.Context) error
}

// MongoRepository реализация Repository для MongoDB
type MongoRepository struct {
	client   *mongo.Client
	database *mongo.Database
	chats    *mongo.Collection
	messages *mongo.Collection
}

func NewMongoRepository(ctx context.Context, mongoURL, dbName string) (*MongoRepository, error) {
	client, err := mongo.Connect(ctx, options.Client().ApplyURI(mongoURL))
	if err != nil {
		return nil, err
	}

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

	// Текстовый индекс для полнотекстового поиска по содержимому сообщений
	textIndex := mongo.IndexModel{
		Keys: bson.D{
			{Key: "text", Value: "text"},
			{Key: "first_name", Value: "text"},
			{Key: "last_name", Value: "text"},
			{Key: "username", Value: "text"},
		},
		Options: options.Index().SetDefaultLanguage("russian"),
	}

	if _, err := r.messages.Indexes().CreateOne(ctx, textIndex); err != nil {
		return err
	}

	return nil
}

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

func (r *MongoRepository) SaveMessage(ctx context.Context, message *models.MessageDocument) error {
	now := time.Now()
	message.CreatedAt = now

	_, err := r.messages.InsertOne(ctx, message)
	return err
}

func (r *MongoRepository) ChatExists(ctx context.Context, chatID int64) (bool, error) {
	filter := bson.M{"chat_id": chatID}
	count, err := r.chats.CountDocuments(ctx, filter)
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

func (r *MongoRepository) GetRecentMessages(
	ctx context.Context,
	chatID int64,
	days int,
) ([]*models.MessageDocument, error) {
	since := time.Now().AddDate(0, 0, -days)

	filter := bson.M{
		"chat_id": chatID,
		"date": bson.M{
			"$gte": since,
		},
	}

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

// GetLastMessages получает последние N сообщений из чата
func (r *MongoRepository) GetLastMessages(
	ctx context.Context,
	chatID int64,
	limit int,
) ([]*models.MessageDocument, error) {
	filter := bson.M{
		"chat_id": chatID,
	}

	// Сортируем по дате в убывающем порядке и ограничиваем количество
	opts := options.Find().
		SetSort(bson.D{{Key: "date", Value: -1}}).
		SetLimit(int64(limit))

	cursor, err := r.messages.Find(ctx, filter, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var messages []*models.MessageDocument
	if err := cursor.All(ctx, &messages); err != nil {
		return nil, err
	}

	// Поскольку мы получили сообщения в убывающем порядке,
	// нужно развернуть их обратно для правильной хронологии
	for i, j := 0, len(messages)-1; i < j; i, j = i+1, j-1 {
		messages[i], messages[j] = messages[j], messages[i]
	}

	return messages, nil
}

func (r *MongoRepository) Close(ctx context.Context) error {
	return r.client.Disconnect(ctx)
}
