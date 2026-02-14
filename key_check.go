package main

import (
	"fmt"
	"log"

	"github.com/evgensoft/gigachat"
)

func main() {
	client := gigachat.NewClient("019c32f9-459f-78c3-9ab3-435e204b8aa4", "4592538d-1eb5-479b-8f54-668a170ead20")

	req := &gigachat.ChatRequest{
		Model: gigachat.ModelGigaChat,
		Messages: []gigachat.Message{
			{Role: "user", Content: "Составь тестик по матеметике на 5 вопросов"},
		},
	}

	resp, err := client.Chat(req)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(resp.Choices[0].Message.Content)
}
