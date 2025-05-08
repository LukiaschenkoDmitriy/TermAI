package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/LukiaschenkoDmitriy/TermAI/pkg/config"
	"github.com/LukiaschenkoDmitriy/TermAI/pkg/http/response"
)

type APIError struct {
	Error struct {
		Message string `json:"message"`
		Type    string `json:"type"`
		Code    string `json:"code"`
	} `json:"error"`
}

type Client struct {
	APIKey string
	Model string
	systemMessage map[string]any
}

const (
	OPENAI_URL = "https://api.openai.com/v1/chat/completions"
)

func New() *Client {
	config := config.New();
	config.Load();

	client := &Client{
		APIKey: config.Settings.APIKey,
		Model:  config.Settings.Model,
	}

	client.systemMessage = map[string]any{
		"role": "system",
		"content": strings.Join(config.Settings.Rules, "\n"),
	}

	return client
}

func (c * Client) ConvertMessages(messages []string) []map[string]any {
	convertedMessages := make([]map[string]any, len(messages))
	for i, message := range messages {
		convertedMessages[i] = map[string]any{
			"role": "user",
			"content": message,
		}
	}
	return convertedMessages
}

func (c *Client) SendRequest(messages []string) (*response.Response, error) {
	requestBody := make(map[string]any)	
	
	requestBody["model"] = c.Model
	requestBody["store"] = false
	requestBody["messages"] = c.ConvertMessages(messages);

	jsonBody, err := json.Marshal(requestBody)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", OPENAI_URL, bytes.NewBuffer(jsonBody))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.APIKey))

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	// Check for API errors
	if resp.StatusCode != http.StatusOK {
		var apiError APIError
		if err := json.Unmarshal(body, &apiError); err != nil {
			return nil, fmt.Errorf("failed to parse error response: %v", err)
		}
		return nil, fmt.Errorf("API error: %s (type: %s, code: %s)", 
			apiError.Error.Message, 
			apiError.Error.Type, 
			apiError.Error.Code)
	}

	return response.ParseResponse(body)
}
