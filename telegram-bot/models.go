package main

type User struct {
	ID         int
	Username   string
	Password   string
	TelegramID int64
}

type UserState struct {
	TelegramID int64
	State      string
	Data       string
}
