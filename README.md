# Sueta Bot

Telegram бот с использованием LLM для генерации ответов на основе контекста чата.

## Особенности

- Использует SQLite для хранения истории сообщений
- Интеграция с OpenRouter для LLM запросов
- Webhook-based архитектура
- Docker поддержка
- Автоматическое определение сообщений, адресованных боту

## Требования

- Go 1.24+
- SQLite3
- Docker (опционально)

## Установка

### Локальная установка

1. Клонируйте репозиторий:
```bash
git clone https://github.com/semyon-ancherbak/sueta.git
cd sueta
```

2. Установите зависимости:
```bash
go mod download
```

3. Настройте переменные окружения:
```bash
cp .env.example .env
# Отредактируйте .env файл со своими значениями
```

4. Создайте директорию для базы данных:
```bash
mkdir -p data
```

5. Соберите и запустите:
```bash
go build -o bot ./cmd/bot
./bot
```

### Docker

1. Настройте переменные окружения в docker-compose.yml

2. Запустите:
```bash
docker-compose up -d
```

## Конфигурация

Создайте файл `.env` на основе `.env.example`:

- `PORT` - порт для HTTP сервера (по умолчанию: 8080)
- `TELEGRAM_TOKEN` - токен Telegram бота (получите у @BotFather)
- `WEBHOOK_URL` - URL для webhook (ваш публичный домен + /webhook)
- `DATABASE_PATH` - путь к файлу SQLite базы данных (по умолчанию: ./data/sueta.db)
- `OPENROUTER_API_KEY` - ключ API для OpenRouter

## База данных

Проект использует SQLite для хранения:
- Информации о чатах
- Истории сообщений
- Метаданных сообщений

База данных автоматически создается при первом запуске.

## API

### Webhook

`POST /webhook/{token}` - обработка обновлений от Telegram

## Структура проекта

```
.
├── cmd/bot/           # Точка входа приложения
├── internal/
│   ├── config/        # Конфигурация
│   ├── handler/       # HTTP обработчики
│   ├── llm/           # LLM клиент
│   ├── models/        # Модели данных
│   ├── repository/    # Слой данных (SQLite)
│   └── telegram/      # Telegram клиент
├── data/              # Директория для SQLite базы данных
├── Dockerfile         # Docker конфигурация
└── docker-compose.yml # Docker Compose конфигурация
```

Новая версия автоматически создаст SQLite базу данных при первом запуске.
