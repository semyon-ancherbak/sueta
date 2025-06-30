# Sueta Telegram Bot

Telegram-бот на Go 1.24, который ведёт осмысленный диалог и хранит память в MongoDB.

## 🚀 Быстрый старт

### Локальный запуск

1. **Клонируйте репозиторий:**
```bash
git clone https://github.com/semyon-ancherbak/sueta.git
cd sueta
```

2. **Настройте переменные окружения:**
```bash
cp .env.example .env
# Отредактируйте .env файл, добавив ваши токены и настройки
```

3. **Установите зависимости:**
```bash
go mod tidy
```

4. **Запустите бота:**
```bash
go run ./cmd/bot
```

Бот будет запущен на порту 8080. Webhook endpoint: `http://localhost:8080/webhook`

### Запуск через Docker

```bash
# Создайте .env файл с настройками
cp .env.example .env

# Запустите через docker-compose
docker-compose up --build
```

## 📝 Тестирование webhook

Для тестирования webhook можно использовать curl:

```bash
curl -X POST http://localhost:8080/webhook \
  -H "Content-Type: application/json" \
  -d '{
    "update_id": 123456789,
    "message": {
      "message_id": 1,
      "from": {
        "id": 123456,
        "is_bot": false,
        "first_name": "Test",
        "username": "testuser"
      },
      "chat": {
        "id": 123456,
        "type": "private"
      },
      "date": 1640995200,
      "text": "Привет, бот!"
    }
  }'
```

## 🧪 Тесты

Запуск всех тестов:
```bash
go test ./...
```

Запуск тестов с подробным выводом:
```bash
go test -v ./...
```

## 📂 Структура проекта

```
├── cmd/bot/              # Точка входа приложения
├── internal/             # Внутренняя логика
│   ├── handler/          # HTTP/Telegram webhook handlers
│   ├── memory/           # Интерфейсы и реализация памяти (TODO)
│   ├── llm/              # Обёртка для LLM запросов (TODO)
│   └── config/           # Конфигурация
├── pkg/                  # Публичные пакеты (TODO)
├── Dockerfile            # Docker образ
├── docker-compose.yml    # Docker Compose конфигурация
└── .env.example          # Пример переменных окружения
```

## ⚙️ Конфигурация

Все настройки задаются через переменные окружения:

| Переменная | Описание | Значение по умолчанию |
|------------|----------|----------------------|
| `PORT` | Порт HTTP сервера | `8080` |
| `TELEGRAM_TOKEN` | Токен Telegram бота | - |
| `WEBHOOK_URL` | URL для webhook | - |
| `MONGO_URL` | URL MongoDB | `mongodb://localhost:27017` |
| `MONGO_DATABASE` | Имя базы данных | `sueta` |
| `OPENROUTER_API_KEY` | Ключ OpenRouter API | - |

## 🔧 Разработка

Текущая функциональность:
- ✅ Получение webhook от Telegram
- ✅ Парсинг сообщений
- ✅ Вывод информации в консоль
- ✅ HTTP сервер с middleware
- ✅ Базовые тесты
- ✅ Docker контейнеризация

Планируется:
- [ ] Интеграция с MongoDB
- [ ] Система памяти (short-term/long-term)
- [ ] Интеграция с LLM (OpenRouter)
- [ ] Генерация ответов
- [ ] Отправка сообщений обратно в Telegram

## 📋 TODO

- [ ] Добавить интерфейсы для памяти (`internal/memory/`)
- [ ] Реализовать MongoDB store с TTL индексами
- [ ] Добавить LLM клиент (`internal/llm/`)
- [ ] Настроить Telegram API для отправки сообщений
- [ ] Добавить CI/CD pipeline
- [ ] Добавить больше тестов
