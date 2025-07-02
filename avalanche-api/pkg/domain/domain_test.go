package domain

import (
	"testing"

	"github.com/snowplow-incubator/avalanche-api/pkg/errors"
)

func TestExtractEventID(t *testing.T) {
	tests := []struct {
		name        string
		event       map[string]string
		eventIndex  int
		expected    string
		expectError bool
	}{
		{
			name:        "valid event ID",
			event:       map[string]string{"eventId": "test-123"},
			eventIndex:  0,
			expected:    "test-123",
			expectError: false,
		},
		{
			name:        "missing event ID",
			event:       map[string]string{"otherField": "value"},
			eventIndex:  5,
			expected:    "",
			expectError: true,
		},
		{
			name:        "empty event ID",
			event:       map[string]string{"eventId": ""},
			eventIndex:  2,
			expected:    "",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := ExtractEventID(tt.event, tt.eventIndex)

			if tt.expectError {
				if err == nil {
					t.Errorf("expected error but got none")
				}
				var valErr *errors.ValidationError
				if !errors.IsValidationError(err, &valErr) {
					t.Errorf("expected ValidationError, got %T", err)
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
				if result != tt.expected {
					t.Errorf("expected %q, got %q", tt.expected, result)
				}
			}
		})
	}
}

func TestExtractAttributes(t *testing.T) {
	tests := []struct {
		name        string
		event       map[string]string
		extractKeys []string
		expected    map[string]string
	}{
		{
			name: "extract existing keys",
			event: map[string]string{
				"userId":    "123",
				"sessionId": "abc",
				"ignored":   "value",
			},
			extractKeys: []string{"userId", "sessionId"},
			expected: map[string]string{
				"userId":    "123",
				"sessionId": "abc",
			},
		},
		{
			name: "skip missing keys",
			event: map[string]string{
				"userId": "123",
			},
			extractKeys: []string{"userId", "missing"},
			expected: map[string]string{
				"userId": "123",
			},
		},
		{
			name: "skip empty values",
			event: map[string]string{
				"userId":  "123",
				"emptyId": "",
			},
			extractKeys: []string{"userId", "emptyId"},
			expected: map[string]string{
				"userId": "123",
			},
		},
		{
			name:        "empty extract keys",
			event:       map[string]string{"userId": "123"},
			extractKeys: []string{},
			expected:    map[string]string{},
		},
		{
			name:        "nil extract keys",
			event:       map[string]string{"userId": "123"},
			extractKeys: nil,
			expected:    map[string]string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ExtractAttributes(tt.event, tt.extractKeys)

			if len(result) != len(tt.expected) {
				t.Errorf("expected %d attributes, got %d", len(tt.expected), len(result))
			}

			for key, expectedValue := range tt.expected {
				if actualValue, exists := result[key]; !exists {
					t.Errorf("expected key %q not found", key)
				} else if actualValue != expectedValue {
					t.Errorf("for key %q: expected %q, got %q", key, expectedValue, actualValue)
				}
			}

			for key := range result {
				if _, expected := tt.expected[key]; !expected {
					t.Errorf("unexpected key %q found", key)
				}
			}
		})
	}
}
