package client

import (
	"bytes"
	"fmt"
	"io"
	"net/http"

	"github.com/dmytrii/termai/pkg/config"
)

type Client struct {
	APIKey string
	Model string
	Rules []string
}

func New() *Client {
	config := config.New();
	config.Load();

	return &Client{
		APIKey: config.Settings.APIKey,
		Model:  config.Settings.Model,
		Rules:  config.Settings.Rules,
	}
}

func (c *Client) SendRequest(requestBody []byte) ([]byte, error) {
	url := "https://api.openai.com/v1/chat/completions"

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(requestBody))
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
