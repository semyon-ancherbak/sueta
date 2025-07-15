# Этап сборки
FROM golang:1.24-alpine AS builder

# Устанавливаем необходимые пакеты для SQLite
RUN apk add --no-cache git ca-certificates gcc musl-dev sqlite-dev

# Устанавливаем рабочую директорию
WORKDIR /app

# Копируем файлы go.mod и go.sum
COPY go.mod go.sum ./

# Загружаем зависимости
RUN go mod download

# Копируем исходный код
COPY . .

# Собираем приложение с поддержкой CGO для SQLite
RUN CGO_ENABLED=1 GOOS=linux go build -a -installsuffix cgo -o bot ./cmd/bot

# Финальный этап
FROM alpine:latest

# Устанавливаем сертификаты, timezone и SQLite
RUN apk --no-cache add ca-certificates tzdata sqlite

# Создаем пользователя для запуска приложения
RUN adduser -D -s /bin/sh appuser

# Создаем директорию для данных
RUN mkdir -p /app/data && chown appuser:appuser /app/data

# Устанавливаем рабочую директорию
WORKDIR /app

# Копируем собранное приложение
COPY --from=builder /app/bot .

# Меняем владельца файла
RUN chown appuser:appuser bot

# Переключаемся на пользователя appuser
USER appuser

# Открываем порт
EXPOSE 8080

# Запускаем приложение
CMD ["./bot"]
