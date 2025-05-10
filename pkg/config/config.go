package config

import (
	"os"
	"path/filepath"

	"github.com/spf13/viper"
)

type ConfigSettings struct {
	APIKey string `mapstructure:"api_key"`
	Model  string `mapstructure:"model"`
	Rules  []string `mapstructure:"rules"`
}

type Config struct {
	ConfigPath string
	Settings ConfigSettings
}

func New() *Config {
	return &Config{
		ConfigPath: "",
		Settings: ConfigSettings{
			APIKey: "",
			Model:  "gpt-4o-mini",
			Rules:  []string{
				"TermAI System - You are TermAI, a terminal assistant.",
				"TermAI System - You are able to execute commands in the terminal.",
				"TermAI System - Do nothing if it's dangerous to do so. For example, rm -rf /",
				"TermAI System - All commands who are in the json commands will be executed in the terminal automatically.",
				"TermAI System - Your anwer must be in json with: 'answer: string', 'commands: []string', 'error: string'",
				"TermAI System - If the user wants to discuss various topics (like feelings, daily life, etc.), engage in a friendly conversation.",
				"TermAI System - If more information is needed to complete a task, ALWAYS use 'more_information' as the first command, followed MANDATORILY by specific commands to gather the required information, which must be automatically executed in the terminal. Example: {\"commands\": [\"more_information\", \"ls -la\"]}",
				"TermAI System - If a task is unclear or lacks sufficient information, ALWAYS break it down into subtasks, identify the necessary steps, and include specific commands after 'more_information' to gather required information and successfully complete the task. Example: For an unclear request like 'Analyze the project,' subtasks may include listing directory contents and checking configuration files, resulting in commands like {\"commands\": [\"more_information\", \"ls -la\", \"cat package.json\"]}",
				"TermAI System - Safe operations include: reading files (cat, less, more), listing directories (ls), checking file permissions, viewing file content, and other non-destructive operations.",
				"TermAI System - Nothing else only this json",
			},
		},
	}
}

func (c *Config) saveDefault() error {
	viper.SetDefault("api_key", c.Settings.APIKey);
	viper.SetDefault("model", c.Settings.Model);
	viper.SetDefault("rules", c.Settings.Rules);

	return viper.WriteConfigAs(c.ConfigPath)
}

func (c *Config) Load() error {
	viper.SetConfigName("config")
	viper.SetConfigType("json")

	home, err := os.UserHomeDir()
	if err != nil {
		return err
	}

	configPath := filepath.Join(home, ".termai")
	c.ConfigPath = filepath.Join(configPath, "config.json")

	if err := os.MkdirAll(configPath, 0755); err != nil {
		return err
	}

	viper.AddConfigPath(configPath)

	if _, err := os.Stat(c.ConfigPath); os.IsNotExist(err) {
		if err := c.saveDefault(); err != nil {
			return err
		}
	}

	if err := viper.ReadInConfig(); err != nil {
		return err
	}

	return viper.Unmarshal(&c.Settings)
}

func (c *Config) Get(key string) any {
	switch key {
	case "api_key":
		return c.Settings.APIKey
	case "model":
		return c.Settings.Model
	case "rules":
		return c.Settings.Rules
	default:
		return nil
	}
}

func (c *Config) Set(key string, value any) error {
	switch key {
	case "api_key":
		c.Settings.APIKey = value.(string)
	case "model":
		c.Settings.Model = value.(string)
	case "rules":
		c.Settings.Rules = value.([]string)
	}

	viper.Set(key, value)
	return viper.WriteConfig()
} 