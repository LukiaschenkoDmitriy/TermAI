package parser_test

import (
	"testing"

	"github.com/LukiaschenkoDmitriy/TermAI/pkg/cmd/parser"
	"github.com/spf13/cobra"
)

func TestCMDParser_New(t *testing.T) {
	p := parser.New()
	if p == nil {
		t.Error("Expected non-nil parser")
	}

	if p.RootCMD == nil {
		t.Error("Expected non-nil RootCMD")
	}

	if p.ConfigCMD == nil {
		t.Error("Expected non-nil ConfigCMD")
	}

	if p.ExecuteCMD == nil {
		t.Error("Expected non-nil ExecuteCMD")
	}
}

func TestCMDParser_AddRulesToMessage(t *testing.T) {
	p := parser.New()
	rules := []string{"rule1", "rule2"}

	p.AddRulesToMessage(rules)

	expected := "Rule: rule1\nRule: rule2\n"
	if p.LastMessage != expected {
		t.Errorf("Expected message '%s', got '%s'", expected, p.LastMessage)
	}
}

func TestCMDParser_Init(t *testing.T) {
	p := parser.New()
	p.Init()

	// Check if commands are properly added
	rootCmd := p.RootCMD
	if len(rootCmd.Commands()) != 2 {
		t.Errorf("Expected 2 commands, got %d", len(rootCmd.Commands()))
	}

	// Check if flags are properly set
	executeCmd := p.ExecuteCMD
	if executeCmd.Flags().Lookup("message") == nil {
		t.Error("Expected 'message' flag to be set")
	}

	configCmd := p.ConfigCMD
	if configCmd.Flags().Lookup("api_key") == nil {
		t.Error("Expected 'api_key' flag to be set")
	}
	if configCmd.Flags().Lookup("model") == nil {
		t.Error("Expected 'model' flag to be set")
	}
}

func TestCMDParser_ConfigRun(t *testing.T) {
	p := parser.New()
	p.Init()

	// Test setting API key
	p.Config.Settings.APIKey = "test-key"
	p.ConfigRun(&cobra.Command{}, []string{})

	// Test setting model
	p.Config.Settings.Model = "test-model"
	p.ConfigRun(&cobra.Command{}, []string{})

	// Test displaying current settings
	p.Config.Settings.APIKey = ""
	p.Config.Settings.Model = ""
	p.ConfigRun(&cobra.Command{}, []string{})
} 