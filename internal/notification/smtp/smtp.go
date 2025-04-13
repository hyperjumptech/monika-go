package smtp

import (
	"bytes"
	_ "embed"
	"html/template"
	"hyperjumptech/monika/internal/loader"
	"hyperjumptech/monika/internal/logger"
	"time"

	mail "github.com/xhit/go-simple-mail/v2"
)

type Content struct {
	Content struct {
		Body string
	}
}

//go:embed default-template.html
var defaultTemplate string

func GeneratePayload(message string) Content {
	return Content{
		Content: struct{ Body string }{
			Body: message,
		},
	}
}

func SendNotification(notification loader.ConfigNotification, message Content) {
	logger := logger.GetLogger()
	server := mail.NewSMTPClient()
	host, port, username, password, recipients, ok := notification.Data.GetSMTPConfig()
	if !ok {
		return
	}
	server.Host = host
	server.Port = port
	server.Username = username
	server.Password = password
	if port == 465 || port == 587 {
		server.Encryption = mail.EncryptionSTARTTLS
	}
	server.KeepAlive = false
	server.ConnectTimeout = 10 * time.Second
	server.SendTimeout = 10 * time.Second
	server.KeepAlive = false
	server.Authentication = mail.AuthLogin

	client, err := server.Connect()
	if err != nil {
		logger.Error().Err(err).Str("context", "notification").Str("type", "smtp").Msg("Failed to connect to SMTP server")
		return
	}
	defer client.Close()
	email := mail.NewMSG()
	email.SetFrom("Monika <" + username + ">")
	email.AddTo(recipients...)
	email.SetSubject("Monika Notification")
	templ, err := template.New("email-template").Parse(defaultTemplate)
	if err != nil {
		logger.Error().Err(err).Str("context", "notification").Str("type", "smtp").Msg("Failed to parse email template")
		return
	}
	var buffer bytes.Buffer
	if err := templ.Execute(&buffer, message.Content); err != nil {
		logger.Error().Err(err).Str("context", "notification").Str("type", "smtp").Msg("Failed to execute email template")
		return
	}
	body := buffer.String()
	email.SetBody(mail.TextHTML, body)
	if err := email.Send(client); err != nil {
		logger.Error().Err(err).Str("context", "notification").Str("type", "smtp").Msg("Failed to send SMTP notification")
		logger.Error().Err(err).Str("context", "notification").Str("type", "smtp").Msg(err.Error())
		return
	}
	logger.Info().Str("context", "notification").Str("type", "smtp").Msg("SMTP notification sent successfully")
	logger.Debug().Str("context", "notification").Str("type", "smtp").Msgf("SMTP notification sent to %s", recipients)
	logger.Debug().Str("context", "notification").Str("type", "smtp").Msgf("SMTP notification sent with message: %s", message.Content)
}
