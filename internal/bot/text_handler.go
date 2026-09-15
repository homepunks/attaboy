package bot

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"

	"github.com/homepunks/attaboy/internal/config"
)

func handleTextMessage(upd Update, cfg config.Config) {
	// userText := upd.Message.Text
	reply(upd, cfg, "greetings from attaboy! i can help you be present when absent.")
}

func reply(upd Update, cfg config.Config, text string) {
	if err := sendMessage(upd.Message.Chat.ID, cfg, text); err != nil {
		log.Printf("Could not send message to %s (@%s): %v",
			upd.Message.Chat.Name, upd.Message.Chat.Username, err)
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

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("sendMessage failed: %s: %s", resp.Status, body)
	}

	return nil
}
