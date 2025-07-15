package rag

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestExtractSearchKeywords(t *testing.T) {
	service := &Service{}

	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "Simple question",
			input:    "Как дела с проектом?",
			expected: "дела проектом",
		},
		{
			name:     "Question with bot mention",
			input:    "Жорик, что ты думаешь о погоде?",
			expected: "думаешь погоде",
		},
		{
			name:     "Complex question",
			input:    "Помнишь, мы говорили о том Docker контейнере?",
			expected: "помнишь говорили docker контейнере",
		},
		{
			name:     "Empty message",
			input:    "",
			expected: "",
		},
		{
			name:     "Only stop words",
			input:    "что как где когда",
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := service.extractSearchKeywords(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestRemoveBotMentions(t *testing.T) {
	service := &Service{}

	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "Remove Жорик",
			input:    "Жорик, как дела?",
			expected: " как дела?",
		},
		{
			name:     "Remove Жора",
			input:    "Привет, Жора!",
			expected: "Привет, !",
		},
		{
			name:     "No mentions",
			input:    "Обычное сообщение",
			expected: "обычное сообщение",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := service.removeBotMentions(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}
