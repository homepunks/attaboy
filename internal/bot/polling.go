package bot

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"

	"github.com/homepunks/attaboy/internal/config"
)

func PollUpdates(offset int64, cfg config.Config) {
	for {
		url := fmt.Sprintf("%s%s/getUpdates?offset=%d&timeout=30",
			cfg.BaseURL, cfg.BotToken, offset)

		resp, err := http.Get(url)
		if err != nil {
			log.Printf("Error getting updates: %v\n", err)
			time.Sleep(5 * time.Second)
			continue
		}

		body, err := io.ReadAll(resp.Body)
		resp.Body.Close()
		if err != nil {
			log.Printf("Error reading body: %v\n", err)
			continue
		}

		var apiResponse struct {
			OK     bool     `json:"ok"`
			Result []Update `json:"result"`
		}

		json.Unmarshal(body, &apiResponse)
		if !apiResponse.OK {
			log.Printf("API error: %v", string(body))
			continue
		}

		for _, upd := range apiResponse.Result {
			offset = upd.UpdateID + 1

			go func(upd Update) {
				handleUpdate(upd, cfg)
			}(upd)
		}
	}
}

func handleUpdate(upd Update, cfg config.Config) {
	if upd.Message.Text != "" {
		log.Printf("Received: `%s` from %s (@%s)",
			upd.Message.Text, upd.Message.Chat.Name, upd.Message.Chat.Username)
		
		handleTextMessage(upd, cfg)
	}
}
