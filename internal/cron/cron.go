package cron

import (
	"hyperjumptech/monika/internal/loader"
	"hyperjumptech/monika/internal/logger"
	"time"

	ssl "hyperjumptech/monika/internal/cron/jobs"

	"github.com/go-co-op/gocron/v2"
)

func StartCron(cron gocron.Scheduler) {
	logger := logger.GetLogger()

	// Start cron
	cron.Start()

	// Job to check for SSL
	_, err := cron.NewJob(
		gocron.DurationJob(time.Duration(10)*time.Second),
		gocron.NewTask(ssl.Check, loader.GetConfig()),
	)
	if err != nil {
		logger.Warn().Err(err).Str("context", "cron").Str("type", "ssl").Msg("Failed to run SSL checker job")
	}
	logger.Info().Str("context", "cron").Str("type", "ssl").Msg("SSL checker job started at 10 seconds interval")
}
