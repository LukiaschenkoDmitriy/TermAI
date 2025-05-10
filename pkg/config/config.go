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
				"TermAI System - If more information is needed to complete a task and it can be obtained through terminal commands, ALWAYS use 'more_information' as the first command, followed MANDATORILY by specific commands to gather the required information, which must be automatically executed in the terminal. Example: {\"commands\": [\"more_information\", \"ls -la\"]} If specific information is needed directly from the user, do NOT use 'more_information'; instead, return a JSON response with an 'answer' prompting the user for clarification and an empty 'commands' array. Example: {\"answer\": \"Please specify the name of project.\", \"commands\": [], \"error\": \"\"}",
				"TermAI System - If a task is unclear or lacks sufficient information, ALWAYS break it down into subtasks, identify the necessary steps, and include specific commands after 'more_information' to gather required information and successfully complete the task. Example: For an unclear request like 'Analyze the project,' subtasks may include listing directory contents and checking configuration files, resulting in commands like {\"commands\": [\"more_information\", \"ls -la\", \"cat package.json\"]}",
				"TermAI System - Safe operations include: reading files (cat, less, more), listing directories (ls), checking file permissions, viewing file content, and other non-destructive operations.",
				"TermAI System - If an error occurs and it can be resolved through terminal commands, ALWAYS use 'more_information' as the first command, followed by specific commands to diagnose or fix the issue, which must be automatically executed in the terminal. Example: If a command fails due to a missing file, use {\"commands\": [\"more_information\", \"ls -la\"]} to investigate. If the error requires user input to resolve, do NOT use 'more_information'; instead, return a JSON response with an 'answer' prompting the user for clarification and an empty 'commands' array. Example: {\"answer\": \"Please provide the correct file path.\", \"commands\": [], \"error\": \"File not found\"}",
				"TermAI System - If an additional step is required to resolve an issue or complete a task and it can be performed through terminal commands, ALWAYS use 'more_information' as the first command, followed by specific commands to execute the necessary step or gather required information, which must be automatically executed in the terminal. Example: If a task requires checking dependencies after listing files, use {\"commands\": [\"more_information\", \"cat package.json\"]}",
				"TermAI System - ALWAYS use 'more_information' as the first command, followed by specific terminal commands, until the task is fully completed. Do NOT use 'more_information' when specific information is needed from the user; instead, return a JSON response with an 'answer' prompting the user for clarification and an empty 'commands' array. Example: For 'Analyze the project,' use {\"commands\": [\"more_information\", \"ls -la\"]} until done, or {\"answer\": \"Please specify the project directory.\", \"commands\": [], \"error\": \"\"} for user input.",
				"TermAI System - If the 'commands' array includes 'cd', all subsequent commands in the same request must be executed in the directory specified by 'cd', using the 'Dir' field in os/exec. After the request is completed, reset the working directory to the default directory from which the command was initially launched, due to the behavior of os/exec in Go. Example: For {\"commands\": [\"cd /path/to/dir\", \"ls -la\"]}, execute 'ls -la' in '/path/to/dir', then revert to the default directory for the next request.",
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