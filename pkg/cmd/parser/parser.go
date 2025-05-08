package parser

import (
	"fmt"
	"log"

	"github.com/dmytrii/termai/pkg/config"
	"github.com/spf13/cobra"
)

type CMDParser struct {
	ApiKey string
	Model string
	RootCMD *cobra.Command
	ConfigCMD *cobra.Command
}

func New() *CMDParser {
	parser := &CMDParser{
		RootCMD: &cobra.Command{
			Use:   "termai",
			Short: "TermAI - terminal assistant",
		},
		ConfigCMD: &cobra.Command{
			Use:   "config",
			Short: "Configure TermAI settings",
		},
	};

	return parser;
}

func (parser *CMDParser) Init() {
	parser.RootCMD.AddCommand(parser.ConfigCMD);

	parser.ConfigCMD.Flags().StringVar(&parser.ApiKey, "api_key", "", "Set API key")
	parser.ConfigCMD.Flags().StringVar(&parser.Model, "model", "", "Set model")

	parser.ConfigCMD.Run = parser.ConfigRun
}

func (parser* CMDParser) ConfigRun(cmd *cobra.Command, args []string) {
	cfg := config.New()
	if err := cfg.Load(); err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	if parser.ApiKey != "" {
		if err := cfg.Set("api_key", parser.ApiKey); err != nil {
			log.Fatalf("Failed to set API key: %v", err)
		}
		fmt.Println("API key updated")
	}

	if parser.Model != "" {
		if err := cfg.Set("model", parser.Model); err != nil {
			log.Fatalf("Failed to set model: %v", err)
		}
		fmt.Println("Model updated")
	}

	if parser.Model == "" && parser.ApiKey == "" {
		fmt.Printf("\nCurrent settings:\n")
		fmt.Printf("API Key: %s\n", cfg.Settings.APIKey)
		fmt.Printf("Model: %s\n", cfg.Settings.Model)
	}
}

func (parser* CMDParser) Execute() error {
	return parser.RootCMD.Execute();
}