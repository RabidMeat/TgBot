package main

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"
)

type GigaChatRequest struct {
	Model    string        `json:"model"`
	Messages []GigaMessage `json:"messages"`
	Stream   bool          `json:"stream"`
}

type GigaMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type GigaResponse struct {
	Choices []struct {
		Message GigaMessage `json:"message"`
	} `json:"choices"`
}

var gigaToken string
var tokenExpires time.Time

func init() {
	authData := os.Getenv("GIGACHAT_AUTH_DATA")
	if authData == "" {
		authData = "MDE5YzMyZjktNDU5Zi03OGMzLTlhYjMtNDM1ZTIwNGI4YWE0Ojk3OTQzMDE5LTU5MWEtNDJlOS04MGY0LTA4YjNhNTBiMThhMQ=="
	}

	token, err := getGigaToken(authData)
	if err == nil {
		gigaToken = token
		tokenExpires = time.Now().Add(50 * time.Minute) // токен живет 1 час
	}
}

func getGigaToken(authData string) (string, error) {
	url := "https://ngw.devices.sberbank.ru:9443/api/v2/oauth"

	// ✅ ИСПРАВЛЕНО: Content-Type для OAuth!
	data := strings.NewReader("scope=GIGACHAT_API_PERS")

	req, err := http.NewRequest("POST", url, data)
	if err != nil {
		return "", err
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Authorization", "Basic "+authData)
	req.Header.Set("RqUID", fmt.Sprintf("%d", time.Now().UnixNano()))

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	var tokenResp struct {
		AccessToken string `json:"access_token"`
	}
	json.Unmarshal(body, &tokenResp)

	return tokenResp.AccessToken, nil
}

func gigaChatRequest(prompt string) string {
	// ✅ Обновляем токен если просрочен
	if time.Now().After(tokenExpires) || gigaToken == "" {
		authData := os.Getenv("GIGACHAT_AUTH_DATA")
		if authData == "" {
			authData = "MDE5YzMyZjktNDU5Zi03OGMzLTlhYjMtNDM1ZTIwNGI4YWE0Ojk3OTQzMDE5LTU5MWEtNDJlOS04MGY0LTA4YjNhNTBiMThhMQ=="
		}
		token, err := getGigaToken(authData)
		if err != nil {
			return "❌ Ошибка авторизации GigaChat"
		}
		gigaToken = token
		tokenExpires = time.Now().Add(50 * time.Minute)
	}

	url := "https://gigachat.devices.sberbank.ru/api/v1/chat/completions"

	request := GigaChatRequest{
		Model: "GigaChat:latest",
		Messages: []GigaMessage{
			{Role: "user", Content: prompt},
		},
		Stream: false,
	}

	jsonData, _ := json.Marshal(request)

	req, _ := http.NewRequest("POST", url, bytes.NewReader(jsonData))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+gigaToken)
	req.Header.Set("RqUID", fmt.Sprintf("%d", time.Now().UnixNano()))

	// В gigaChatRequest(), замените client:
	client := &http.Client{
		Timeout: 30 * time.Second,
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: true, // ← ОТКЛЮЧАЕМ проверку!
			},
		},
	}

	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	// ✅ Логируем ответ для отладки
	fmt.Printf("GigaChat status: %d\n", resp.StatusCode)

	var gigaResp GigaResponse
	json.Unmarshal(body, &gigaResp)

	if len(gigaResp.Choices) > 0 {
		return gigaResp.Choices[0].Message.Content
	}

	// ✅ Показываем ошибку API
	var errorResp struct {
		Error struct {
			Message string `json:"message"`
		} `json:"error"`
	}
	json.Unmarshal(body, &errorResp)
	return fmt.Sprintf("❌ GigaChat API: %s (Status: %d)", errorResp.Error.Message, resp.StatusCode)
}

func GenerateTask(subject, topic, difficulty, classNum string) string {
	prompt := fmt.Sprintf(`Информатика %s класс. Тема "%s". Сложность "%s".
    
Создай задание:
1. Вопрос?
A) 
B) 
C) 
D) 

Только задание в этом формате! Без лишнего текста!`, classNum, topic, difficulty)

	return gigaChatRequest(prompt)
}

func CheckAnswer(task, userAnswer string) string {
	prompt := fmt.Sprintf(`ЗАДАНИЕ:
%s

ОТВЕТ: %s

Ответь КРАТКО: ✅ ПРАВИЛЬНО! Объяснение... ИЛИ ❌ НЕПРАВИЛЬНО. Правильный: X`, task, userAnswer)

	return gigaChatRequest(prompt)
}
