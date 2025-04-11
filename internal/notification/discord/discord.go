package discord

import (
	"bytes"
	"encoding/json"
	"hyperjumptech/monika/internal/loader"
	"hyperjumptech/monika/internal/logger"
	"net/http"
)

type Content struct {
	Content string `json:"content"`
}

func GeneratePayload(message string) Content {
	return Content{
		Content: message,
	}
}

func SendNotification(notification loader.ConfigNotification, message Content) {
	client := &http.Client{}
	send(client, notification, message)
}

func send(client *http.Client, notification loader.ConfigNotification, message Content) {
	logger := logger.GetLogger()

	// Create a new HTTP Client
	defer client.CloseIdleConnections()

	payload := new(bytes.Buffer)
	err := json.NewEncoder(payload).Encode(message)
	if err != nil {
		logger.Error().Err(err).Str("context", "notification").Str("type", "discord").Msg("Failed to encode Discord notification payload")
	}

	resp, err := http.Post(notification.Data.URL, "application/json", payload)
	if err != nil {
		logger.Error().Err(err).Str("context", "notification").Str("type", "discord").Msg("Failed to send Discord notification")
	}

	defer resp.Body.Close()
}
