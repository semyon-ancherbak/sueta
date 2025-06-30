#!/bin/bash

echo "Тестирование webhook..."

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
      "text": "Привет, бот! Как дела?"
    }
  }' \
  -w "\nHTTP Status: %{http_code}\n"
