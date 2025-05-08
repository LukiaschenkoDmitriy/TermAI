package parser

import (
	"encoding/json"
	"fmt"
	"log"
	"os/exec"
	"strings"

	"github.com/LukiaschenkoDmitriy/TermAI/pkg/config"
	"github.com/LukiaschenkoDmitriy/TermAI/pkg/http/client"
	"github.com/spf13/cobra"
)

type CMDParser struct {
	Config *config.Config
	LastMessage string
	RootCMD *cobra.Command
	ConfigCMD *cobra.Command
	ExecuteCMD *cobra.Command
	client *client.Client
}

func New() *CMDParser {
	cnf := config.New()
	if err := cnf.Load(); err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	parser := &CMDParser{
		Config: cnf,
		LastMessage: "",
		RootCMD: &cobra.Command{
			Use:   "termai",
			Short: "TermAI - terminal assistant",
		},
		ConfigCMD: &cobra.Command{
			Use:   "config",
			Short: "Configure TermAI settings",
		},
		ExecuteCMD: &cobra.Command{
			Use:   "execute",
			Short: "Execute TermAI",
		},
	}

	parser.client = client.New()

	return parser
}

func (parser *CMDParser) Init() {
	parser.RootCMD.AddCommand(parser.ConfigCMD);
	parser.RootCMD.AddCommand(parser.ExecuteCMD);

	parser.ConfigCMD.Flags().StringVar(&parser.Config.Settings.APIKey, "api_key", "", "Set API key")
	parser.ConfigCMD.Flags().StringVar(&parser.Config.Settings.Model, "model", "", "Set model")

	parser.ExecuteCMD.Flags().StringVar(&parser.LastMessage, "message", "", "Message")

	parser.ConfigCMD.Run = parser.ConfigRun
	parser.ExecuteCMD.Run = parser.ExecuteRun
}

func (parser* CMDParser) ConfigRun(cmd *cobra.Command, args []string) {
	cfg := config.New()
	if err := cfg.Load(); err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	if parser.Config.Settings.APIKey != "" {
		if err := cfg.Set("api_key", parser.Config.Settings.APIKey); err != nil {
			log.Fatalf("Failed to set API key: %v", err)
		}
		fmt.Println("API key updated")
	}

	if parser.Config.Settings.Model != "" {
		if err := cfg.Set("model", parser.Config.Settings.Model); err != nil {
			log.Fatalf("Failed to set model: %v", err)
		}
		fmt.Println("Model updated")
	}

	if parser.Config.Settings.Model == "" && parser.Config.Settings.APIKey == "" {
		fmt.Printf("\nCurrent settings:\n")
		fmt.Printf("API Key: %s\n", cfg.Settings.APIKey)
		fmt.Printf("Model: %s\n", cfg.Settings.Model)
	}
}

func (parser* CMDParser) ExecuteRun(cmd *cobra.Command, args []string) {
	parser.AddRulesToMessage(parser.Config.Settings.Rules)

	response, err := parser.client.SendRequest([]string{parser.LastMessage})
	if err != nil {
		log.Fatalf("Failed to send request: %v", err)
	}

	// Debug: Print raw content with escaped characters
	fmt.Printf("Raw content (with escapes): %q\n", response.Choices[0].Message.Content)

	// Remove markdown code block markers
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
		log.Fatal("Exiting due to parsing error")
	}

	fmt.Println(parsedContent.Answer)

	if len(parsedContent.Commands) > 0 {
		fmt.Println("\nExecuting commands:")
		for _, cmd := range parsedContent.Commands {
			cmd = strings.Replace(cmd, "echo '", "echo \"", -1)
			cmd = strings.Replace(cmd, "' >", "\" >", -1)
			cmd = strings.Replace(cmd, "' >>", "\" >>", -1)

			fmt.Printf("\nExecuting: %s\n", cmd)
			execCmd := exec.Command("bash", "-c", cmd)
			output, err := execCmd.CombinedOutput()
			if err != nil {
				fmt.Printf("Error executing command: %v\n", err)
				continue
			}
			fmt.Printf("Output:\n%s\n", output)
		}
	}
}

func (parser* CMDParser) AddRulesToMessage(rules []string) {
	for _, rule := range rules {
		parser.LastMessage += fmt.Sprintf("Rule: %s\n", rule)
	}
}

func (parser* CMDParser) Execute() error {
	return parser.RootCMD.Execute();
}