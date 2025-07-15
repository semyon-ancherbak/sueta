# Telegram Import Tool Makefile

.PHONY: help import import-dry import-verbose import-test build clean

# Переменные
EXPORT_FILE ?= example_export.json
BINARY_NAME = telegram-importer

help: ## Показать справку
	@echo "Доступные команды:"
	@echo "  import         - Импорт из файла (укажите EXPORT_FILE=path)"
	@echo "  import-dry     - Пробный запуск импорта"
	@echo "  import-verbose - Импорт с подробным выводом"
	@echo "  import-test    - Тест парсинга без БД"
	@echo "  build          - Собрать бинарник"
	@echo "  clean          - Удалить собранные файлы"
	@echo ""
	@echo "Примеры:"
	@echo "  make import EXPORT_FILE=/path/to/export.json"
	@echo "  make import-dry EXPORT_FILE=/path/to/export.json"
	@echo "  make import-test    # использует example_export.json"

import: ## Импорт данных из Telegram экспорта
	@if [ ! -f "$(EXPORT_FILE)" ]; then \
		echo "Файл $(EXPORT_FILE) не найден. Укажите правильный путь через EXPORT_FILE="; \
		exit 1; \
	fi
	go run main.go -file "$(EXPORT_FILE)"

import-dry: ## Пробный запуск импорта (только статистика)
	@if [ ! -f "$(EXPORT_FILE)" ]; then \
		echo "Файл $(EXPORT_FILE) не найден. Укажите правильный путь через EXPORT_FILE="; \
		exit 1; \
	fi
	go run main.go -file "$(EXPORT_FILE)" -dry-run -v

import-verbose: ## Импорт с подробным выводом
	@if [ ! -f "$(EXPORT_FILE)" ]; then \
		echo "Файл $(EXPORT_FILE) не найден. Укажите правильный путь через EXPORT_FILE="; \
		exit 1; \
	fi
	go run main.go -file "$(EXPORT_FILE)" -v

import-test: ## Тест парсинга без подключения к БД
	go run main.go -file "$(EXPORT_FILE)" -dry-run -v --no-db

build: ## Собрать бинарный файл
	go build -o bin/$(BINARY_NAME) cmd/importer/main.go

clean: ## Удалить собранные файлы
	rm -rf bin/

# Показать статус MongoDB
mongo-status: ## Проверить статус MongoDB
	@echo "Проверяем подключение к MongoDB..."
	@docker ps | grep mongo || echo "MongoDB контейнер не запущен"
	@mongo --eval "db.adminCommand('ping')" --quiet 2>/dev/null && echo "MongoDB доступен" || echo "MongoDB недоступен"

# Быстрый старт MongoDB через Docker
mongo-start: ## Запустить MongoDB через Docker
	docker run -d --name sueta-mongo -p 27017:27017 -v sueta-mongo-data:/data/db mongo:latest

mongo-stop: ## Остановить MongoDB контейнер
	docker stop sueta-mongo || true
	docker rm sueta-mongo || true
