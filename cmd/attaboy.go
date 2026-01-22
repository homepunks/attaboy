package main

import (
	"log"
	"time"
	"fmt"
	"net/http"
	"encoding/json"
	"io"

	"github.com/homepunks/attaboy/internal/config"
	"github.com/homepunks/attaboy/internal/bot"
)

func main() {
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Printf("Could not start the bot: %v\n", err)
		return
	}

	log.Println("Attaboy!")

	pollUpdates(0, *cfg)
}

func pollUpdates(offset int64, cfg config.Config) {
	for {
		url := fmt.Sprintf("%s%s/getUpdates?offset=%d&timeout=30",
			cfg.BaseURL, cfg.BotToken, offset)
		
		resp, err := http.Get(url)
		if err != nil {
			log.Printf("Error getting updates: %v\n", err)
			time.Sleep(5 * time.Second)
			continue
		}

		body, _ := io.ReadAll(resp.Body)
		resp.Body.Close()

		var apiResponse struct {
			OK bool `json:"ok"`
			Result []bot.Update `json:"result"`
		}

		json.Unmarshal(body, &apiResponse)
		if !apiResponse.OK {
			log.Printf("API error: %v", string(body))
			continue
		}

		for _, upd := range apiResponse.Result {
			offset = upd.UpdateID + 1

			handleUpdate(upd, cfg)
		}
	}
}

func handleUpdate(upd bot.Update, cfg config.Config) {
	if upd.Message.Text != "" {
		bot.HandleTextMessage(upd, cfg)
	}
}
