#!/bin/bash

echo "Тестирование webhook..."

echo "1. Обычное сообщение (не адресованное боту):"
curl -X POST http://localhost:8080/webhook \
  -H "Content-Type: application/json" \
  -d '{
    "update_id": 123456789,
    "message": {
      "message_id": 1,
      "from": {
        "id": 123456,
        "is_bot": false,
        "first_name": "Семён",
        "last_name": "Тестов",
        "username": "stest"
      },
      "chat": {
        "id": 123456,
        "type": "private"
      },
      "date": 1640995200,
      "text": "Привет всем! Как дела?"
    }
  }' \
  -w "\nHTTP Status: %{http_code}\n\n"

echo "2. Сообщение, адресованное боту:"
curl -X POST http://localhost:8080/webhook \
  -H "Content-Type: application/json" \
  -d '{
    "update_id": 123456790,
    "message": {
      "message_id": 2,
      "from": {
        "id": 123456,
        "is_bot": false,
        "first_name": "Семён",
        "last_name": "Тестов",
        "username": "stest"
      },
      "chat": {
        "id": 123456,
        "type": "private"
      },
      "date": 1640995260,
      "text": "Толик, расскажи анекдот!"
    }
  }' \
  -w "\nHTTP Status: %{http_code}\n"
