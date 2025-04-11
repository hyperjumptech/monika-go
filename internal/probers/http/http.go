package http

import (
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	assertion "hyperjumptech/monika/internal/assertion"
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

// HttpResult represents the result of an HTTP request
type HttpResult struct {
	StatusCode   int
	ResponseTime float64
	Body         string
	Headers      map[string]string
	Size         int
}

// ProbeStatusReason represents the reason for a probe's status change
type ProbeStatusReason struct {
	AlertQuery   string
	AlertMessage string
	RequestURL   string
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

				var reason ProbeStatusReason
				failed := false

				// Make HTTP request to the URL
				for _, request := range probe.Requests {
					// Send the request
					resp, err := sendRequest(request, timeout)

					// If error, mark as failed
					if err != nil {
						reason = ProbeStatusReason{
							AlertQuery:   "error != nil",
							AlertMessage: err.Error(),
							RequestURL:   request.URL,
						}
						failed = true

						// Log error without trying to access resp fields
						logger.Info().Str("context", "probe").Str("type", "http").Msgf("%s - %s - %s - %s - Error: %s",
							probe.Name, probeHealth.Status, request.Method, request.URL, err.Error())

						break
					}

					logger.Info().Str("context", "probe").Str("type", "http").Msgf("%s - %s - %s - %s - %d - %.3fms", probe.Name, probeHealth.Status, request.Method, request.URL, resp.StatusCode, resp.ResponseTime)

					// If response time is greater than the timeout, mark as failed
					if resp.ResponseTime > float64(request.Timeout) {
						reason = ProbeStatusReason{
							AlertQuery:   fmt.Sprintf("response.time > %d", request.Timeout),
							AlertMessage: "Request timed out",
							RequestURL:   request.URL,
						}
						failed = true
						break
					}

					// Evaluate alert query expressions from the config file
					for _, alert := range request.Alerts {
						alertTriggered := assertion.Evaluate(alert.Query, map[string]interface{}{
							"response": map[string]interface{}{
								"status":  resp.StatusCode,
								"time":    resp.ResponseTime,
								"body":    resp.Body,
								"headers": resp.Headers,
								"size":    resp.Size,
							},
						})

						// Alert condition met, take action
						if alertTriggered {
							// Store the reason for the incident
							reason = ProbeStatusReason{
								AlertQuery:   alert.Query,
								AlertMessage: alert.Message,
								RequestURL:   request.URL,
							}

							failed = true
							break
						}
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

							// Construct a more detailed notification message
							notificationMsg := fmt.Sprintf(
								"Probe is now in an incident state\n\n"+
									"Probe: %s\n"+
									"Alert: %s\n"+
									"Message: %s\n"+
									"URL: %s",
								probe.Name,
								reason.AlertQuery,
								reason.AlertMessage,
								reason.RequestURL,
							)

							// Send notification to the configured channel(s)
							logger.Info().Str("context", "probe").Str("type", "http").Msgf("Probe %s is unhealthy, sending notification to the configured channel(s)", probe.Name)
							for _, notification := range config.Notifications {
								notifier.SendNotification(notification, notificationMsg)
							}
						} else {
							// Else, just log
							logger.Info().Str("context", "probe").Str("type", "http").Msgf("Alert detected for probe %s: %s. Attempt %d of %d until it may be considered an incident", probe.Name, reason.AlertMessage, probeHealth.IncidentCount, probeHealth.IncidentThreshold)
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

							// Construct a more detailed notification message
							notificationMsg := fmt.Sprintf(
								"Probe is now in a healthy state\n\n"+
									"Probe: %s\n"+
									"All checks passed successfully for %d consecutive attempts",
								probe.Name,
								probeHealth.RecoveryThreshold,
							)

							// Send notification to the configured channel(s)
							logger.Info().Str("context", "probe").Str("type", "http").Msgf("Probe %s is healthy, sending notification to the configured channel(s)", probe.Name)
							for _, notification := range config.Notifications {
								notifier.SendNotification(notification, notificationMsg)
							}
						} else {
							// Else, just log
							logger.Info().Str("context", "probe").Str("type", "http").Msgf("Alert has been resolved for probe %s: %s. Attempt %d of %d until it may be considered a recovery", probe.Name, reason.AlertMessage, probeHealth.RecoveryCount, probeHealth.RecoveryThreshold)
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

	// Read the response body
	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	bodyString := string(bodyBytes)

	// Convert headers from map[string][]string to map[string]string
	// For headers with multiple values, we'll join them with a comma
	headers := make(map[string]string)
	for key, values := range resp.Header {
		headers[key] = strings.Join(values, ", ")
	}

	elapsed := time.Since(start)
	return &HttpResult{
		StatusCode:   resp.StatusCode,
		ResponseTime: float64(elapsed.Milliseconds()),
		Body:         bodyString,
		Headers:      headers,
		Size:         len(bodyBytes), // More accurate than ContentLength which can be -1
	}, nil
}
