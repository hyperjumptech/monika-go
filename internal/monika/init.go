package monika

import (
	"flag"
	"hyperjumptech/monika/internal/loader"
	"hyperjumptech/monika/internal/logger"
	notifier "hyperjumptech/monika/internal/notification"
	"hyperjumptech/monika/internal/probers"
	"hyperjumptech/monika/tools"
	"os"
)

func Init() {
	// Initialize logger
	logger := logger.GetLogger()

	// Short flags definitions
	configShortFlag := flag.String("c", "", "Path to config file")

	// Long flags definitions
	configLongFlag := flag.String("config", "", "Path to config file")

	// Parse flags
	flag.Parse()

	// If config flag is set, use it
	var fileToRead string
	if *configShortFlag != "" || *configLongFlag != "" {
		if *configShortFlag != "" {
			fileToRead = *configShortFlag
		} else {
			fileToRead = *configLongFlag
		}
	} else {
		// If not, use default config file
		fileToRead = "monika.yml"
	}

	// Check whether the file exists
	if _, err := os.Stat(fileToRead); os.IsNotExist(err) {
		logger.Fatal().Msgf("Monika configuration file does not exists: %s", fileToRead)
		os.Exit(1)
	}

	// Read file contents
	contents, err := os.Open(fileToRead)
	if err != nil {
		logger.Fatal().Err(err).Msg("Failed to read Monika configuration file")
		os.Exit(1)
	}
	defer contents.Close()

	// Parse Monika configuration file as a struct
	conf, err := loader.LoadConfig(contents)
	if err != nil {
		logger.Fatal().Err(err).Msg("Failed to parse Monika configuration file")
		os.Exit(1)
	}

	// Send startup message
	for _, notification := range conf.Notifications {
		notifier.SendNotification(notification, "Monika is starting up")
	}

	// Run probing
	var geolocation *tools.GeolocationIP
	geolocation, _ = tools.GetGeolocationIP()
	if geolocation != nil {
		logger.Info().Msgf("Monika is running from %s, %s (%s - %s)", geolocation.City, geolocation.Country, geolocation.Isp, geolocation.Query)
	}

	logger.Info().Msgf("Running %d probes with %d notifications", len(conf.Probes), len(conf.Notifications))

	probers.InitializeProbes(conf)
}
