# Этап сборки
FROM golang:1.24-alpine AS builder

# Устанавливаем необходимые пакеты
RUN apk add --no-cache git ca-certificates

# Устанавливаем рабочую директорию
WORKDIR /app

# Копируем файлы go.mod и go.sum
COPY go.mod go.sum ./

# Загружаем зависимости
RUN go mod download

# Копируем исходный код
COPY . .

# Собираем приложение
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o bot ./cmd/bot

# Финальный этап
FROM alpine:latest

# Устанавливаем сертификаты и timezone
RUN apk --no-cache add ca-certificates tzdata

# Создаем пользователя для запуска приложения
RUN adduser -D -s /bin/sh appuser

# Устанавливаем рабочую директорию
WORKDIR /root/

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
