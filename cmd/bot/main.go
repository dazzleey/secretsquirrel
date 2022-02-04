package main

import (
	"log"
	"secretsquirrel/config"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

var cfg config.Config

func init() {
	config.LoadConfig(&cfg)
}

func main() {
	bot := initBot()

	updateConfig := tgbotapi.NewUpdate(0)
	updateConfig.Timeout = 60
	updates := bot.Api.GetUpdatesChan(updateConfig)

	log.Printf("Authorized on account %s", bot.Api.Self.UserName)

	for u := range updates {
		bot.handleUpdate(u)
	}
}
