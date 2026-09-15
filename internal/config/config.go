package config

import (
	"errors"
	"os"
)

type Config struct {
	BotToken       string
	BaseURL        string
	MoodleURL      string
	MoodleUsername string
	MoodlePassword string
}

func LoadConfig() (*Config, error) {
	token := os.Getenv("TG_BOT_TOKEN")
	if token == "" {
		return nil, errors.New("TG_BOT_TOKEN not found in your env")
	}

	username := os.Getenv("MOODLE_USERNAME")
	if username == "" {
		return nil, errors.New("MOODLE_USERNAME not found in your env")
	}

	password := os.Getenv("MOODLE_PASSWORD")
	if password == "" {
		return nil, errors.New("MOODLE_PASSWORD not found in your env")
	}

	return &Config{
		BotToken:       token,
		BaseURL:        "https://api.telegram.org/bot",
		MoodleURL:      "https://moodle.nu.edu.kz",
		MoodleUsername: username,
		MoodlePassword: password,
	}, nil
}
