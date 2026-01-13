package main

import (
	"log"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func main() {
	token := "8303723414:AAHN3_C5U8vtKOour2A0H8QJL6Ck0Vnmbxk" // положи токен в переменную окружения BOT_TOKEN
	if token == "" {
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
				msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Авторизация")

				// 2 кнопки в один ряд (inline keyboard)
				keyboard := tgbotapi.NewInlineKeyboardMarkup(
					tgbotapi.NewInlineKeyboardRow(
						tgbotapi.NewInlineKeyboardButtonData("Зарегистрироваться", "reg"),
						tgbotapi.NewInlineKeyboardButtonData("Войти", "sign_in"),
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
			case "reg":
				text = "Эта функция пока не доступна"
			case "sign_in":
				text = "Эта функция пока не доступна"
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
