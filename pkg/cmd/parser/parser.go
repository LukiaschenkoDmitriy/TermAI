package parser

import (
	"fmt"
	"log"

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
	};

	return parser;
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
	client := client.New();

	response, err := client.SendRequest([]string{parser.LastMessage})
	if err != nil {
		log.Fatalf("Failed to send request: %v", err)
	}
	fmt.Println(string(response))
}

func (parser* CMDParser) Execute() error {
	return parser.RootCMD.Execute();
}