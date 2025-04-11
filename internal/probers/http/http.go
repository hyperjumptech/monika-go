package http

import (
	"net/http"
	"strings"
	"time"

	"hyperjumptech/monika/internal/loader"
	"hyperjumptech/monika/internal/logger"
	notifier "hyperjumptech/monika/internal/notification"
)

// ProbeStatus represents the status of a probe
type ProbeStatus string

const (
	HEALTHY  ProbeStatus = "Healthy"
	INCIDENT ProbeStatus = "Incident"
)

// Probehealth holds the current status of a probe and its thresholds
type ProbeHealth struct {
	Status            ProbeStatus
	IncidentCount     int
	RecoveryCount     int
	RecoveryThreshold int
	IncidentThreshold int
}

type HttpResult struct {
	StatusCode   int
	ResponseTime float64
}

func CreateProbes(config *loader.Config) {
	logger := logger.GetLogger()

	for _, probe := range config.Probes {
		// Determine the largest recovery and incident thresholds
		recoveryThreshold := probe.Requests[0].RecoveryThreshold
		incidentThreshold := probe.Requests[0].IncidentThreshold

		for _, request := range probe.Requests {
			if request.RecoveryThreshold > recoveryThreshold {
				recoveryThreshold = request.RecoveryThreshold
			}

			if request.IncidentThreshold > incidentThreshold {
				incidentThreshold = request.IncidentThreshold
			}
		}

		probeHealth := &ProbeHealth{
			Status:            HEALTHY,
			RecoveryThreshold: recoveryThreshold,
			IncidentThreshold: incidentThreshold,
		}

		// Run probe using goroutine
		go func(probe loader.ConfigProbe, probeHealth *ProbeHealth) {
			interval := time.Duration(probe.Interval) * time.Second
			timeout := time.Duration(probe.Requests[0].Timeout) * time.Millisecond

			// Start the check for each request
			for {
				time.Sleep(interval)

				// Make HTTP request to the URL
				failed := false
				for _, request := range probe.Requests {
					// Send the request
					resp, err := sendRequest(request, timeout)
					logger.Info().Str("context", "probe").Str("type", "http").Msgf("%s - %s - %s - %s - %d - %.3fms", probe.Name, probeHealth.Status, request.Method, request.URL, resp.StatusCode, resp.ResponseTime)

					if err != nil {
						// If error, mark as failed
						failed = true
						break
					} else if resp.StatusCode < 200 || resp.StatusCode >= 300 {
						// If status code is not between 200 and 300, mark as failed
						failed = true
						break
					} else if resp.ResponseTime > float64(request.Timeout) {
						// If response time is greater than the timeout, mark as failed
						failed = true
						break
					} else if resp.ResponseTime > 2000 {
						// If response time is greater than 2 seconds, mark as failed
						failed = true
						break
					} else {
						// Otherwise, mark as successful
						failed = false
					}
				}

				// Handle requests results
				if failed {
					// If any request failed, handle incident detection
					// Add the incident count and reset the recovery count
					probeHealth.IncidentCount++
					probeHealth.RecoveryCount = 0

					// If the probe is in healthy state, check if it has reached the incident threshold
					if probeHealth.Status == HEALTHY {
						// If the incident count is greater than or equal to the incident threshold, mark the probe as incident
						if probeHealth.IncidentCount >= probeHealth.IncidentThreshold {
							probeHealth.Status = INCIDENT

							// Send notification to the configured channel(s)
							logger.Info().Str("context", "probe").Str("type", "http").Msgf("Probe %s has failed, sending notification to the configured channel(s)", probe.Name)
							for _, notification := range config.Notifications {
								notifier.SendNotification(notification, "Probe "+probe.Name+" is now in an incident state")
							}
						} else {
							// Else, just log
							logger.Info().Str("context", "probe").Str("type", "http").Msgf("Probe %s is failing, attempt %d of %d until it may be considered an incident", probe.Name, probeHealth.IncidentCount, probeHealth.IncidentThreshold)
						}
					}
				} else {
					// If all request were successful, handle recovery detection
					// Add the recovery count and reset the incident count
					probeHealth.RecoveryCount++
					probeHealth.IncidentCount = 0

					// If the probe is in incident state, check if it has reached the recovery threshold
					if probeHealth.Status == INCIDENT {
						// If the incident count is greater than or equal to the recover threshold, mark the probe as healthy
						if probeHealth.RecoveryCount >= probeHealth.RecoveryThreshold {
							probeHealth.Status = HEALTHY

							// Send notification to the configured channel(s)
							logger.Info().Str("context", "probe").Str("type", "http").Msgf("Probe %s has recovered, sending notification to the configured channel(s)", probe.Name)
							for _, notification := range config.Notifications {
								notifier.SendNotification(notification, "Probe "+probe.Name+" is now in an healthy state")
							}
						} else {
							// Else, just log
							logger.Info().Str("context", "probe").Str("type", "http").Msgf("Probe %s is recovering, attempt %d of %d until it may be considered a recovery", probe.Name, probeHealth.RecoveryCount, probeHealth.RecoveryThreshold)
						}
					}
				}
			}
		}(probe, probeHealth)
	}
}

func sendRequest(request loader.ConfigProbeRequest, timeout time.Duration) (*HttpResult, error) {
	// Create a new HTTP Client
	client := &http.Client{
		Timeout: time.Duration(timeout) * time.Millisecond,
	}
	defer client.CloseIdleConnections()

	// Create a new HTTP request
	start := time.Now()
	method := strings.ToUpper(request.Method)
	req, err := http.NewRequest(method, request.URL, nil)
	if err != nil {
		return nil, err
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}

	elapsed := time.Since(start)
	return &HttpResult{
		StatusCode:   resp.StatusCode,
		ResponseTime: float64(elapsed.Milliseconds()) / 1_000,
	}, nil
}
