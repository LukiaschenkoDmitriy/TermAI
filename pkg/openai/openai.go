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

type OpenAIResponse struct {
	Answer string `json:"answer"`
	Commands []string `json:"commands"`
	Finished bool `json:"finished"`
	Error string `json:"error"`
	NeedUserInput bool `json:"need_user_input"`
	CDTo string `json:"cd_to"`
}

func New() *OpenAI {
	return &OpenAI{
		Client: client.New(),
	}
}

func (o *OpenAI) SendMessage(message string) (*response.Response, error) {
	return o.Client.SendRequest([]string{message})
}

func (o *OpenAI) ProcessResponse(response *response.Response) (OpenAIResponse, error) {
	content := response.Choices[0].Message.Content
	content = strings.TrimPrefix(content, "```json\n")
	content = strings.TrimSuffix(content, "\n```")

	var parsedContent OpenAIResponse

	if err := json.Unmarshal([]byte(content), &parsedContent); err != nil {
		log.Printf("Failed to parse content. Error: %v\n", err)
		log.Printf("Content that caused error: %s\n", content)
		
		content = strings.ReplaceAll(content, "\n", " ")
		content = strings.ReplaceAll(content, "\r", "")
		content = strings.TrimSpace(content)

		content = strings.ReplaceAll(content, "}\",\"answer", ",\"answer")
		content = strings.ReplaceAll(content, "}\",\"error", ",\"error")
		content = strings.ReplaceAll(content, "}\",\"commands", ",\"commands")
		content = strings.ReplaceAll(content, "}\",\"cd_to", ",\"cd_to")
		
		content = strings.ReplaceAll(content, "\"[", "[")
		content = strings.ReplaceAll(content, "]\"", "]")
		
		content = strings.ReplaceAll(content, ",,", ",")
		content = strings.ReplaceAll(content, ",\"}", "}")
		content = strings.ReplaceAll(content, ",\"}", "}")
		
		if err := json.Unmarshal([]byte(content), &parsedContent); err != nil {
			commandsStart := strings.Index(content, "\"commands\":[")
			commandsEnd := strings.LastIndex(content, "]")
			
			if commandsStart != -1 && commandsEnd != -1 {
				commandsStr := content[commandsStart+11:commandsEnd+1]
				
				commandsStr = strings.ReplaceAll(commandsStr, "\"", "")
				
				var commands []string
				var currentCmd strings.Builder
				inQuotes := false
				
				for i := 0; i < len(commandsStr); i++ {
					char := commandsStr[i]
					
					if char == '\'' {
						inQuotes = !inQuotes
						currentCmd.WriteByte(char)
					} else if char == ',' && !inQuotes {
						commands = append(commands, strings.TrimSpace(currentCmd.String()))
						currentCmd.Reset()
					} else {
						currentCmd.WriteByte(char)
					}
				}
				
				if currentCmd.Len() > 0 {
					commands = append(commands, strings.TrimSpace(currentCmd.String()))
				}
				
				var formattedCommands []string
				for _, cmd := range commands {
					escapedCmd := strings.ReplaceAll(cmd, "\\", "\\\\")
					escapedCmd = strings.ReplaceAll(escapedCmd, "\"", "\\\"")
					formattedCommands = append(formattedCommands, fmt.Sprintf("\"%s\"", escapedCmd))
				}
				
				cdTo := ""
				if cdToStart := strings.Index(content, "\"cd_to\":"); cdToStart != -1 {
					cdToValueStart := cdToStart + 7
					cdToValueEnd := strings.Index(content[cdToValueStart:], "\"")
					if cdToValueEnd != -1 {
						cdToValueEnd = cdToValueStart + cdToValueEnd
						cdTo = content[cdToValueStart:cdToValueEnd]
					}
				}
				
				newContent := fmt.Sprintf(`{"answer":"","commands":[%s],"error":"","cd_to":"%s"}`, strings.Join(formattedCommands, ","), cdTo)
				log.Printf("Created new JSON: %s\n", newContent)
				
				if err := json.Unmarshal([]byte(newContent), &parsedContent); err != nil {
					return OpenAIResponse{}, fmt.Errorf("failed to parse response after reconstruction: %v", err)
				}
			} else {
				return OpenAIResponse{}, fmt.Errorf("failed to find commands in response")
			}
		}
	}

	return parsedContent, nil
}

func (o *OpenAI) ReplaceCommandsFormat(commands []string) []string {
	for i, cmd := range commands {
		commands[i] = strings.Replace(cmd, "echo '", "echo \"", -1)
		commands[i] = strings.Replace(cmd, "' >", "\" >", -1)
		commands[i] = strings.Replace(cmd, "' >>", "\" >>", -1)
	}

	return commands
}

func (o *OpenAI) ExecuteCommands(commands []string, dir string) (string, error) {
	var allOutput string
	var errorCmd error

	commands = o.ReplaceCommandsFormat(commands)

	if len(commands) > 0 {
		for _, cmd := range commands {
			if strings.HasPrefix(cmd, "cd") {
				allOutput += "System: Command 'cd ...' is not available, use 'cd_to' field to specify working directory.\n"
				return allOutput, fmt.Errorf("command not available")
			}

			showFullOutput := false
			if strings.HasPrefix(cmd, "*") {
				showFullOutput = true
				cmd = strings.TrimPrefix(cmd, "*")
			}

			execCmd := exec.Command("bash", "-c", cmd)

			if dir != "" {
				execCmd.Dir = dir
			}

			output, err := execCmd.CombinedOutput()
			
			if err != nil {
				exitErr, ok := err.(*exec.ExitError)
				errorCmd = fmt.Errorf("Error (exit code %d):\n%s\n", exitErr.ExitCode(), string(output))
				if ok {
					allOutput += fmt.Sprintf("Error (exit code %d):\n%s\n", exitErr.ExitCode(), string(output))
					fmt.Printf("Error (exit code %d):\n%s\n", exitErr.ExitCode(), string(output))
				} else {
					allOutput += fmt.Sprintf("Error: %v\nOutput: %s\n", err, string(output))
					fmt.Printf("Error: %v\nOutput: %s\n", err, string(output))
				}
				break
			}

			if len(output) > 0 {
				if showFullOutput {
					allOutput += fmt.Sprintf("Command output:\n%s\n", output)
				} else {
					allOutput += fmt.Sprintf("Command %s was successfully executed\n", cmd)
				}
				fmt.Printf("Output:\n%s\n", output)
			}
		}
	}

	return allOutput, errorCmd
}

func (o *OpenAI) AddToHistory(aiResponse *response.Response, commandOutput string) {
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