package main

import (
	"database/sql"
	"fmt"
	"log"

	_ "github.com/lib/pq"
)

var db *sql.DB

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
func GetUserState(telegramID int64) (string, string, error) {
	if stateData, exists := userStates[telegramID]; exists {
		return stateData.State, stateData.Data, nil
	}
	return "", "", nil
}

func DeleteUserState(telegramID int64) error {
	_, err := db.Exec(`DELETE FROM user_states WHERE telegram_id = $1`, telegramID)
	return err
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
