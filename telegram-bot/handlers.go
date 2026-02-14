package main

import (
	"fmt"
	"log"
	"strconv"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

var Subjects = map[string]map[int][]string{
	"МАТЕМАТИКА": {
		5: {"Натуральные числа: действия, свойства, делимость, степени", "Отрицательные числа", "Дроби", "Площадь фигур", "Умножение и деление"},
		6: {"Обыкновенные дроби", "Проценты", "Уравнения", "Площадь и периметр"},
		7: {"Линейные уравнения", "Системы уравнений, неравенства", "Квадратные уравнения", "Подобие фигур"},
		8: {"Квадратные уравнения", "Степени", "Тригонометрия", "Координатная плоскость"},
		9: {"Логарифмы", "Производные", "Интегралы", "Тригонометрические функции"},
	},
	"ИНФОРМАТИКА": {
		5: {"Алгоритмы", "Переменные", "Циклы", "Условия"},
		6: {"Массивы", "Функции", "Алгоритмы сортировки", "Файлы"},
		7: {"Структуры данных", "Рекурсия", "Графы", "Базы данных"},
		8: {"Объектно-ориентированное программирование", "Сети", "Базы данных SQL", "Алгоритмы поиска"},
		9: {"Машинное обучение", "Криптография", "Параллельное программирование", "Большие данные"},
	},
	"РУССКИЙ ЯЗЫК": {
		5: {"Правописание Н/НН", "Приставки", "Суффиксы", "Склонение"},
		6: {"Причастия", "Действительные причастия", "Сложные предложения", "Прямая речь"},
		7: {"Дієприкметники", "Условное наклонение", "Синтаксис", "Пунктуация"},
		8: {"Стилистика", "Лексика", "Фразеологизмы", "Орфоэпия"},
		9: {"Анализ текста", "Сочинение-рассуждение", "Литературные тропы", "Диалекты"},
	},
}

func GetThemes(subject string, class int) []string {
	log.Printf("GetThemes: subject='%s' class=%d", subject, class)

	if themes, ok := Subjects[subject]; ok {
		if t, ok := themes[class]; ok && len(t) > 0 {
			log.Printf("✅ Найдены темы: %v", t)
			return t
		} else {
			log.Printf("❌ Темы для класса %d не найдены", class)
		}
	} else {
		log.Printf("❌ Предмет '%s' не найден", subject)
	}
	return []string{"Тема не найдена"}
}

func parseClass(classStr string) int {
	log.Printf("parseClass input: '%s'", classStr)

	if num, err := strconv.Atoi(classStr); err == nil {
		log.Printf("parseClass result: %d (from Atoi)", num)
		return num
	}

	switch classStr {
	case "5 класс":
		return 5
	case "6 класс":
		return 6
	case "7 класс":
		return 7
	case "8 класс":
		return 8
	case "9 класс":
		return 9
	}
	log.Printf("parseClass default: 5")
	return 5
}

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
	if update.Message == nil {
		return
	}

	chatID := update.Message.Chat.ID
	userID := update.Message.From.ID
	text := update.Message.Text

	// 1️⃣ GIGA CHECK
	stateObj, exists := GetUserState(userID)
	if exists && stateObj.State == "waiting_answer" {
		task := GetUserTask(userID)
		feedback := CheckAnswer(task, text)

		msg := tgbotapi.NewMessage(chatID, feedback)
		msg.ParseMode = "Markdown"
		bot.Send(msg)
		DeleteUserState(userID)
		return
	}

	// 2️⃣ РЕГИСТРАЦИЯ
	stateStr, dataStr, ok := GetUserStateString(userID)
	if ok {
		switch stateStr {
		case StateWaitingUsernameReg:
			SaveUserState(userID, StateWaitingPasswordReg, text)
			bot.Send(tgbotapi.NewMessage(chatID, "🔐 Введите пароль:"))
			return
		case StateWaitingPasswordReg:
			RegisterUser(dataStr, text, userID)
			DeleteUserState(userID)
			bot.Send(tgbotapi.NewMessage(chatID, "✅ Зарегистрирован!"))
			return
		}
	}

	// 3️⃣ ОСНОВНОЕ МЕНЮ
	msg := tgbotapi.NewMessage(chatID, "🔐 Регистрация\n🆕 Зарегистрироваться")
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
			"math": "МАТЕМАТИКА",
			"info": "ИНФОРМАТИКА",
			"rus":  "РУССКИЙ ЯЗЫК",
		}
		subjectName := subjectNames[subject]
		classNum, _ := GetUserClass(userID)

		msgText := fmt.Sprintf("📚 **%s** (%s класс)\n\nВыберите тему:", subjectName, classNum)
		msg := tgbotapi.NewMessage(chatID, msgText)
		msg.ParseMode = "Markdown"

		// ✅ Парсим класс в int
		classInt := parseClass(classNum)
		log.Printf("DEBUG: subject='%s' classInt=%d", subjectName, classInt)
		themes := GetThemes(subjectName, classInt)

		keyboard := tgbotapi.NewInlineKeyboardMarkup()
		for i, theme := range themes {
			// ✅ Фиксированные короткие префиксы!
			prefix := map[string]string{
				"МАТЕМАТИКА":   "МА",
				"ИНФОРМАТИКА":  "ИНО",
				"РУССКИЙ ЯЗЫК": "РУС",
			}[subjectName]

			// ✅ Правильный формат: "МА_7_easy_THEME1"
			callback := fmt.Sprintf("%s_%d_easy_THEME%d", prefix, classInt, i+1)
			log.Printf("Создан callback: '%s'", callback)

			keyboard.InlineKeyboard = append(keyboard.InlineKeyboard,
				tgbotapi.NewInlineKeyboardRow(tgbotapi.NewInlineKeyboardButtonData(theme, callback)))
		}
		keyboard.InlineKeyboard = append(keyboard.InlineKeyboard,
			tgbotapi.NewInlineKeyboardRow(tgbotapi.NewInlineKeyboardButtonData("🔙 Назад", "start_lessons")))

		msg.ReplyMarkup = keyboard
		bot.Send(msg)

	default:
		if strings.HasPrefix(data, "МА_") || strings.HasPrefix(data, "ИНО_") || strings.HasPrefix(data, "РУС_") {
			parts := strings.Split(data, "_")
			if len(parts) != 4 {
				bot.Request(tgbotapi.NewCallback(cb.ID, "Ошибка формата"))
				return
			}

			subjectCode := parts[0] // "МА", "ИНО", "РУС"
			classStr := parts[1]    // "7"
			difficulty := parts[2]  // "easy"
			themeStr := parts[3]    // "THEME1"

			classNum, err := strconv.Atoi(classStr)
			if err != nil {
				log.Printf("Ошибка парсинга класса: %v", err)
				return
			}

			subjectNames := map[string]string{
				"МА":  "МАТЕМАТИКА",
				"ИНО": "ИНФОРМАТИКА",
				"РУС": "РУССКИЙ ЯЗЫК",
			}
			subject := subjectNames[subjectCode]

			themeIdx, err := strconv.Atoi(themeStr[5:]) // "THEME1" → "1" → 1
			if err != nil {
				log.Printf("Ошибка парсинга темы: %v", err)
				return
			}

			themes := GetThemes(subject, classNum)
			if themeIdx-1 >= len(themes) || themeIdx-1 < 0 {
				log.Printf("Неверный индекс темы: %d", themeIdx)
				return
			}
			topic := themes[themeIdx-1]

			task := GenerateTask(subject, topic, difficulty, fmt.Sprintf("%d класс", classNum))
			SaveUserState(userID, "waiting_answer", task)

			msgText := fmt.Sprintf("🎯 **%s** (%d класс)\n📖 **Тема:** %s\n\n%s\n\n📝 Напишите ответ:",
				strings.ToUpper(difficulty), classNum, topic, task)

			msg := tgbotapi.NewMessage(chatID, msgText)
			msg.ParseMode = "Markdown"
			msg.ReplyMarkup = tgbotapi.NewInlineKeyboardMarkup(
				tgbotapi.NewInlineKeyboardRow(tgbotapi.NewInlineKeyboardButtonData("🔙 Главное меню", "main")),
			)
			bot.Send(msg)
			return
		}

		log.Printf("Неизвестный callback: %s", data)
		bot.Request(tgbotapi.NewCallback(cb.ID, "Неизвестная команда"))
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
func getThemeButtons(subject, classStr, difficulty string, classNum int) [][]tgbotapi.InlineKeyboardButton {
	class := parseClass(classStr)
	themes := GetThemes(subject, class)
	buttons := [][]tgbotapi.InlineKeyboardButton{}

	subjectPrefix := map[string]string{
		"МАТЕМАТИКА":   "МА",
		"ИНФОРМАТИКА":  "ИНО",
		"РУССКИЙ ЯЗЫК": "РУС",
	}[subject]

	for i, theme := range themes {
		callback := fmt.Sprintf("%s_%d_%s_THEME%d", subjectPrefix, class, difficulty[:3], i+1)
		buttons = append(buttons, tgbotapi.NewInlineKeyboardRow(tgbotapi.NewInlineKeyboardButtonData(theme, callback)))
	}
	return buttons
}
