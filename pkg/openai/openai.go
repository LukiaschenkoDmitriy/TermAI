package openai

import (
	"encoding/json"
	"fmt"
	"log"
	"os/exec"
	"strings"

	"github.com/LukiaschenkoDmitriy/TermAI/pkg/history"
	"github.com/LukiaschenkoDmitriy/TermAI/pkg/http/client"
	"github.com/LukiaschenkoDmitriy/TermAI/pkg/http/response"
)

type OpenAI struct {
	Response *response.Response
	Client   *client.Client
}

func New() *OpenAI {
	return &OpenAI{
		Client: client.New(),
	}
}

func (o *OpenAI) SendMessage(message string) (*response.Response, error) {
	return o.Client.SendRequest([]string{message})
}

func (o *OpenAI) ProcessResponse(response *response.Response) (string, []string, string, error) {
	content := response.Choices[0].Message.Content
	content = strings.TrimPrefix(content, "```json\n")
	content = strings.TrimSuffix(content, "\n```")

	var parsedContent struct {
		Answer   string   `json:"answer"`
		Commands []string `json:"commands"`
		Error    string   `json:"error"`
	}

	if err := json.Unmarshal([]byte(content), &parsedContent); err != nil {
		log.Printf("Failed to parse content. Error: %v\n", err)
		log.Printf("Content that caused error: %s\n", content)
		
		content = strings.ReplaceAll(content, "\n", " ")
		content = strings.ReplaceAll(content, "\r", "")
		content = strings.TrimSpace(content)
		
		log.Printf("Cleaned content for retry: %s\n", content)
		
		if err := json.Unmarshal([]byte(content), &parsedContent); err != nil {
			return "", nil, "", fmt.Errorf("failed to parse response after cleaning: %v", err)
		}
	}

	return parsedContent.Answer, parsedContent.Commands, parsedContent.Error, nil
}

func (o *OpenAI) ReadFiles(commands []string) string {
	var allOutput string

	for _, cmd := range commands {
		cmd = strings.Replace(cmd, "echo '", "echo \"", -1)
		cmd = strings.Replace(cmd, "' >", "\" >", -1)
		cmd = strings.Replace(cmd, "' >>", "\" >>", -1)

		if commands[0] != "more_information" {
			fmt.Printf("\nExecuting: %s\n", cmd)
		}

		execCmd := exec.Command("bash", "-c", cmd)
		output, err := execCmd.CombinedOutput()
		
		if err != nil {
			exitErr, ok := err.(*exec.ExitError)
			if ok {
				allOutput += fmt.Sprintf("Error (exit code %d):\n%s\n", exitErr.ExitCode(), string(output))
				fmt.Printf("Error (exit code %d):\n%s\n", exitErr.ExitCode(), string(output))
			} else {
				allOutput += fmt.Sprintf("Error: %v\nOutput: %s\n", err, string(output))
				fmt.Printf("Error: %v\nOutput: %s\n", err, string(output))
			}
			continue	
		}

		if len(output) > 0 {
			if commands[0] != "more_information" {
				fmt.Printf("Output:\n%s\n", output)
			}
			allOutput += fmt.Sprintf("Output:\n%s\n", output)
		}
	}

	return allOutput
}

func (o *OpenAI) ExecuteCommands(commands []string, log bool) string {
	var allOutput string

	if len(commands) > 0 {
		for _, cmd := range commands {
			cmd = strings.Replace(cmd, "echo '", "echo \"", -1)
			cmd = strings.Replace(cmd, "' >", "\" >", -1)
			cmd = strings.Replace(cmd, "' >>", "\" >>", -1)

			if log {
				fmt.Printf("\nExecuting: %s\n", cmd)
			}

			allOutput += fmt.Sprintf("Executing: %s\n", cmd)
			
			execCmd := exec.Command("bash", "-c", cmd)
			output, err := execCmd.CombinedOutput()
			
			if err != nil {
				exitErr, ok := err.(*exec.ExitError)
				if ok {
					allOutput += fmt.Sprintf("Error (exit code %d):\n%s\n", exitErr.ExitCode(), string(output))
					fmt.Printf("Error (exit code %d):\n%s\n", exitErr.ExitCode(), string(output))
				} else {
					allOutput += fmt.Sprintf("Error: %v\nOutput: %s\n", err, string(output))
					fmt.Printf("Error: %v\nOutput: %s\n", err, string(output))
				}
				continue
			}

			if len(output) > 0 {
				if log {
					fmt.Printf("Output:\n%s\n", output)
				}
				allOutput += fmt.Sprintf("Output:\n%s\n", output)
			}
		}
	}

	return allOutput
}

func (o *OpenAI) AddToHistory(userMessage string, aiResponse *response.Response, commandOutput string) {
	lastMessage := o.Client.Messages[len(o.Client.Messages)-1]
	o.Client.History.AddMessage(history.ClientMessage{
		Role: lastMessage.Role,
		Content: lastMessage.Content,
	}, aiResponse.Choices[0].Message, commandOutput)
}

func (o *OpenAI) AddRulesToHistoryIfNotExists(rules []string) {
	if !o.Client.History.IsSystemRulesExists() {
		o.Client.History.AddMessage(history.ClientMessage{
			Role: "system",
			Content: fmt.Sprintf("system: %s", strings.Join(rules, "\n")),
		}, response.Message{
			Role: "system",
			Content: "",
		}, "")
	}
}

func (o *OpenAI) HandleMoreInformation(commands []string, previousContext string) (*response.Response, error) {
	message := fmt.Sprintf("System: More Information Request\nContext: %s\nCommands to execute:\n%s", 
		previousContext,
		strings.Join(commands, "\n"))
	
	response, err := o.SendMessage(message)
	if err != nil {
		return nil, fmt.Errorf("failed to request more information: %v", err)
	}

	return response, nil
} 