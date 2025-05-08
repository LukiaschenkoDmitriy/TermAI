package response

import (
	"encoding/json"
	"fmt"
)

type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type Choice struct {
	Index   int     `json:"index"`
	Message Message `json:"message"`
}

type Response struct {
	ID      string   `json:"id"`
	Object  string   `json:"object"`
	Created int64    `json:"created"`
	Model   string   `json:"model"`
	Choices []Choice `json:"choices"`
}

func ParseResponse(response []byte) (*Response, error) {
	var result Response
	if err := json.Unmarshal(response, &result); err != nil {
		return nil, fmt.Errorf("failed to parse response: %v", err)
	}

	if len(result.Choices) == 0 {
		return nil, fmt.Errorf("no choices in response")
	}

	return &result, nil
} 