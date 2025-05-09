package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/LukiaschenkoDmitriy/TermAI/pkg/config"
	"github.com/LukiaschenkoDmitriy/TermAI/pkg/history"
	"github.com/LukiaschenkoDmitriy/TermAI/pkg/http/response"
)

type APIError struct {
	Error struct {
		Message string `json:"message"`
		Type    string `json:"type"`
		Code    string `json:"code"`
	} `json:"error"`
}

type Message struct {
	Role string `json:"role"`
	Content string `json:"content"`
}

type Client struct {
	APIKey string
	Model string
	Messages []Message
	History history.History
}

const (
	OPENAI_URL = "https://api.openai.com/v1/chat/completions"
)

func New() *Client {
	config := config.New();
	config.Load();

	historyCfg := history.New();
	historyCfg.Load();

	client := &Client{
		APIKey: config.Settings.APIKey,
		Model:  config.Settings.Model,
		Messages: []Message{},
		History: *historyCfg,
	}

	return client
}

func (c *Client) ConvertHistoryToMessages() []Message {
	historyMessages := make([]Message, len(c.History.Messages) * 2)

	for i, message := range c.History.Messages {
		additionalContext := "";

		if (len(message.CommandOutput) > 0) {
			additionalContext = "\n Executed Commands:" + message.CommandOutput;
		}

		historyMessages[i] = Message{
			Role: message.UserMessage.Role,
			Content: additionalContext + message.UserMessage.Content,
		}
		historyMessages[i + len(c.History.Messages)] = Message{
			Role: message.AIMessage.Role,
			Content: message.AIMessage.Content,
		}
	}

	return historyMessages;
}

func (c *Client) AddMessages(role string,messages []string) {
	for _, message := range messages {
		c.Messages = append(c.Messages, Message{
			Role: role,
			Content: message,
		})
	}
}

func (c *Client) SendRequest(messages []string) (*response.Response, error) {
	requestBody := make(map[string]any)

	messages = []string{strings.Join(messages, "\n")};

	c.AddMessages("user", messages);
	
	requestBody["model"] = c.Model
	requestBody["store"] = false
	requestBody["messages"] = append(c.ConvertHistoryToMessages(), c.Messages...)

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

	response, err := response.ParseResponse(body)

	if err != nil {
		return nil, err
	}

	return response, nil
}
