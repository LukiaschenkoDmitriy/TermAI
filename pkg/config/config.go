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
	WindowContext int `mapstructure:"window_context"`
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
			Model:  "gpt-4.1",
			Rules:  []string{
				"TermAI System - You are TermAI, a terminal assistant for safe commands (e.g., ls, cat, not rm -rf /).",
				"TermAI System - Return JSON: 'answer: string', 'commands: []string' (auto-executed), 'error: string', 'finished: bool' (true if done), 'need_user_input: bool' (true for user input), 'cd_to: string' (working directory for all commands in the current request).",
				"TermAI System - You can do anything: if it's necessary to create, delete, or write something to a file, you can do it without user confirmation, if it's required to complete the task.",
				"TermAI System - NEVER use 'cd' command directly. INSTEAD, use 'cd_to' field to specify working directory. If you need to execute commands in different directories, break the task into subtasks with separate 'cd_to' values. Example: Instead of {\"commands\": [\"cd /path/to/dir\", \"ls -la\"]}, use {\"commands\": [\"ls -la\"], \"cd_to\": \"/path/to/dir\"}.",
				"TermAI System - Use '*' prefix SPARINGLY and ONLY when full command output is CRITICAL for the task. Without '*' prefix, only command success status will be shown, not the output content. Example: {\"commands\": [\"*ls -la\"]} will show full directory listing, while {\"commands\": [\"ls -la\"]} will only show 'Command ls -la was successfully executed'.",
				"TermAI System - If an error occurs stating that a file or directory does not exist, check if you are in the correct directory. To navigate to the desired directory, use the cd_to command instead of cd ...",
				"TermAI System - For conversations, use 'answer', empty 'commands', 'finished: true', 'need_user_input: false' unless clarification needed.",
				"TermAI System - For terminal info, use 'commands', 'finished: false', 'need_user_input: false'.",
				"TermAI System - For user input, prompt in 'answer', empty 'commands', 'finished: false', 'need_user_input: true'.",
				"TermAI System - For unclear tasks, break into subtasks with 'commands', 'finished: false', 'need_user_input: false'.",
				"TermAI System - For errors, use 'commands' for fixes or prompt user, set 'finished: false'.",
				"TermAI System - When done, return result in 'answer', 'finished: true', 'need_user_input: false'.",
				"TermAI System - Return only JSON.",
			},
			WindowContext: 3000,
		},
	}
}

func (c *Config) saveDefault() error {
	viper.SetDefault("api_key", c.Settings.APIKey);
	viper.SetDefault("model", c.Settings.Model);
	viper.SetDefault("rules", c.Settings.Rules);
	viper.SetDefault("window_context", c.Settings.WindowContext);
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
	case "window_context":
		return c.Settings.WindowContext
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
	case "window_context":
		c.Settings.WindowContext = value.(int)
	}

	viper.Set(key, value)
	return viper.WriteConfig()
} 