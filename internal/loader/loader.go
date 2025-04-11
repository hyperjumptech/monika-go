package loader

import (
	"bufio"
	"errors"
	"io"
	"strconv"
	"strings"

	"github.com/goccy/go-yaml"
	"github.com/google/uuid"
)

type ConfigNotificationData struct {
	URL string `yaml:"url"`
}

type ConfigNotification struct {
	ID   string                 `yaml:"id"`
	Type string                 `yaml:"type"`
	Data ConfigNotificationData `yaml:"data"`
}

type ConfigProbePing struct {
	Uri string `yaml:"uri"`
}

type ConfigProbeRequest struct {
	Timeout           int16  `yaml:"timeout"`
	Method            string `yaml:"method"`
	URL               string `yaml:"url"`
	RecoveryThreshold int    `yaml:"recovery_threshold"`
	IncidentThreshold int    `yaml:"incident_threshold"`
}

type ConfigProbe struct {
	ID       string
	Name     string `yaml:"name"`
	Interval int8   `yaml:"interval"`
	Requests []ConfigProbeRequest
	Ping     ConfigProbePing
}

type Config struct {
	Probes        []ConfigProbe        `yaml:"probes"`
	Notifications []ConfigNotification `yaml:"notifications"`
}

var loadedConfig *Config

func LoadConfig(reader io.Reader) (*Config, error) {
	var contents []string

	// Initialize a scanner
	scanner := bufio.NewScanner(reader)

	// Add each line to the contents slice
	for scanner.Scan() {
		contents = append(contents, scanner.Text())
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}

	// Convert contents to a single string
	config := strings.Join(contents, "\n")

	// Parse the Monika configuration file
	var configYAML Config
	err := yaml.Unmarshal([]byte(config), &configYAML)
	if err != nil {
		return nil, err
	}

	// Create a parsed configuration
	configStruct := Config{
		Probes:        make([]ConfigProbe, 0),
		Notifications: make([]ConfigNotification, 0),
	}

	// Assign probes and probe requests
	for index, probe := range configYAML.Probes {
		// If probe ID is not set, generate a new one
		var probeID, probeName string
		var probeInterval int8
		var probeRequests []ConfigProbeRequest
		var probePing ConfigProbePing

		// If ID is not set, generate a new one
		if probe.ID == "" {
			probeID = uuid.NewString()
		} else {
			probeID = probe.ID
		}

		// If name is not set, set it to "Probe <index>"
		if probe.Name == "" {
			probeName = "Probe " + strconv.Itoa(index+1)
		} else {
			probeName = probe.Name
		}

		// If interval is not set, set it to 10 seconds
		if probe.Interval == 0 {
			probeInterval = 10 // Default interval, 10 seconds
		} else {
			probeInterval = probe.Interval
		}

		probeStruct := ConfigProbe{
			ID:       probeID,
			Name:     probeName,
			Interval: probeInterval,
			Requests: make([]ConfigProbeRequest, 0),
			Ping:     ConfigProbePing{},
		}

		if probe.Requests == nil {
			probeRequests = make([]ConfigProbeRequest, 0)
		} else {
			probeRequests = probe.Requests
		}

		if probe.Ping == (ConfigProbePing{}) {
			probePing = ConfigProbePing{}
		} else {
			probePing = probe.Ping
		}

		if probe.Ping != (ConfigProbePing{}) {
			// Handle socket mapping
			probeStruct.Ping = ConfigProbePing{
				Uri: probePing.Uri,
			}

			configStruct.Probes = append(configStruct.Probes, probeStruct)
		} else {
			// Handle request mapping
			for _, request := range probeRequests {
				var requestURL, requestMethod string
				var requestTimeout int16
				var requestRecoveryThreshold, requestIncidentThreshold int

				// Coalesce null values
				// If URL is not set, throw an
				if request.URL == "" {
					return nil, errors.New("Missing URL in probe request for probe ID: " + probeID)
				} else {
					requestURL = request.URL
				}

				// If method is not set, set it to GET
				if request.Method == "" {
					requestMethod = "GET"
				} else {
					requestMethod = request.Method
				}

				// If timeout is not set, set it to 10 seconds
				if request.Timeout == 0 {
					requestTimeout = 10_000 // Default timeout, 10 seconds
				} else {
					requestTimeout = request.Timeout
				}

				// If recovery threshold is not set, set it to 5 times
				if request.RecoveryThreshold == 0 {
					requestRecoveryThreshold = 5 // Default recovery threshold, 5 times
				} else {
					requestRecoveryThreshold = request.RecoveryThreshold
				}

				// If incident threshold is not set, set it to 3 times
				if request.IncidentThreshold == 0 {
					requestIncidentThreshold = 3 // Default incident threshold, 3 times
				} else {
					requestIncidentThreshold = request.IncidentThreshold
				}

				// Assign values
				probeRequest := ConfigProbeRequest{
					URL:               requestURL,
					Timeout:           requestTimeout,
					Method:            requestMethod,
					RecoveryThreshold: requestRecoveryThreshold,
					IncidentThreshold: requestIncidentThreshold,
				}
				probeStruct.Requests = append(probeStruct.Requests, probeRequest)
			}

			configStruct.Probes = append(configStruct.Probes, probeStruct)
		}
	}

	// Handle notifications
	for _, notification := range configYAML.Notifications {
		notificationStruct := ConfigNotification{
			ID:   notification.ID,
			Type: notification.Type,
			Data: notification.Data,
		}
		configStruct.Notifications = append(configStruct.Notifications, notificationStruct)
	}

	// Set the config
	loadedConfig = &configStruct

	return &configStruct, nil
}

func GetConfig() *Config {
	return loadedConfig
}
