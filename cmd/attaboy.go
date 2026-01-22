package main

import (
	"log"

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

	bot.PollUpdates(0, *cfg)
}
