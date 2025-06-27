package openai

import (
	"bufio"
	"encoding/json"
	"fmt"
	"log"
	"os/exec"
	"strings"

	"github.com/LukiaschenkoDmitriy/TermAI/pkg/config"
	"github.com/LukiaschenkoDmitriy/TermAI/pkg/history"
	"github.com/LukiaschenkoDmitriy/TermAI/pkg/http/client"
	"github.com/LukiaschenkoDmitriy/TermAI/pkg/http/response"
	"github.com/LukiaschenkoDmitriy/TermAI/pkg/openai/context"
)

type OpenAI struct {
	Response *response.Response
	Client   *client.Client
	Context *context.Context
	Config *config.Config
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
	config := config.New();
	config.Load();

	return &OpenAI{
		Client: client.New(),
		Context: context.New(),
		Config: config,
	}
}

func (o *OpenAI) SendMessage(message string) (*response.Response, error) {
	return o.Client.SendRequest([]string{message})
}

func (o *OpenAI) ProcessResponse(response *response.Response) (OpenAIResponse, error) {
	if response == nil || len(response.Choices) == 0 {
		return OpenAIResponse{}, fmt.Errorf("invalid response: response or choices is nil")
	}

	content := response.Choices[0].Message.Content
	content = strings.TrimPrefix(content, "```json\n")
	content = strings.TrimSuffix(content, "\n```")

	var parsedContent OpenAIResponse

	if err := json.Unmarshal([]byte(content), &parsedContent); err != nil {
		log.Printf("Failed to parse content. Error: %v\n", err)
		log.Printf("Content that caused error: %s\n", content)

		return OpenAIResponse{}, err;
	}

	return parsedContent, nil
}

func (o *OpenAI) ReplaceCommandsFormat(commands []string) []string {
	for i, cmd := range commands {
		cmd = strings.ReplaceAll(cmd, "\\$", "$")
		
		commands[i] = cmd
	}

	return commands
}

func (o *OpenAI) ExecuteCommands(commands []string, dir string, answer string) (string, error) {
	var allOutput string
	var errorCmd error

	commands = o.ReplaceCommandsFormat(commands)

	if len(commands) > 0 {
		for _, cmd := range commands {
			if strings.HasPrefix(cmd, "cd") {
				allOutput += "System: Command 'cd ...' is not available, use 'cd_to' field to specify working directory INSTEAD 'cd ...'.\n"
				return allOutput, fmt.Errorf("command not available")
			}

			showFullOutput := false
			if strings.HasPrefix(cmd, "*") {
				showFullOutput = true
				cmd = strings.TrimPrefix(cmd, "*")
			}

			cmd = strings.TrimSpace(cmd)
			if strings.Count(cmd, "'")%2 != 0 {
				cmd = strings.Replace(cmd, "'", "\"", -1)
			}

			fmt.Printf("Executing: %s\n", cmd)

			execCmd := exec.Command("bash", "-c", cmd)
			if dir != "" {
				execCmd.Dir = dir
			}

			// Create pipes for stdout and stderr
			stdout, err := execCmd.StdoutPipe()
			if err != nil {
				return allOutput, fmt.Errorf("failed to create stdout pipe: %v", err)
			}
			stderr, err := execCmd.StderrPipe()
			if err != nil {
				return allOutput, fmt.Errorf("failed to create stderr pipe: %v", err)
			}

			// Start the command
			if err := execCmd.Start(); err != nil {
				return allOutput, fmt.Errorf("failed to start command: %v", err)
			}

			// Create a scanner for stdout
			stdoutScanner := bufio.NewScanner(stdout)
			go func() {
				for stdoutScanner.Scan() {
					line := stdoutScanner.Text()
					fmt.Printf("> %s\n", line)
					allOutput += line + "\n"
				}
			}()

			// Create a scanner for stderr
			stderrScanner := bufio.NewScanner(stderr)
			go func() {
				for stderrScanner.Scan() {
					line := stderrScanner.Text()
					fmt.Printf("! %s\n", line)
					allOutput += "Error: " + line + "\n"
				}
			}()

			// Wait for the command to complete
			if err := execCmd.Wait(); err != nil {
				exitErr, ok := err.(*exec.ExitError)
				if ok {
					errorCmd = fmt.Errorf("Error (exit code %d)", exitErr.ExitCode())
					fmt.Printf("Error (exit code %d)\n", exitErr.ExitCode())
				} else {
					errorCmd = fmt.Errorf("Error: %v", err)
					fmt.Printf("Error: %v\n", err)
				}
				break
			}

			if !showFullOutput {
				fmt.Printf("Command %s was successfully executed\n", cmd)
			}
		}
	}

	// Trim output to last 15000 characters if too long
	if len(allOutput) > 15000 {
		allOutput = allOutput[len(allOutput)-15000:]
	}

	return allOutput, errorCmd
}

func (o *OpenAI) AddToHistory(aiResponse *response.Response, commandOutput string) {
	lastMessage := o.Client.Messages[len(o.Client.Messages)-1]
	o.Client.History.AddMessage(history.ClientMessage{
		Role: lastMessage.Role,
		Content: lastMessage.Content,
	}, aiResponse.Choices[0].Message, commandOutput, false)
}

func (o *OpenAI) AddRulesToHistoryIfNotExists(rules []string) {
	if !o.Client.History.IsSystemRulesExists() {
		o.Client.History.AddMessage(history.ClientMessage{
			Role: "system",
			Content: fmt.Sprintf("system: %s", strings.Join(rules, "\n")),
		}, response.Message{
			Role: "system",
			Content: "",
		}, "", true)
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

func (o *OpenAI) CropIfWindowContextIsFull() {
	tokens := o.Context.CalculateContextTokens();
	o.Client.History.Load();

	o.Context.History.Load();

	fmt.Printf("TMP: %d\n", tokens);

	if (tokens > o.Config.Settings.WindowContext) {
		messages, firstIndex, lastIndex := o.Context.GetNotCoppedMessageByTokens(o.Config.Settings.WindowContext);
		lastMessage := o.Context.ConvertMessagesToOneMessage(messages);

		if (firstIndex == -1 && lastIndex == -1) {
			return;
		}

		o.Client.History.DeleteByIndexRange(firstIndex + 1, lastIndex);

		responseContent, _ := o.Client.SendRequestWithoutContext([]string{strings.Join(o.Config.Settings.Rules, "\n") + fmt.Sprintf("TermAI System: Compress the entire provided text without losing any information:\n%s\n%s\n%s", lastMessage.UserMessage.Content, lastMessage.AIMessage.Content, lastMessage.CommandOutput)});
		transformedResponse, _ := o.ProcessResponse(responseContent);

		o.Client.History.Messages[firstIndex].UserMessage.Content = "TermAISystem Context: " + transformedResponse.Answer;
		o.Client.History.Messages[firstIndex].AIMessage.Content = "";
		o.Client.History.Messages[firstIndex].CommandOutput = "";
		o.Client.History.Messages[firstIndex].Cropped = true;

		o.Client.History.Save();
	}
}