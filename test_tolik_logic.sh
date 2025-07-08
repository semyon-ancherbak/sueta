#!/bin/bash

WEBHOOK_URL="http://localhost:8080/webhook/AAHRCKPB-W9jAoDg9ISGj880nqjqe4Q4FKA"

echo "Тестируем новую логику проверки сообщений для бота..."
echo

# Тест 1: Сообщение с именем "толик"
echo "=== Тест 1: Сообщение 'Толик, привет!' ==="
curl -X POST "$WEBHOOK_URL" \
  -H "Content-Type: application/json" \
  -d '{
    "update_id": 1001,
    "message": {
      "message_id": 1,
      "from": {
        "id": 123456789,
        "is_bot": false,
        "first_name": "Тестовый",
        "username": "testuser"
      },
      "chat": {
        "id": -987654321,
        "type": "group",
        "title": "Тестовая группа"
      },
      "date": 1641024000,
      "text": "Толик, привет!"
    }
  }'
echo -e "\n"

sleep 2

# Тест 2: Сообщение со склонением "толику"
echo "=== Тест 2: Сообщение 'Спасибо толику за помощь' ==="
curl -X POST "$WEBHOOK_URL" \
  -H "Content-Type: application/json" \
  -d '{
    "update_id": 1002,
    "message": {
      "message_id": 2,
      "from": {
        "id": 123456789,
        "is_bot": false,
        "first_name": "Тестовый",
        "username": "testuser"
      },
      "chat": {
        "id": -987654321,
        "type": "group",
        "title": "Тестовая группа"
      },
      "date": 1641024060,
      "text": "Спасибо толику за помощь"
    }
  }'
echo -e "\n"

sleep 2

# Тест 3: Сообщение БЕЗ имени (не должен реагировать)
echo "=== Тест 3: Сообщение 'Помоги мне' (без обращения к боту) ==="
curl -X POST "$WEBHOOK_URL" \
  -H "Content-Type: application/json" \
  -d '{
    "update_id": 1003,
    "message": {
      "message_id": 3,
      "from": {
        "id": 123456789,
        "is_bot": false,
        "first_name": "Тестовый",
        "username": "testuser"
      },
      "chat": {
        "id": -987654321,
        "type": "group",
        "title": "Тестовая группа"
      },
      "date": 1641024120,
      "text": "Помоги мне"
    }
  }'
echo -e "\n"

sleep 2

# Тест 4: Reply на сообщение бота (эмулируем ответ)
echo "=== Тест 4: Reply на сообщение бота ==="
curl -X POST "$WEBHOOK_URL" \
  -H "Content-Type: application/json" \
  -d '{
    "update_id": 1004,
    "message": {
      "message_id": 4,
      "from": {
        "id": 123456789,
        "is_bot": false,
        "first_name": "Тестовый",
        "username": "testuser"
      },
      "chat": {
        "id": -987654321,
        "type": "group",
        "title": "Тестовая группа"
      },
      "date": 1641024180,
      "text": "Да, спасибо!",
      "reply_to_message": {
        "message_id": 100,
        "from": {
          "id": 987654321,
          "is_bot": true,
          "first_name": "Sueta Bot"
        },
        "chat": {
          "id": -987654321,
          "type": "group"
        },
        "date": 1641024150,
        "text": "Как дела?"
      }
    }
  }'
echo -e "\n"

echo "Все тесты отправлены! Проверьте логи бота."
