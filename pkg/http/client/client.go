package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/LukiaschenkoDmitriy/TermAI/pkg/config"
)

type Client struct {
	APIKey string
	Model string
	Rules []string
}

const (
	OPENAI_URL = "https://api.openai.com/v1/chat/completions"
)

func New() *Client {
	config := config.New();
	config.Load();

	return &Client{
		APIKey: config.Settings.APIKey,
		Model:  config.Settings.Model,
		Rules:  config.Settings.Rules,
	}
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

func (c *Client) SendRequest(messages []string) ([]byte, error) {
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

	return body, nil
}
