package main

import (
	"database/sql"
	"fmt"
	"log"

	_ "github.com/lib/pq"
)

var db *sql.DB
var userStates map[int64]UserState
var userClasses map[int64]string // ← Только здесь!

func init() {
	userStates = make(map[int64]UserState)
	userClasses = make(map[int64]string)
}

func SaveUserClass(telegramID int64, class string) {
	userClasses[telegramID] = class
}

func GetUserClass(telegramID int64) (string, bool) {
	class, exists := userClasses[telegramID]
	return class, exists
}
func InitDB(connStr string) {
	var err error
	db, err = sql.Open("postgres", connStr)
	if err != nil {
		log.Fatal("Ошибка подключения к БД:", err)
	}

	// Тест подключения
	err = db.Ping()
	if err != nil {
		log.Fatal("Не удается подключиться к БД:", err)
	}

	log.Println("✅ Подключение к существующей базе данных успешно")
}

func SaveUserState(telegramID int64, state, data string) {
	userStates[telegramID] = UserState{
		TelegramID: telegramID,
		State:      state,
		Data:       data,
	}
}

// В database.go измените:
func GetUserState(telegramID int64) (UserState, bool) { // ← 2 возврата!
	state, exists := userStates[telegramID]
	return state, exists
}

func DeleteUserState(telegramID int64) {
	delete(userStates, telegramID)
}

func RegisterUser(username, password string, telegramID int64) error {
	_, err := db.Exec(
		`INSERT INTO users (username, password, telegram_id) VALUES ($1, $2, $3)`,
		username, password, telegramID,
	)
	return err
}

func CheckUser(username, password string) (int64, error) {
	var telegramID int64
	err := db.QueryRow(
		`SELECT telegram_id FROM users WHERE username = $1 AND password = $2`,
		username, password,
	).Scan(&telegramID)

	if err == sql.ErrNoRows {
		return 0, fmt.Errorf("пользователь не найден")
	}
	return telegramID, err
}

// Проверяем, авторизован ли пользователь в системе
func IsUserAuthorized(telegramID int64) bool {
	var count int
	err := db.QueryRow("SELECT COUNT(*) FROM users WHERE telegram_id = $1", telegramID).Scan(&count)
	if err != nil {
		return false
	}
	return count > 0 // ✅ Возвращаем bool правильно
}

// Получаем имя пользователя (если авторизован)
func GetUserUsername(telegramID int64) (string, bool) {
	var username string
	err := db.QueryRow(
		`SELECT username FROM users WHERE telegram_id = $1`,
		telegramID,
	).Scan(&username)

	return username, err == nil
}

func GetUserTask(telegramID int64) string {
	state, exists := userStates[telegramID]
	if exists {
		return state.Data
	}
	return ""
}
func GetUserStateString(telegramID int64) (string, string, bool) {
	state, exists := userStates[telegramID]
	if exists {
		return state.State, state.Data, true
	}
	return "", "", false
}
