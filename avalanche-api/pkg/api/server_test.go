package api

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"testing"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/snowplow-incubator/avalanche-api/pkg/config"
	"github.com/snowplow-incubator/avalanche-api/pkg/domain"
	"github.com/snowplow-incubator/avalanche-api/pkg/services"
)

type mockProfileService struct {
	processEventFunc  func(ctx context.Context, event map[string]string) (*domain.Profile, error)
	processEventsFunc func(ctx context.Context, events []map[string]string) ([]services.EventResult, error)
}

func (m *mockProfileService) ProcessEvent(ctx context.Context, event map[string]string) (*domain.Profile, error) {
	if m.processEventFunc != nil {
		return m.processEventFunc(ctx, event)
	}
	return &domain.Profile{
		ID:        uuid.New(),
		CreatedAt: time.Now(),
	}, nil
}

func (m *mockProfileService) ProcessEvents(ctx context.Context, events []map[string]string) ([]services.EventResult, error) {
	if m.processEventsFunc != nil {
		return m.processEventsFunc(ctx, events)
	}

	results := make([]services.EventResult, len(events))
	for i := range events {
		results[i] = services.EventResult{
			Profile: &domain.Profile{
				ID:        uuid.New(),
				CreatedAt: time.Now(),
			},
		}
	}
	return results, nil
}

func TestServer_healthHandler(t *testing.T) {
	server := &Server{
		app:            fiber.New(),
		config:         &config.Config{},
		profileService: &mockProfileService{},
	}
	server.setupRoutes()

	req, _ := http.NewRequest("GET", "/health", nil)
	resp, err := server.app.Test(req)
	if err != nil {
		t.Fatalf("failed to test health endpoint: %v", err)
	}

	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected status 200, got %d", resp.StatusCode)
	}
}

func TestServer_eventsHandler(t *testing.T) {
	tests := []struct {
		name           string
		requestBody    interface{}
		setupService   func(*mockProfileService)
		setupConfig    func(*config.Config)
		expectedStatus int
		checkResponse  func(t *testing.T, body []byte)
	}{
		{
			name:           "valid events request",
			requestBody:    []map[string]string{{"eventId": "test-1", "userId": "123"}},
			setupService:   func(ms *mockProfileService) {}, // Use default
			setupConfig:    func(cfg *config.Config) { cfg.MaxRequestSize = 1000 },
			expectedStatus: http.StatusOK,
			checkResponse: func(t *testing.T, body []byte) {
				var response map[string]interface{}
				if err := json.Unmarshal(body, &response); err != nil {
					t.Fatalf("failed to unmarshal response: %v", err)
				}
				if results, ok := response["results"].([]interface{}); !ok || len(results) != 1 {
					t.Error("expected results array with 1 item")
				}
			},
		},
		{
			name:           "empty events request",
			requestBody:    []map[string]string{},
			setupService:   func(ms *mockProfileService) {},
			setupConfig:    func(cfg *config.Config) { cfg.MaxRequestSize = 1000 },
			expectedStatus: http.StatusOK,
			checkResponse: func(t *testing.T, body []byte) {
				var response map[string]interface{}
				if err := json.Unmarshal(body, &response); err != nil {
					t.Fatalf("failed to unmarshal response: %v", err)
				}
				if results, ok := response["results"].([]interface{}); !ok || len(results) != 0 {
					t.Error("expected empty results array")
				}
			},
		},
		{
			name:           "invalid JSON request",
			requestBody:    "invalid json",
			setupService:   func(ms *mockProfileService) {},
			setupConfig:    func(cfg *config.Config) { cfg.MaxRequestSize = 1000 },
			expectedStatus: http.StatusBadRequest,
			checkResponse: func(t *testing.T, body []byte) {
				var response map[string]interface{}
				if err := json.Unmarshal(body, &response); err != nil {
					t.Fatalf("failed to unmarshal response: %v", err)
				}
				if _, ok := response["error"]; !ok {
					t.Error("expected error field in response")
				}
			},
		},
		{
			name: "request too large",
			requestBody: []map[string]string{
				{"eventId": "test-1", "userId": "123"},
				{"eventId": "test-2", "userId": "456"},
				{"eventId": "test-3", "userId": "789"},
			},
			setupService:   func(ms *mockProfileService) {},
			setupConfig:    func(cfg *config.Config) { cfg.MaxRequestSize = 2 },
			expectedStatus: http.StatusBadRequest,
			checkResponse: func(t *testing.T, body []byte) {
				var response map[string]interface{}
				if err := json.Unmarshal(body, &response); err != nil {
					t.Fatalf("failed to unmarshal response: %v", err)
				}
				if errorMsg, ok := response["error"].(string); !ok {
					t.Error("expected error message")
				} else if errorMsg == "" {
					t.Error("expected non-empty error message")
				}
			},
		},
		{
			name:        "service error",
			requestBody: []map[string]string{{"eventId": "test-1", "userId": "123"}},
			setupService: func(ms *mockProfileService) {
				ms.processEventsFunc = func(ctx context.Context, events []map[string]string) ([]services.EventResult, error) {
					return nil, fiber.NewError(fiber.StatusInternalServerError, "service error")
				}
			},
			setupConfig:    func(cfg *config.Config) { cfg.MaxRequestSize = 1000 },
			expectedStatus: http.StatusInternalServerError,
			checkResponse: func(t *testing.T, body []byte) {
				var response map[string]interface{}
				if err := json.Unmarshal(body, &response); err != nil {
					t.Fatalf("failed to unmarshal response: %v", err)
				}
				if _, ok := response["error"]; !ok {
					t.Error("expected error field in response")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			mockService := &mockProfileService{}
			tt.setupService(mockService)

			cfg := &config.Config{}
			tt.setupConfig(cfg)

			server := &Server{
				app:            fiber.New(),
				config:         cfg,
				profileService: mockService,
			}
			server.setupRoutes()

			var requestBody []byte
			var err error
			if str, ok := tt.requestBody.(string); ok {
				requestBody = []byte(str)
			} else {
				requestBody, err = json.Marshal(tt.requestBody)
				if err != nil {
					t.Fatalf("failed to marshal request body: %v", err)
				}
			}

			req, _ := http.NewRequest("POST", "/events", bytes.NewReader(requestBody))
			req.Header.Set("Content-Type", "application/json")

			resp, err := server.app.Test(req)
			if err != nil {
				t.Fatalf("failed to test events endpoint: %v", err)
			}

			if resp.StatusCode != tt.expectedStatus {
				t.Errorf("expected status %d, got %d", tt.expectedStatus, resp.StatusCode)
			}

			if tt.checkResponse != nil {
				body, err := io.ReadAll(resp.Body)
				if err != nil {
					t.Fatalf("failed to read response body: %v", err)
				}
				tt.checkResponse(t, body)
			}
		})
	}
}
