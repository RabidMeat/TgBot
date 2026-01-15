package main

import (
	"log"
	"os"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func main() {
	// Токен бота из переменной окружения

	token := "8303723414:AAHN3_C5U8vtKOour2A0H8QJL6Ck0Vnmbxk" // положи токен в переменную окружения BOT_TOKEN
	if token == "" {
		log.Fatal("BOT_TOKEN is empty. Set it like: export BOT_TOKEN='123:ABC'")
	}

	// Строка подключения к PostgreSQL
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

			// Обрабатываем команды
			if update.Message.IsCommand() {
				if command == "start" || command == "menu" {
					handleStart(bot, update)
				} else if command == "clear" {
					handleClear(bot, update)
				} else if command == "accdelete" {
					handleAccDelete(bot, update) // Новая команда!
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
