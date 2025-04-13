package notification

import (
	"net/http"

	"hyperjumptech/monika/internal/loader"
	"hyperjumptech/monika/internal/logger"
	"hyperjumptech/monika/internal/notification/discord"
	"hyperjumptech/monika/internal/notification/smtp"
)

func SendNotification(notification loader.ConfigNotification, message string) {
	logger := logger.GetLogger()
	// Create a new HTTP Client
	client := &http.Client{}
	defer client.CloseIdleConnections()

	switch notification.Type {
	case "discord":
		discord.SendNotification(notification, discord.GeneratePayload(message))
	case "smtp":
		smtp.SendNotification(notification, smtp.GeneratePayload(message))
	default:
		logger.Error().Str("context", "notification").Str("type", "discord").Msgf("Unsupported notification type: %s", notification.Type)
	}
}
