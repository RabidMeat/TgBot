package main

import (
	"log"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func main() {
	token := "" // положи токен в переменную окружения BOT_TOKEN
	if token == "8303723414:AAHN3_C5U8vtKOour2A0H8QJL6Ck0Vnmbxk" {
		log.Fatal("BOT_TOKEN is empty. Set it like: export BOT_TOKEN='123:ABC'")
	}

	bot, err := tgbotapi.NewBotAPI(token)
	if err != nil {
		log.Fatal(err)
	}

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 30

	updates := bot.GetUpdatesChan(u)

	for update := range updates {
		// 1) Команды/сообщения
		if update.Message != nil {
			if update.Message.IsCommand() && update.Message.Command() == "start" {
				msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Выбери кнопку 👇")

				// 3 кнопки в один ряд (inline keyboard)
				keyboard := tgbotapi.NewInlineKeyboardMarkup(
					tgbotapi.NewInlineKeyboardRow(
						tgbotapi.NewInlineKeyboardButtonData("Кнопка 1", "btn_1"),
						tgbotapi.NewInlineKeyboardButtonData("Кнопка 2", "btn_2"),
						tgbotapi.NewInlineKeyboardButtonData("Кнопка 3", "btn_3"),
					),
				)

				msg.ReplyMarkup = keyboard
				if _, err := bot.Send(msg); err != nil {
					log.Println("send error:", err)
				}
			}
		}

		// 2) Нажатия на inline-кнопки (callback)
		if update.CallbackQuery != nil {
			cb := update.CallbackQuery
			data := cb.Data

			var text string
			switch data {
			case "btn_1":
				text = "Ты нажал: Кнопка 1"
			case "btn_2":
				text = "Ты нажал: Кнопка 2"
			case "btn_3":
				text = "Ты нажал: Кнопка 3"
			default:
				text = "Неизвестная кнопка"
			}

			// обязательно отвечаем на callback, чтобы “часики” исчезли
			_, _ = bot.Request(tgbotapi.NewCallback(cb.ID, ""))

			// отправим сообщение в чат
			msg := tgbotapi.NewMessage(cb.Message.Chat.ID, text)
			if _, err := bot.Send(msg); err != nil {
				log.Println("send error:", err)
			}
		}
	}
}
