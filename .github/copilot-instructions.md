# GitHub Copilot Instructions

## 🚀 Цели проекта
- Реализовать Telegram-бота `sueta` на Go 1.24, который ведёт осмысленный диалог и хранит кратковременную и долговременную память в MongoDB.
- MVP: понимать контекст беседы, помнить последние сообщения за 3 дня.

## 📂 Репозиторий
- URL: https://github.com/semyon-ancherbak/sueta
- Главная директория:
  - `/cmd/bot` — точка входа приложения
  - `/internal` — внутренняя логика (не экспортируется)
    - `/handler` — HTTP/Telegram webhook
    - `/memory` — интерфейсы и реализация short-/long-term памяти
    - `/llm` — обёртка для запросов к LLM
    - `/config` — загрузка окружения
  - `/pkg` — публичные пакеты (API), которые потенциально могут быть импортированы извне
  - `Dockerfile`, `docker-compose.yml`, `.env.example`

## 🛠️ Технологический стек
- Язык: Go 1.24 (модули Go)
- Модуль: `github.com/semyon-ancherbak/sueta`
- База данных: MongoDB (официальный Go-драйвер)
- Конфигурация: переменные окружения через [godotenv](https://github.com/joho/godotenv)
- Сборка и запуск: Docker & docker-compose
- https://openrouter.ai — для LLM запросов

## 📐 Стиль кода
- Идиоматичный Go: `go fmt`, `golangci-lint`
- Обработка ошибок: всегда проверять `err` и корректно реагировать
- Контексты: прокидывать `context.Context` для возможности отмены/тайм-аута
- Тесты: unit-тесты для ключевых модулей с помощью [testify](https://github.com/stretchr/testify)

## 🔐 Безопасность
- Секреты в `.env`, не в репозитории
- Возможность перезагрузки конфигурации без рестарта через watcher

## 🐳 Docker / CI
- **Dockerfile**: multi-stage сборка на `golang:1.24-alpine`, финальный образ — `scratch` или `alpine`
- **docker-compose**:
  - `bot` с маунтом `.env`
  - `mongo` с volume для данных
- **GitHub Actions**:
  - Шаги: `go fmt` → `golangci-lint` → `go test`
  - Сборка и пуш Docker-образа

## 🤖 Copilot-agent настройки
- **Роли**: «AGENT», понимающий Telegram API, Go 1.24, MongoDB, Docker.
- **Уточнять**: если контекст неясен, спрашивать детали (например, формат LLM-промптом).
- **Генерировать**: код с обработкой ошибок, с мок-интерфейсами для тестов.
- **REST-хендлеры**: на `github.com/go-chi/chi` или `gorilla/mux`.
- **Память**: в реализации short-term хранилища сразу добавлять TTL-индекс на 3 дня.

## 📝 Примеры подсказок для Copilot
1. _“Generate a Go HTTP handler in `handler/webhook.go` that reads Telegram update from JSON, validates, saves it to MongoDB, and returns 200 OK.”_
2. _“Implement `memory.ShortTermStore` using MongoDB with TTL index of 3 days.”_
3. _“Write multi-stage Dockerfile for Go 1.24 application, copy binary into minimal image.”_

---

> **Hint for Copilot**:  
> «You are a helpful coding assistant (agent mode). The repository root is at `https://github.com/semyon-ancherbak/sueta`. Use Go 1.24 modules, handle errors carefully, write unit tests, follow the folder structure above.»
