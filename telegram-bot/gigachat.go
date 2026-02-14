package main

import (
	"fmt"
	"log"

	"github.com/evgensoft/gigachat"
)

var client *gigachat.Client

func init() {
	// ✅ ВАШИ РАБОЧИЕ ключи из примера!
	client = gigachat.NewClient("019c32f9-459f-78c3-9ab3-435e204b8aa4", "4592538d-1eb5-479b-8f54-668a170ead20")
	fmt.Println("✅ GigaChat клиент инициализирован!")
}

func GenerateTask(subject, topic, difficulty, classNum string) string {
	req := &gigachat.ChatRequest{
		Model: gigachat.ModelGigaChat,
		Messages: []gigachat.Message{
			{
				Role: "user",
				Content: fmt.Sprintf(`%s, %s класс. Тема: "%s". Сложность: %s.
                
Создай задание в формате вопроса и 4 вариантов ответов A-D.
Сам ответ не пиши.
                
Без объяснений!`, subject, classNum, topic, difficulty),
			},
		},
	}

	resp, err := client.Chat(req)
	if err != nil {
		log.Printf("GigaChat GenerateTask error: %v", err)
		return "❌ Ошибка генерации задания"
	}

	return resp.Choices[0].Message.Content
}

func CheckAnswer(task, userAnswer string) string {
	req := &gigachat.ChatRequest{
		Model: gigachat.ModelGigaChat,
		Messages: []gigachat.Message{
			{
				Role: "user",
				Content: fmt.Sprintf(`Проверь ответ:

ЗАДАНИЕ:
%s

ОТВЕТ: %s

Ответь: ✅ ПРАВИЛЬНО! Объяснение. ИЛИ ❌ НЕПРАВИЛЬНО. Правильный: X`, task, userAnswer),
			},
		},
	}

	resp, err := client.Chat(req)
	if err != nil {
		log.Printf("GigaChat CheckAnswer error: %v", err)
		return "❌ Ошибка проверки"
	}

	return resp.Choices[0].Message.Content
}
