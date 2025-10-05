package config

import (
	"fmt"
	"strings"

	"github.com/spf13/viper"
)

type Config struct {
	Server   ServerConfig   `mapstructure:"server"`
	Database DatabaseConfig `mapstructure:"database"`
	Telegram TelegramConfig `mapstructure:"telegram"`
	Log      LogConfig      `mapstructure:"log"`
	LLM      LLMConfig      `mapstructure:"llm"`
}

type ServerConfig struct {
	Port string `mapstructure:"port"`
}

type DatabaseConfig struct {
	Host     string `mapstructure:"host"`
	Port     string `mapstructure:"port"`
	User     string `mapstructure:"user"`
	Password string `mapstructure:"password"`
	DBName   string `mapstructure:"dbname"`
	SSLMode  string `mapstructure:"sslmode"`
}

type TelegramConfig struct {
	Token  string `mapstructure:"token"`
	Debug  bool   `mapstructure:"debug"`
	Admins []int  `mapstructure:"admins"`
}

type LogConfig struct {
	Level string `mapstructure:"level"`
	File  string `mapstructure:"file"`
}

// LLMConfig holds generic LLM provider settings. Supported providers:
// - For OpenRouter, set provider: "openrouter", api_key, model_name (e.g. "x-ai/grok-4-fast:free"),
//   and optionally base_url (defaults to https://openrouter.ai/api/v1).
// - For Artemox, set provider: "artemox", api_key, model_name (e.g. "gpt-4o-mini").
type LLMConfig struct {
	Provider  string `mapstructure:"provider"`
	APIKey    string `mapstructure:"api_key"`
	ModelName   string `mapstructure:"model_name"`
	MaxTokens int    `mapstructure:"max_tokens"`
	BaseURL   string `mapstructure:"base_url"`
}

func LoadConfig(configPath string) (*Config, error) {
	v := viper.New()
	
	v.SetConfigName("config")
	v.SetConfigType("yaml")
	v.AddConfigPath(configPath)
	v.AddConfigPath(".")
	v.AutomaticEnv()
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	
	if err := v.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("error reading config file: %w", err)
	}
	
	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("unable to decode config into struct: %w", err)
	}
	
	return &cfg, nil
}
