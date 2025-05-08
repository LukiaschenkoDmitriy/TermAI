package client_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/LukiaschenkoDmitriy/TermAI/pkg/http/client"
	"github.com/LukiaschenkoDmitriy/TermAI/pkg/http/response"
)

func TestClient_ConvertMessages(t *testing.T) {
	c := client.New()
	messages := []string{"test message 1", "test message 2"}

	converted := c.ConvertMessages(messages)

	if len(converted) != len(messages) {
		t.Errorf("Expected %d messages, got %d", len(messages), len(converted))
	}

	for i, msg := range converted {
		if msg["role"] != "user" {
			t.Errorf("Expected role 'user', got '%s'", msg["role"])
		}
		if msg["content"] != messages[i] {
			t.Errorf("Expected content '%s', got '%s'", messages[i], msg["content"])
		}
	}
}

func TestClient_SendRequest(t *testing.T) {
	// Create a test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Check request headers
		if r.Header.Get("Content-Type") != "application/json" {
			t.Errorf("Expected Content-Type 'application/json', got '%s'", r.Header.Get("Content-Type"))
		}

		// Check request body
		var requestBody map[string]interface{}
		if err := json.NewDecoder(r.Body).Decode(&requestBody); err != nil {
			t.Errorf("Failed to decode request body: %v", err)
		}

		// Send mock response
		mockResponse := response.Response{
			ID:      "test-id",
			Object:  "chat.completion",
			Created: 1234567890,
			Model:   "test-model",
			Choices: []response.Choice{
				{
					Message: response.Message{
						Content: "```json\n{\"answer\": \"Test response\", \"commands\": [], \"error\": \"\"}\n```",
					},
				},
			},
		}

		json.NewEncoder(w).Encode(mockResponse)
	}))
	defer server.Close()

	// Create client with test server URL
	c := client.New()
	c.APIKey = "test-key"
	c.Model = "test-model"

	// Send test request
	response, err := c.SendRequest([]string{"test message"})
	if err != nil {
		t.Errorf("SendRequest failed: %v", err)
	}

	if response == nil {
		t.Error("Expected non-nil response")
	}

	if len(response.Choices) == 0 {
		t.Error("Expected non-empty choices")
	}
}

func TestClient_SendRequest_Error(t *testing.T) {
	// Create a test server that returns an error
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"error": map[string]string{
				"message": "Test error",
				"type":    "invalid_request",
				"code":    "test_error",
			},
		})
	}))
	defer server.Close()

	// Create client with test server URL
	c := client.New()
	c.APIKey = "test-key"
	c.Model = "test-model"

	// Send test request
	_, err := c.SendRequest([]string{"test message"})
	if err == nil {
		t.Error("Expected error, got nil")
	}
} 