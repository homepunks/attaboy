package bot

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/homepunks/attaboy/internal/config"
)

func handleTextMessage(upd Update, cfg config.Config) {
	chatID := upd.Message.Chat.ID
	// userText := upd.Message.Text
	text := "greetings from attaboy! i can help you be present when absent."

	if err := sendMessage(chatID, cfg, text); err != nil {
		log.Printf("")
	}
}

func sendMessage(chatID int64, cfg config.Config, text string) error { 
	url := fmt.Sprintf("%s%s/sendMessage", cfg.BaseURL, cfg.BotToken)

	msg := map[string]any{
		"chat_id": chatID,
		"text":    text,
	}

	jsonData, err := json.Marshal(msg)
	if err != nil {
		return err
	}

	resp, err := http.Post(url, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return err
	}

	defer resp.Body.Close()

	return nil
}
