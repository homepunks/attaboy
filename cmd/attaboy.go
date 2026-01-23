package main

import (
	"log"

	"github.com/homepunks/attaboy/internal/bot"
	"github.com/homepunks/attaboy/internal/config"
)

func main() {
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Printf("Could not start the bot: %v\n", err)
		return
	}

	log.Println("Attaboy!")

	bot.PollUpdates(0, *cfg)
}
