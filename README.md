# Sueta - Умный Telegram бот с RAG

Telegram-бот на Go 1.24 с функцией RAG (Retrieval-Augmented Generation) для долгосрочной памяти и осмысленного диалога.

## 🚀 Особенности

- **Долгосрочная память**: RAG система для поиска релевантных старых сообщений
- **Кратковременная память**: хранение последних сообщений за 3 дня
- **MongoDB**: надежное хранение данных с полнотекстовым поиском
- **OpenRouter**: интеграция с современными LLM моделями
- **Docker**: простое развертывание с docker-compose
- **Импорт данных**: утилита для загрузки истории чатов из экспорта Telegram

## 📂 Структура проекта

```
├── cmd/
│   ├── bot/           # Основное приложение бота
│   └── importer/      # Утилита импорта данных из Telegram
├── internal/
│   ├── config/        # Конфигурация
│   ├── handler/       # HTTP/Webhook обработчики
│   ├── llm/           # LLM клиент (OpenRouter)
│   ├── memory/        # Интерфейсы памяти (не реализовано)
│   ├── models/        # Модели данных
│   ├── rag/           # RAG сервис
│   ├── repository/    # MongoDB репозиторий
│   └── telegram/      # Telegram клиент
├── docs/              # Документация
└── pkg/               # Публичные пакеты (пусто)
```

## 🛠️ Быстрый старт

### 1. Клонирование и настройка

```bash
git clone https://github.com/semyon-ancherbak/sueta.git
cd sueta
cp .env.example .env
```

### 2. Настройка переменных окружения

Отредактируйте `.env` файл:

```bash
# Telegram
TELEGRAM_TOKEN=your_bot_token_here
WEBHOOK_URL=https://yourdomain.com/webhook

# MongoDB
MONGO_URL=mongodb://localhost:27017
MONGO_DATABASE=sueta

# OpenRouter (LLM)
OPENROUTER_API_KEY=your_openrouter_key_here

# RAG настройки
RAG_MAX_RELEVANT_MESSAGES=5
RAG_RECENT_DAYS_EXCLUDE=3

# Сервер
PORT=8080
```

### 3. Запуск с Docker

```bash
# Запуск всех сервисов
docker-compose up -d

# Просмотр логов
docker-compose logs -f bot
```

### 4. Локальная разработка

```bash
# Установка зависимостей
go mod download

# Запуск MongoDB
docker run -d --name mongo -p 27017:27017 mongo:latest

# Запуск бота
go run cmd/bot/main.go
```

## 📥 Импорт данных из Telegram

Для улучшения RAG можно импортировать историю чатов:

### 1. Экспорт из Telegram Desktop

1. Откройте Telegram Desktop
2. Settings → Advanced → Export Telegram data
3. Выберите чат и формат JSON
4. Убедитесь, что включены Messages

### 2. Импорт в базу данных

```bash
# Переход в директорию импортера
cd cmd/importer

# Пробный запуск (только статистика)
make import-dry EXPORT_FILE=/path/to/export.json

# Реальный импорт
make import EXPORT_FILE=/path/to/export.json

# С подробным выводом
make import-verbose EXPORT_FILE=/path/to/export.json
```

Подробная документация: [cmd/importer/README.md](cmd/importer/README.md)

## 🧠 Как работает RAG

1. **Поиск контекста**: При получении сообщения система ищет релевантные старые сообщения по ключевым словам
2. **Комбинирование**: Найденные сообщения добавляются к недавней истории
3. **Генерация ответа**: LLM получает расширенный контекст и генерирует более осмысленный ответ

### Пример работы

```
Пользователь: "Жорик, помнишь наш разговор о Docker?"

1. Система ищет старые сообщения с "Docker"
2. Находит: "месяц назад обсуждали docker-compose..."
3. LLM получает этот контекст + недавние сообщения
4. Ответ: "Да, помню! Мы говорили о docker-compose для..."
```

## 🔧 Разработка

### Требования

- Go 1.24+
- MongoDB 4.4+
- Docker & docker-compose (опционально)

### Запуск тестов

```bash
# Все тесты
go test ./...

# Тесты с покрытием
go test -cover ./...

# Тесты импортера
go test ./cmd/importer/...

# Тесты RAG
go test ./internal/rag/...
```

### Линтинг

```bash
golangci-lint run
```

### Сборка

```bash
# Основной бот
go build -o bin/bot cmd/bot/main.go

# Импортер
go build -o bin/importer cmd/importer/main.go
```

## 📚 Документация

- [RAG система](docs/rag.md) - подробное описание работы RAG
- [Импорт данных](cmd/importer/README.md) - руководство по импорту
- [Changelog RAG](CHANGELOG_RAG.md) - история изменений RAG

## 🐳 Docker

### Разработка

```bash
# Сборка образа
docker build -t sueta .

# Запуск с MongoDB
docker-compose up
```

### Продакшн

```bash
# Только production сервисы
docker-compose -f docker-compose.prod.yml up -d
```

## 🔐 Безопасность

- Все секреты хранятся в `.env` файле
- `.env` добавлен в `.gitignore`
- MongoDB индексы защищают от дублирования данных
- Валидация входящих webhook-ов

## 📈 Производительность

- **Полнотекстовый поиск**: MongoDB text index для быстрого поиска
- **Индексирование**: Составные индексы для оптимизации запросов
- **Пагинация**: Ограничение количества результатов RAG
- **TTL**: Автоматическое удаление старых записей (настраиваемо)

## 🤝 Contributing

1. Fork репозитория
2. Создайте feature branch: `git checkout -b feature/new-feature`
3. Сделайте изменения и добавьте тесты
4. Запустите тесты: `go test ./...`
5. Commit: `git commit -am 'Add new feature'`
6. Push: `git push origin feature/new-feature`
7. Создайте Pull Request

## 📄 Лицензия

MIT License. См. файл [LICENSE](LICENSE) для деталей.

## 🆘 Поддержка

- GitHub Issues: [github.com/semyon-ancherbak/sueta/issues](https://github.com/semyon-ancherbak/sueta/issues)
- Telegram: @semyon_ancherbak

---

**Примечание**: Проект находится в активной разработке. API может изменяться.
