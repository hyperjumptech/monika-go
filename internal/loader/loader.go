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

// ConfigNotificationData holds the configuration for notifications
type ConfigNotificationData struct {
	// Discord notification
	URL string `yaml:"url,omitempty"`

	// SMTP notification
	Recipients []string `yaml:"recipients,omitempty"`
	Host       string   `yaml:"hostname,omitempty"`
	Port       int      `yaml:"port,omitempty"`
	Username   string   `yaml:"username,omitempty"`
	Password   string   `yaml:"password,omitempty"`
}

// GetDiscordConfig returns the Discord configuration
func (d *ConfigNotificationData) GetDiscordConfig() (string, bool) {
	if d.URL != "" {
		return d.URL, true
	}
	return "", false
}

// GetSMTPConfig returns the SMTP configuration if all required fields are present
func (d *ConfigNotificationData) GetSMTPConfig() (host string, port int, username, password string, recipients []string, ok bool) {
	if d.Host != "" && d.Port > 0 && len(d.Recipients) > 0 {
		return d.Host, d.Port, d.Username, d.Password, d.Recipients, true
	}
	return "", 0, "", "", nil, false
}

type ConfigNotification struct {
	ID   string                 `yaml:"id"`
	Type string                 `yaml:"type"`
	Data ConfigNotificationData `yaml:"data"`
}

type ConfigProbePing struct {
	Uri    string                    `yaml:"uri"`
	Alerts []ConfigProbeRequestAlert `yaml:"alerts"`
}

type ConfigProbeRequestAlert struct {
	Query   string `yaml:"query"`
	Message string `yaml:"message"`
}

type ConfigProbeRequest struct {
	Timeout           int16                     `yaml:"timeout"`
	Method            string                    `yaml:"method"`
	URL               string                    `yaml:"url"`
	RecoveryThreshold int                       `yaml:"recovery_threshold"`
	IncidentThreshold int                       `yaml:"incident_threshold"`
	Alerts            []ConfigProbeRequestAlert `yaml:"alerts"`
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

		if probe.Ping.Uri == "" {
			probePing = ConfigProbePing{}
		} else {
			probePing = probe.Ping
		}

		if probe.Ping.Uri != "" {
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
				var requestAlert []ConfigProbeRequestAlert

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

				// If incident threshold is not set, set it to 5 times
				if request.IncidentThreshold == 0 {
					requestIncidentThreshold = 5 // Default incident threshold, 3 times
				} else {
					requestIncidentThreshold = request.IncidentThreshold
				}

				if len(request.Alerts) == 0 {
					requestAlert = []ConfigProbeRequestAlert{
						{
							Query:   "response.status < 200 || response.status >= 300",
							Message: "Response status is not between 200 and 300",
						},
						{
							Query:   "response.time > 2000",
							Message: "Response time is greater than 2 seconds",
						},
					}
				} else {
					requestAlert = request.Alerts
				}

				// Assign values
				probeRequest := ConfigProbeRequest{
					URL:               requestURL,
					Timeout:           requestTimeout,
					Method:            requestMethod,
					RecoveryThreshold: requestRecoveryThreshold,
					IncidentThreshold: requestIncidentThreshold,
					Alerts:            requestAlert,
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
