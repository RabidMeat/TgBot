package main

type User struct {
	ID         int    `json:"id"`
	Username   string `json:"username"`
	Password   string `json:"password"` // В продакшене используйте хэширование!
	TelegramID int64  `json:"telegram_id"`
}

type UserState struct {
	TelegramID int64  `json:"telegram_id"`
	State      string `json:"state"`
	Data       string `json:"data"` // временные данные (имя пользователя)
}

// In-memory хранилище состояний (если не используете БД для состояний)
var userStates = make(map[int64]UserState)
