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

func (c *Context) ConvertMessagesToOneMessage(messages []history.Message) history.Message {
	message := history.Message{}

	for _, message_ := range messages {
		message.CommandOutput += message_.CommandOutput
		message.AIMessage.Content += message_.AIMessage.Content
		message.UserMessage.Content += message_.UserMessage.Content
	}

	return message;
}

func (c *Context) GetNotCoppedMessageByTokens(tokens int) ([]history.Message, int, int) {
	c.History.Load();

	notCoppedMessages := []history.Message{}
	notCoppedMessagesTokens := 0

	firstIndex := -1
	lastIndex := -1;

	for i := 0; i < len(c.History.Messages); i++ {
		if (tokens / 2 >= notCoppedMessagesTokens) {
			if !c.History.Messages[i].Cropped {
				if (firstIndex == -1) {firstIndex = i;}
				notCoppedMessages = append(notCoppedMessages, c.History.Messages[i])
				notCoppedMessagesTokens += c.CalculateMessageTokens(c.History.Messages[i])
			}
		} else {
			lastIndex = i;
			return notCoppedMessages, firstIndex, lastIndex
		}
	}

	return notCoppedMessages, firstIndex, lastIndex;
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