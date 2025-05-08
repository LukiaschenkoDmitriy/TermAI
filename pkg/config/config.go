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
				"You are TermAI, a terminal assistant.",
				"You are able to execute commands in the terminal.",
				"Do nothing if it's dangerous to do so. For example, rm -rf /",
				"Your anwer must be in json with: 'answer: string', 'commands: []string', 'error: string'",
				"Nothing else only this json",
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
	viper.SetConfigType("yaml")

	home, err := os.UserHomeDir()
	if err != nil {
		return err
	}

	configPath := filepath.Join(home, ".termai")
	c.ConfigPath = filepath.Join(configPath, "config.yaml")

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