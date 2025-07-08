package handler

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestWebhookHandler_containsTolikVariation(t *testing.T) {
	handler := &WebhookHandler{
		botName: "Жорик",
	}

	tests := []struct {
		name     string
		message  string
		expected bool
	}{
		{
			name:     "Message with bot name",
			message:  "Привет, Жорик!",
			expected: true,
		},
		{
			name:     "Message with bot name in different case",
			message:  "жорик, как дела?",
			expected: true,
		},
		{
			name:     "Message with bot name in middle",
			message:  "Эй Жорик расскажи анекдот",
			expected: true,
		},
		{
			name:     "Message without bot name",
			message:  "Привет всем!",
			expected: false,
		},
		{
			name:     "Empty message",
			message:  "",
			expected: false,
		},
		{
			name:     "Message with similar name",
			message:  "Анатолий хороший человек",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := handler.containsTolikVariation(tt.message)
			assert.Equal(t, tt.expected, result, "Expected %v for message: %s", tt.expected, tt.message)
		})
	}
}
