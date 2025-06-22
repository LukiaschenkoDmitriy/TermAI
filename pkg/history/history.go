package history

import (
	"os"
	"path/filepath"

	"github.com/LukiaschenkoDmitriy/TermAI/pkg/http/response"
	"github.com/spf13/viper"
)

type ClientMessage struct {
	Role string `json:"role"`
	Content string `json:"content"`
}

type History struct {
	ConfigPath string `json:"config_path"`
	Messages []Message `json:"history"`
}

type Message struct {
	UserMessage ClientMessage
	AIMessage response.Message
	CommandOutput string
	Cropped bool
}

func New() *History {
	return &History{
		Messages: []Message{},
	}
}

func (h *History) Load() error {
	viper.SetConfigName("config")
	viper.SetConfigType("json")

	home, err := os.UserHomeDir()
	if err != nil {
		return err
	}

	configPath := filepath.Join(home, ".termai")
	h.ConfigPath = filepath.Join(configPath, "config.json")

	if err := os.MkdirAll(configPath, 0755); err != nil {
		return err
	}

	viper.AddConfigPath(configPath)

	if _, err := os.Stat(h.ConfigPath); os.IsNotExist(err) {
		if err := h.saveDefault(); err != nil {
			return err
		}
	}

	if err := viper.ReadInConfig(); err != nil {
		return err
	}

	var messages []Message
	if err := viper.UnmarshalKey("history", &messages); err != nil {
		return err
	}
	h.Messages = messages

	return nil
}

func (h *History) saveDefault() error {
	dir := filepath.Dir(h.ConfigPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	viper.Set("history", []Message{})
	return viper.WriteConfigAs(h.ConfigPath)
}

func (h *History) IsSystemRulesExists() bool {
	for _, message := range h.Messages {
		if message.UserMessage.Role == "system" {
			return true
		}
	}

	return false
}

func (h *History) DeleteByIndex(index int) error {
	h.Messages = append(h.Messages[:index], h.Messages[index+1:]...)

	viper.Set("history", h.Messages)
	if err := viper.WriteConfigAs(h.ConfigPath); err != nil {
		return err
	}
	return nil
}

func (h *History) DeleteByIndexRange(firstIndex int, lastIndex int) error {
	h.Messages = append(h.Messages[:firstIndex], h.Messages[lastIndex+1:]...)

	viper.Set("history", h.Messages)
	if err := viper.WriteConfigAs(h.ConfigPath); err != nil {
		return err
	}
	return nil
}

func (h *History) AddMessage(userMessage ClientMessage, aiMessage response.Message, commandOutput string, cropped bool) error {
	h.Messages = append(h.Messages, Message{
		UserMessage: userMessage,
		AIMessage:   aiMessage,
		CommandOutput: commandOutput,
		Cropped: cropped,
	})

	viper.Set("history", h.Messages)
	if err := viper.WriteConfigAs(h.ConfigPath); err != nil {
		return err
	}
	return nil
}

func (h *History) AddContext(context string) error {
	h.Messages = append(h.Messages, Message{
		UserMessage: ClientMessage{
			Role: "system",
			Content: context,
		},
		AIMessage: response.Message{
			Role: "system",
			Content: "",
		},
		CommandOutput: context,
		Cropped: false,
	})

	viper.Set("history", h.Messages)
	if err := viper.WriteConfigAs(h.ConfigPath); err != nil {
		return err
	}
	return nil
}

func (h *History) Save() error {
	viper.Set("history", h.Messages)
	if err := viper.WriteConfigAs(h.ConfigPath); err != nil {
		return err
	}
	return nil
}


func (h *History) ClearHistory() error {
	h.Messages = []Message{}
	viper.Set("history", h.Messages)
	if err := viper.WriteConfigAs(h.ConfigPath); err != nil {
		return err
	}
	return nil
}

