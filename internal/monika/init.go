package monika

import (
	"flag"
	"hyperjumptech/monika/internal/loader"
	"hyperjumptech/monika/internal/logger"
	notifier "hyperjumptech/monika/internal/notification"
	"hyperjumptech/monika/internal/probers"
	"hyperjumptech/monika/tools"
	"os"
	"path/filepath"

	CRON "hyperjumptech/monika/internal/cron"

	"github.com/fsnotify/fsnotify"
	"github.com/go-co-op/gocron/v2"
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

	// Watch for changes in the config file
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		logger.Fatal().Str("context", "monika").Str("type", "init").Err(err).Msg("Failed to initialize file watcher")
		os.Exit(1)
	}

	// Use absolute path for more reliable watching
	absPath, err := filepath.Abs(fileToRead)
	if err != nil {
		logger.Fatal().Str("context", "monika").Str("type", "init").Err(err).Msgf("Failed to get absolute path for %s", fileToRead)
		os.Exit(1)
	}

	// Add the config file to the watcher
	err = watcher.Add(absPath)
	if err != nil {
		logger.Fatal().Str("context", "monika").Str("type", "init").Err(err).Msgf("Failed to watch file: %s", absPath)
		os.Exit(1)
	}

	// Start the goroutine to handle events
	go func() {
		for {
			select {
			case event, ok := <-watcher.Events:
				if !ok {
					logger.Warn().Str("context", "monika").Str("type", "watcher").Msg("Watcher event channel closed")
					return
				}

				if event.Has(fsnotify.Write) || event.Has(fsnotify.Create) {
					logger.Info().Str("context", "monika").Str("type", "watcher").
						Msgf("File %s has been modified, reloading configuration", event.Name)
					readConfig(fileToRead)
				}
			case err, ok := <-watcher.Errors:
				if !ok {
					logger.Warn().Str("context", "monika").Str("type", "watcher").Msg("Watcher error channel closed")
					return
				}
				logger.Error().Str("context", "monika").Str("type", "watcher").Err(err).Msg("Error watching file")
			}
		}
	}()

	// Read config for the first time
	readConfig(absPath)
}

func readConfig(configPath string) {
	logger := logger.GetLogger()
	// Check whether the file exists
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		logger.Fatal().Str("context", "monika").Str("type", "init").Msgf("Monika configuration file does not exists: %s", configPath)
		os.Exit(1)
	}

	// Read file contents
	contents, err := os.Open(configPath)
	if err != nil {
		logger.Fatal().Str("context", "monika").Str("type", "init").Err(err).Msg("Failed to read Monika configuration file")
		os.Exit(1)
	}
	defer contents.Close()

	// Parse Monika configuration file as a struct
	conf, err := loader.LoadConfig(contents)
	if err != nil {
		logger.Fatal().Str("context", "monika").Str("type", "init").Err(err).Msg("Failed to parse Monika configuration file")
		os.Exit(1)
	}

	// Send startup message
	logger.Info().Str("context", "monika").Str("type", "init").Msgf("Monika configuration loaded from %s", configPath)
	logger.Info().Str("context", "monika").Str("type", "init").Msgf("Running %d probes with %d notifications", len(conf.Probes), len(conf.Notifications))
	for _, notification := range conf.Notifications {
		notifier.SendNotification(notification, "Monika is starting up")
	}

	// Run probing
	go func() {
		var geolocation *tools.GeolocationIP
		geolocation, _ = tools.GetGeolocationIP()
		if geolocation != nil {
			logger.Info().Str("context", "monika").Str("type", "init").Msgf("Monika is running from %s, %s (%s - %s)", geolocation.City, geolocation.Country, geolocation.Isp, geolocation.Query)
		}
	}()

	// Initialize CRON jobs
	cron, err := gocron.NewScheduler()
	if err != nil {
		logger.Warn().Err(err).Str("context", "monika").Str("type", "init").Msg("Failed to initialize CRON scheduler, no CRON jobs will be executed.")
		return
	}
	go CRON.StartCron(cron)

	// Initialize probers
	probers.InitializeProbes(conf)
}
