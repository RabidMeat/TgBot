package main

import (
	"fmt"
	"log"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

const (
	StateWaitingUsernameReg = "waiting_username_reg" // ← ТОЧНО такая строка!
	StateWaitingPasswordReg = "waiting_password_reg"
)

func handleStart(bot *tgbotapi.BotAPI, update tgbotapi.Update) {
	userID := update.Message.From.ID
	chatID := update.Message.Chat.ID

	// Проверяем команды /start И /menu
	command := update.Message.Command()
	if command != "start" && command != "menu" {
		return
	}

	isAuth := IsUserAuthorized(userID)
	username, _ := GetUserUsername(userID)

	var msgText string
	if isAuth {
		msgText = fmt.Sprintf("👋 @%s - вы уже зарегистрированы!\n", username)
	} else {
		msgText = "🔐 Регистрация"
	}

	msg := tgbotapi.NewMessage(chatID, msgText)
	msg.ReplyMarkup = getMainKeyboard(isAuth, username)
	bot.Send(msg)
}

func handleMessage(bot *tgbotapi.BotAPI, update tgbotapi.Update) {
	chatID := update.Message.Chat.ID
	userID := update.Message.From.ID
	text := update.Message.Text
	state, data, _ := GetUserState(userID)
	// ✅ ТЕСТ РЕГИСТРАЦИИ
	if state == StateWaitingUsernameReg {
		SaveUserState(userID, StateWaitingPasswordReg, text)
		msg := tgbotapi.NewMessage(chatID, "🔐 Введите пароль:")
		bot.Send(msg)
		return
	}

	if state == StateWaitingPasswordReg {
		username := data
		RegisterUser(username, text, userID)
		DeleteUserState(userID)
		msg := tgbotapi.NewMessage(chatID, fmt.Sprintf("✅ Регистрация: @%s", username))
		bot.Send(msg)
		return
	}

	// Всё остальное
	msg := tgbotapi.NewMessage(chatID, "🔐 Регистрация\nНажмите 🆕 Зарегистрироваться")
	msg.ReplyMarkup = getMainKeyboard(false, "")
	bot.Send(msg)
}

func handleCallback(bot *tgbotapi.BotAPI, cb tgbotapi.CallbackQuery) {
	chatID := cb.Message.Chat.ID
	userID := cb.From.ID
	data := cb.Data

	// Обязательно отвечаем на callback
	bot.Request(tgbotapi.NewCallback(cb.ID, ""))

	switch data {
	case "reg":
		SaveUserState(userID, StateWaitingUsernameReg, "")
		msg := tgbotapi.NewMessage(chatID, "📝 **Введите имя пользователя:**")
		msg.ParseMode = "Markdown"
		bot.Send(msg)

	case "main":
		isAuth := IsUserAuthorized(userID)
		username, _ := GetUserUsername(userID)

		if isAuth {
			msgText := fmt.Sprintf("🏠Главное меню:\n\n✅ В системе: @%s\n\n📋 **Команды**:\n/menu - главное меню\n/clear - очистить чат\n/accdelete - удалить аккаунт", username)
			msg := tgbotapi.NewMessage(chatID, msgText)
			msg.ParseMode = "Markdown"

			bot.Send(msg)
		} else {
			msgText := "🔐 **Регистрация**\n\nНажмите кнопку для начала"
			msg := tgbotapi.NewMessage(chatID, msgText)
			msg.ParseMode = "Markdown"
			msg.ReplyMarkup = getMainKeyboard(false, "")
			bot.Send(msg)
		}

	case "delete_confirm":
		// НЕ ИСПОЛЬЗУЕТСЯ - только через /accdelete

	case "delete_yes":
		// ✅ ФИНАЛЬНОЕ УДАЛЕНИЕ АККАУНТА
		log.Printf("🗑️ Удаляем аккаунт userID: %d", userID)
		_, err := db.Exec("DELETE FROM users WHERE telegram_id = $1", userID)
		if err != nil {
			log.Println("Ошибка удаления:", err)
		}
		DeleteUserState(userID)

		msg := tgbotapi.NewMessage(chatID, "🗑️ ✅ **АККАУНТ ПОЛНОСТЬЮ УДАЛЁН**\n")
		msg.ParseMode = "Markdown"
		msg.ReplyMarkup = getMainKeyboard(false, "")
		bot.Send(msg)

	case "delete_no":
		// ✅ ОТМЕНА УДАЛЕНИЯ
		username, _ := GetUserUsername(userID)
		msg := tgbotapi.NewMessage(chatID, "✨ **Удаление отменено**\n\nВы по-прежнему в системе")
		msg.ParseMode = "Markdown"
		msg.ReplyMarkup = getMainKeyboard(true, username)
		bot.Send(msg)
	}
}

func getMainKeyboard(isAuthorized bool, username string) tgbotapi.InlineKeyboardMarkup {
	if isAuthorized {
		return tgbotapi.NewInlineKeyboardMarkup(
			tgbotapi.NewInlineKeyboardRow(
				tgbotapi.NewInlineKeyboardButtonData("🏠 Главное меню", "main"),
			),
			tgbotapi.NewInlineKeyboardRow(
				tgbotapi.NewInlineKeyboardButtonData("📋 /menu /clear /accdelete", "main"),
			),
		)
	}
	return tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("🆕 Зарегистрироваться", "reg"),
		),
	)
}

func handleClear(bot *tgbotapi.BotAPI, update tgbotapi.Update) {
	chatID := update.Message.Chat.ID

	// Уведомляем что начали очистку
	clearMsg := tgbotapi.NewMessage(chatID, "🧹 Очищаю чат...")
	sentMsg, _ := bot.Send(clearMsg)

	// Удаляем сообщения Бота начиная с текущего и назад (максимум 100 сообщений)
	currentMsgID := sentMsg.MessageID
	for i := currentMsgID; i > currentMsgID-100 && i > 0; i-- {
		_, err := bot.Request(tgbotapi.NewDeleteMessage(chatID, i))
		if err != nil {
			// Сообщение уже удалено или недоступно - продолжаем
			continue
		}
	}

	// Отправляем чистое меню
	isAuth := IsUserAuthorized(update.Message.From.ID)
	username, _ := GetUserUsername(update.Message.From.ID)
	msgText := "✅ Чат очищен!"
	if isAuth {
		msgText = fmt.Sprintf("✅ Чат очищен, @%s!", username)
	}

	msg := tgbotapi.NewMessage(chatID, msgText)
	msg.ReplyMarkup = getMainKeyboard(isAuth, username)
	bot.Send(msg)
}

func handleAccDelete(bot *tgbotapi.BotAPI, update tgbotapi.Update) {
	userID := update.Message.From.ID
	chatID := update.Message.Chat.ID

	if !IsUserAuthorized(userID) {
		msg := tgbotapi.NewMessage(chatID, "❌ Нет аккаунта!")
		msg.ReplyMarkup = getMainKeyboard(false, "")
		bot.Send(msg)
		return
	}

	username, _ := GetUserUsername(userID)
	msg := tgbotapi.NewMessage(chatID, fmt.Sprintf("⚠️ **УДАЛИТЬ АККАУНТ @%s?**\n\nЭто **необратимо**!", username))
	msg.ParseMode = "Markdown"

	// Кнопки ДА/НЕТ
	keyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("✅ ДА, УДАЛИТЬ", "delete_yes"),
			tgbotapi.NewInlineKeyboardButtonData("❌ НЕТ, ОТМЕНА", "delete_no"),
		),
	)
	msg.ReplyMarkup = keyboard
	bot.Send(msg)
}
