package parser

import (
	"fmt"
	"log"

	"github.com/LukiaschenkoDmitriy/TermAI/pkg/config"
	"github.com/LukiaschenkoDmitriy/TermAI/pkg/history"
	"github.com/LukiaschenkoDmitriy/TermAI/pkg/openai"
	"github.com/spf13/cobra"
)

type CMDParser struct {
	Config *config.Config
	LastMessage string
	Clear bool
	RootCMD *cobra.Command
	ConfigCMD *cobra.Command
	ExecuteCMD *cobra.Command
	HistoryCMD *cobra.Command
	openai *openai.OpenAI
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
		HistoryCMD: &cobra.Command{
			Use:   "history",
			Short: "Show history",
		},
	}
	parser.openai = openai.New()

	return parser
}

func (parser *CMDParser) Init() {
	parser.RootCMD.AddCommand(parser.ConfigCMD);
	parser.RootCMD.AddCommand(parser.ExecuteCMD);
	parser.RootCMD.AddCommand(parser.HistoryCMD);

	parser.ConfigCMD.Flags().StringVar(&parser.Config.Settings.APIKey, "api_key", "", "Set API key")
	parser.ConfigCMD.Flags().StringVar(&parser.Config.Settings.Model, "model", "", "Set model")

	parser.ExecuteCMD.Flags().StringVar(&parser.LastMessage, "message", "", "Message")
	parser.HistoryCMD.Flags().BoolVar(&parser.Clear, "clear", false, "Clear history")

	parser.ConfigCMD.Run = parser.ConfigRun
	parser.ExecuteCMD.Run = parser.ExecuteRun
	parser.HistoryCMD.Run = parser.HistoryRun
}

func (parser *CMDParser) ConfigRun(cmd *cobra.Command, args []string) {
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

func (parser *CMDParser) HistoryRun(cmd *cobra.Command, args []string) {
	history :=history.New()
	history.Load()
	if (parser.Clear) {
		history.ClearHistory()
	} else {
		if (len(history.Messages) > 0) {
			for _, message := range history.Messages {
				fmt.Printf("[%s] %s\n", message.UserMessage.Role, message.UserMessage.Content)
				fmt.Printf("[%s] %s\n", message.AIMessage.Role, message.AIMessage.Content)
				fmt.Printf("[%s] %s\n", message.CommandOutput, message.CommandOutput)
			}
		} else {
			fmt.Println("[TermAI] : History is empty")
		}
	}
}

func (parser *CMDParser) ExecuteRun(cmd *cobra.Command, args []string) {
	parser.ExecuteLogic()
}

func (parser *CMDParser) ExecuteLogic() {
	parser.openai.AddRulesToHistoryIfNotExists(parser.Config.Settings.Rules)

	response, err := parser.openai.SendMessage(parser.LastMessage)
	if err != nil {
		log.Fatalf("Failed to send request: %v", err)
	}

	openaiResponse, err := parser.openai.ProcessResponse(response)
	if err != nil {
		log.Fatalf("Failed to process response: %v", err)
	}

	if openaiResponse.Error != "" {
		log.Printf("Warning: %s", openaiResponse.Error)
	}

	fmt.Println("[TermAI]: " + openaiResponse.Answer)

	allOutput, err := parser.openai.ExecuteCommands(openaiResponse.Commands, openaiResponse.CDTo)

	parser.openai.AddToHistory(response, allOutput)

	parser.openai.CropIfWindowContextIsFull();

	if (openaiResponse.NeedUserInput) {
		return;
	}

	if err != nil {
		parser.LastMessage = "System: " + allOutput
		parser.ExecuteLogic()
		return
	}

	if (!openaiResponse.Finished) {
		parser.LastMessage = "Output: " + allOutput
		parser.ExecuteLogic()
		return
	}
}

func (parser *CMDParser) Execute() error {
	return parser.RootCMD.Execute();
}