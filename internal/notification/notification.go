package notification

import (
	"net/http"

	"hyperjumptech/monika/internal/loader"
	"hyperjumptech/monika/internal/logger"
	discord "hyperjumptech/monika/internal/notification/discord"
)

func SendNotification(notification loader.ConfigNotification, message string) {
	logger := logger.GetLogger()
	// Create a new HTTP Client
	client := &http.Client{}
	defer client.CloseIdleConnections()

	switch notification.Type {
	case "discord":
		discord.SendNotification(notification, discord.GeneratePayload(message))
	default:
		logger.Error().Msgf("Unsupported notification type: %s", notification.Type)
	}
}
