package config

import (
	"errors"
	"os"
)

type Config struct {
	BotToken string
	BaseURL string
}

func LoadConfig() (*Config, error) {
	token := os.Getenv("TG_BOT_TOKEN")
	if token == "" {
		return nil, errors.New("TG_BOT_TOKEN not found in your env")
	}

	return &Config{
		BotToken: token,
		BaseURL: "https://api.telegram.org/bot",
	}, nil
}
