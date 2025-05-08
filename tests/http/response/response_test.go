package response_test

import (
	"testing"

	"github.com/LukiaschenkoDmitriy/TermAI/pkg/http/response"
)

func TestParseResponse(t *testing.T) {
	// Test valid response
	validJSON := `{
		"id": "test-id",
		"object": "chat.completion",
		"created": 1234567890,
		"model": "test-model",
		"choices": [
			{
				"message": {
					"content": "```json\n{\"answer\": \"Test response\", \"commands\": [\"echo test\"], \"error\": \"\"}\n```"
				}
			}
		]
	}`

	resp, err := response.ParseResponse([]byte(validJSON))
	if err != nil {
		t.Errorf("ParseResponse failed: %v", err)
	}

	if resp == nil {
		t.Error("Expected non-nil response")
	}

	if resp.ID != "test-id" {
		t.Errorf("Expected ID 'test-id', got '%s'", resp.ID)
	}

	if len(resp.Choices) == 0 {
		t.Error("Expected non-empty choices")
	}

	// Test invalid JSON
	invalidJSON := `{invalid json}`
	_, err = response.ParseResponse([]byte(invalidJSON))
	if err == nil {
		t.Error("Expected error for invalid JSON")
	}

	// Test missing required fields
	incompleteJSON := `{
		"id": "test-id",
		"object": "chat.completion"
	}`
	_, err = response.ParseResponse([]byte(incompleteJSON))
	if err == nil {
		t.Error("Expected error for incomplete JSON")
	}
}

func TestParseResponse_Content(t *testing.T) {
	// Test response with commands
	jsonWithCommands := `{
		"choices": [
			{
				"message": {
					"content": "```json\n{\"answer\": \"Test response\", \"commands\": [\"echo test1\", \"echo test2\"], \"error\": \"\"}\n```"
				}
			}
		]
	}`

	resp, err := response.ParseResponse([]byte(jsonWithCommands))
	if err != nil {
		t.Errorf("ParseResponse failed: %v", err)
	}

	if len(resp.Choices) == 0 {
		t.Error("Expected non-empty choices")
	}

	// Test response with error
	jsonWithError := `{
		"choices": [
			{
				"message": {
					"content": "```json\n{\"answer\": \"\", \"commands\": [], \"error\": \"Test error\"}\n```"
				}
			}
		]
	}`

	resp, err = response.ParseResponse([]byte(jsonWithError))
	if err != nil {
		t.Errorf("ParseResponse failed: %v", err)
	}

	if len(resp.Choices) == 0 {
		t.Error("Expected non-empty choices")
	}
} 