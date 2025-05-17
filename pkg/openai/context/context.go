package context

import (
	"strings"

	"github.com/LukiaschenkoDmitriy/TermAI/pkg/history"
)

type Context struct {
	History history.History
}

func New() *Context {

	history := history.New();
	history.Load();

	return &Context{
		History: *history,
	}
}

func (c *Context) GetFirstNotCroppedMessage() (history.Message, int) {
	for i := 0; i < len(c.History.Messages); i++ {
		if !c.History.Messages[i].Cropped {
			return c.History.Messages[i], i
		}
	}

	return history.Message{}, -1
}

func (c *Context) CalculateMessageTokens(message history.Message) int {
	words := len(strings.Fields(message.CommandOutput)) + len(strings.Fields(message.AIMessage.Content)) + len(strings.Fields(message.UserMessage.Content))
	return int(float64(words) * 1.3)
}

func (c *Context) CalculateContextTokens() int {
	tokens := 0;

	for _, message := range c.History.Messages {
		tokens += c.CalculateMessageTokens(message)
	}

	return int(tokens);
}