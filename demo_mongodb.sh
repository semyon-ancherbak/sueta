#!/bin/bash

echo "🚀 Запуск демонстрации сохранения в MongoDB..."
echo ""

# Проверяем, запущен ли бот
if ! curl -s http://localhost:8080/webhook > /dev/null 2>&1; then
    echo "❌ Бот не запущен. Запустите его командой: go run ./cmd/bot"
    exit 1
fi

echo "✅ Бот запущен, отправляем тестовые сообщения..."
echo ""

# Первое сообщение (создаст новый чат)
echo "📩 Отправляем первое сообщение..."
curl -X POST http://localhost:8080/webhook \
  -H "Content-Type: application/json" \
  -d '{
    "update_id": 100001,
    "message": {
      "message_id": 10,
      "from": {
        "id": 987654,
        "is_bot": false,
        "first_name": "Анна",
        "last_name": "Петрова",
        "username": "anna_p"
      },
      "chat": {
        "id": 987654,
        "type": "private"
      },
      "date": 1640995200,
      "text": "Привет! Это первое сообщение."
    }
  }' > /dev/null 2>&1

sleep 1

# Второе сообщение (в том же чате)
echo "📩 Отправляем второе сообщение в тот же чат..."
curl -X POST http://localhost:8080/webhook \
  -H "Content-Type: application/json" \
  -d '{
    "update_id": 100002,
    "message": {
      "message_id": 11,
      "from": {
        "id": 987654,
        "is_bot": false,
        "first_name": "Анна",
        "last_name": "Петрова",
        "username": "anna_p"
      },
      "chat": {
        "id": 987654,
        "type": "private"
      },
      "date": 1640995260,
      "text": "А это второе сообщение в том же чате."
    }
  }' > /dev/null 2>&1

sleep 1

# Третье сообщение (новый чат)
echo "📩 Отправляем сообщение от другого пользователя (новый чат)..."
curl -X POST http://localhost:8080/webhook \
  -H "Content-Type: application/json" \
  -d '{
    "update_id": 100003,
    "message": {
      "message_id": 20,
      "from": {
        "id": 111222,
        "is_bot": false,
        "first_name": "Иван",
        "username": "ivan_test"
      },
      "chat": {
        "id": 111222,
        "type": "private"
      },
      "date": 1640995320,
      "text": "Привет от Ивана!"
    }
  }' > /dev/null 2>&1

echo ""
echo "✅ Все сообщения отправлены!"
echo "📊 Проверьте логи бота - должно быть:"
echo "   • 2 сохраненных чата (Анна и Иван)"
echo "   • 3 сохраненных сообщения"
echo "   • Для второго сообщения Анны должно быть 'Чат уже существует'"
echo ""
