package main

import (
	"log"
	"os"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func main() {

	token := "8303723414:AAHN3_C5U8vtKOour2A0H8QJL6Ck0Vnmbxk"
	if token == "" {
		log.Fatal("BOT_TOKEN is empty. Set it like: export BOT_TOKEN='123:ABC'")
	}

	dbConn := os.Getenv("postgres://postgres:Bkmz_2009@localhost:5432/TGbot?sslmode=disable")
	if dbConn == "" {
		dbConn = "postgres://postgres:Bkmz_2009@localhost:5432/TGbot?sslmode=disable"
	}
	InitDB(dbConn)

	bot, err := tgbotapi.NewBotAPI(token)
	if err != nil {
		log.Fatal(err)
	}

	log.Printf("Авторизован как %s", bot.Self.UserName)

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 30
	updates := bot.GetUpdatesChan(u)

	for update := range updates {
		if update.Message != nil {
			command := update.Message.Command()

			if update.Message.IsCommand() {
				if command == "start" {
					handleStart(bot, update)
				} else if command == "clear" {
					handleClear(bot, update)
				} else if command == "accdelete" {
					handleAccDelete(bot, update)
				}
			} else {
				handleMessage(bot, update)
			}
		}
		if update.CallbackQuery != nil {
			handleCallback(bot, *update.CallbackQuery)
		}
	}

}
