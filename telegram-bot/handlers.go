package main

import (
	"fmt"
	"log"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

const (
	StateWaitingUsernameReg = "waiting_username_reg" // ← ТОЧНО такая строка!
	StateWaitingPasswordReg = "waiting_password_reg"
)

func handleStart(bot *tgbotapi.BotAPI, update tgbotapi.Update) {
	userID := update.Message.From.ID
	chatID := update.Message.Chat.ID

	isAuth := IsUserAuthorized(userID)
	username, _ := GetUserUsername(userID)

	var msgText string
	var keyboard tgbotapi.InlineKeyboardMarkup

	if isAuth {
		msgText = fmt.Sprintf("👋 @%s - вы уже зарегистрированы!", username)
		// ✅ ТОЛЬКО кнопка Главное меню
		keyboard = tgbotapi.NewInlineKeyboardMarkup(
			tgbotapi.NewInlineKeyboardRow(
				tgbotapi.NewInlineKeyboardButtonData("🏠 Главное меню", "main"),
			),
		)
	} else {
		msgText = "🔐 Регистрация"
		keyboard = tgbotapi.NewInlineKeyboardMarkup(
			tgbotapi.NewInlineKeyboardRow(
				tgbotapi.NewInlineKeyboardButtonData("🆕 Зарегистрироваться", "reg"),
			),
		)
	}

	msg := tgbotapi.NewMessage(chatID, msgText)
	msg.ReplyMarkup = keyboard
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
			msgText := fmt.Sprintf("🏠Главное меню:\n\n✅ В системе: @%s\n\n📋 **Команды**:\n/clear - очистить чат\n/accdelete - удалить аккаунт", username)
			msg := tgbotapi.NewMessage(chatID, msgText)
			msg.ParseMode = "Markdown"
			msg.ReplyMarkup = getMainMenuKeyboardWithLessons(username) // ✅ НОВЫЕ кнопки!
			bot.Send(msg)
		} else {
			msgText := "🔐 **Регистрация**"
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
		username, _ := GetUserUsername(userID)
		msg := tgbotapi.NewMessage(chatID, "✨ **Удаление отменено**\n\nВы по-прежнему в системе")
		msg.ParseMode = "Markdown"
		msg.ReplyMarkup = getMainKeyboard(true, username)
		bot.Send(msg)
	case "start_lessons":
		// ✅ Для handleMessage() используем message
		username, _ := GetUserUsername(userID)
		if username == "" {
			username = "Пользователь"
		}

		msgText := fmt.Sprintf("📚 **Выбор класса**\n\n@%s, выберите класс:", username)

		msg := tgbotapi.NewMessage(chatID, msgText)
		msg.ParseMode = "Markdown"

		keyboard := tgbotapi.NewInlineKeyboardMarkup(
			tgbotapi.NewInlineKeyboardRow(
				tgbotapi.NewInlineKeyboardButtonData("📖 5 класс", "class_5"),
				tgbotapi.NewInlineKeyboardButtonData("📖 6 класс", "class_6"),
			),
			tgbotapi.NewInlineKeyboardRow(
				tgbotapi.NewInlineKeyboardButtonData("📖 7 класс", "class_7"),
				tgbotapi.NewInlineKeyboardButtonData("📖 8 класс", "class_8"),
			),
			tgbotapi.NewInlineKeyboardRow(
				tgbotapi.NewInlineKeyboardButtonData("📖 9 класс", "class_9"),
			),
		)
		msg.ReplyMarkup = keyboard
		bot.Send(msg)

	case "class_5", "class_6", "class_7", "class_8", "class_9":
		classNum := strings.TrimPrefix(data, "class_")
		username, _ := GetUserUsername(userID)

		// ✅ СОХРАНЯЕМ выбранный класс в состоянии
		SaveUserClass(userID, classNum)

		msgText := fmt.Sprintf("📚 **%s класс**\n\n@%s, выберите предмет:", classNum, username)

		msg := tgbotapi.NewMessage(chatID, msgText)
		msg.ParseMode = "Markdown"

		// ✅ Кнопки предметов: Математика, Информатика, Русский язык
		keyboard := tgbotapi.NewInlineKeyboardMarkup(
			tgbotapi.NewInlineKeyboardRow(
				tgbotapi.NewInlineKeyboardButtonData("📐 Математика", "subject_math"),
				tgbotapi.NewInlineKeyboardButtonData("💻 Информатика", "subject_info"),
			),
			tgbotapi.NewInlineKeyboardRow(
				tgbotapi.NewInlineKeyboardButtonData("📖 Русский язык", "subject_rus"),
			),
			tgbotapi.NewInlineKeyboardRow(
				tgbotapi.NewInlineKeyboardButtonData("🔙 Выбор класса", "start_lessons"),
			),
		)
		msg.ReplyMarkup = keyboard
		bot.Send(msg)
	case "subject_math", "subject_info", "subject_rus":
		subject := strings.TrimPrefix(data, "subject_")
		subjectNames := map[string]string{
			"math": "Математика",
			"info": "Информатика",
			"rus":  "Русский язык",
		}
		subjectName := subjectNames[subject]

		classNum, _ := GetUserClass(userID)
		username, _ := GetUserUsername(userID)

		msgText := fmt.Sprintf("📚 **%s** (%s класс)\n\n@%s, выберите тему урока:", subjectName, classNum, username)

		msg := tgbotapi.NewMessage(chatID, msgText)
		msg.ParseMode = "Markdown"

		// ✅ Кнопки тем уроков
		keyboard := tgbotapi.NewInlineKeyboardMarkup(
			tgbotapi.NewInlineKeyboardRow(
				tgbotapi.NewInlineKeyboardButtonData("📖 Тема 1", "topic_1"),
				tgbotapi.NewInlineKeyboardButtonData("📖 Тема 2", "topic_2"),
			),
			tgbotapi.NewInlineKeyboardRow(
				tgbotapi.NewInlineKeyboardButtonData("📖 Тема 3", "topic_3"),
			),
			tgbotapi.NewInlineKeyboardRow(
				tgbotapi.NewInlineKeyboardButtonData("🔙 Выбор предмета", "class_"+classNum),
			),
		)
		msg.ReplyMarkup = keyboard
		bot.Send(msg)
	case "topic_1", "topic_2", "topic_3":
		topicNum := strings.TrimPrefix(data, "topic_")
		classNum, _ := GetUserClass(userID)

		msgText := fmt.Sprintf("📖 **Тема %s** (%s класс)\n\nВыберите сложность:",
			topicNum, classNum)

		msg := tgbotapi.NewMessage(chatID, msgText)
		msg.ParseMode = "Markdown"
		// ✅ Кнопки сложности
		keyboard := tgbotapi.NewInlineKeyboardMarkup(
			tgbotapi.NewInlineKeyboardRow(
				tgbotapi.NewInlineKeyboardButtonData("🟢 Легко", "difficulty_easy_"+topicNum),
				tgbotapi.NewInlineKeyboardButtonData("🟡 Средне", "difficulty_medium_"+topicNum),
			),
			tgbotapi.NewInlineKeyboardRow(
				tgbotapi.NewInlineKeyboardButtonData("🔴 Сложно", "difficulty_hard_"+topicNum),
			),
			tgbotapi.NewInlineKeyboardRow(
				tgbotapi.NewInlineKeyboardButtonData("📚 Выбор темы", "subject_math"), // вернитесь к предмету
			),
		)
		msg.ReplyMarkup = keyboard
		bot.Send(msg)
	case "difficulty_easy_1", "difficulty_easy_2", "difficulty_easy_3",
		"difficulty_medium_1", "difficulty_medium_2", "difficulty_medium_3",
		"difficulty_hard_1", "difficulty_hard_2", "difficulty_hard_3":

		parts := strings.Split(data, "_")
		difficulty := parts[1] // easy, medium, hard
		topicNum := parts[2]   // 1, 2, 3

		difficultyNames := map[string]string{
			"easy":   "🟢 Легко",
			"medium": "🟡 Средне",
			"hard":   "🔴 Сложно",
		}

		classNum, _ := GetUserClass(userID)
		username, _ := GetUserUsername(userID)

		msgText := fmt.Sprintf("🎯 **%s - Тема %s** (%s класс)\n\n✅ Задание готово!\n\n@%s",
			difficultyNames[difficulty], topicNum, classNum, username)

		msg := tgbotapi.NewMessage(chatID, msgText)
		msg.ParseMode = "Markdown"

		keyboard := tgbotapi.NewInlineKeyboardMarkup(
			tgbotapi.NewInlineKeyboardRow(
				tgbotapi.NewInlineKeyboardButtonData("🔙 Главное меню", "main"),
			),
		)
		msg.ReplyMarkup = keyboard
		bot.Send(msg)

	}

}

func getMainKeyboard(isAuthorized bool, username string) tgbotapi.InlineKeyboardMarkup {
	if isAuthorized {
		return tgbotapi.NewInlineKeyboardMarkup(
			tgbotapi.NewInlineKeyboardRow(
				tgbotapi.NewInlineKeyboardButtonData("🏠 Главное меню", "main"),
			),
		)
	}
	return tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("🆕 Зарегистрироваться", "reg"),
		),
	)
}
func getMainMenuKeyboardWithLessons(username string) tgbotapi.InlineKeyboardMarkup {
	return tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("📚 Начать занятия", "start_lessons"),
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
