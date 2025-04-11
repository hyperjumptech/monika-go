package ping

import (
	"time"

	"hyperjumptech/monika/internal/loader"
	"hyperjumptech/monika/internal/logger"
	notifier "hyperjumptech/monika/internal/notification"

	probing "github.com/prometheus-community/pro-bing"
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

type PingResult struct {
	StatusCode   int
	ResponseTime float64
}

func CreateProbes(config *loader.Config) {
	logger := logger.GetLogger()

	for _, probe := range config.Probes {
		probeHealth := &ProbeHealth{
			Status:            HEALTHY,
			RecoveryThreshold: 5,
			IncidentThreshold: 5,
		}

		// Run probe using goroutine
		go func(probe loader.ConfigProbe, probeHealth *ProbeHealth) {
			interval := time.Duration(probe.Interval) * time.Second

			// Start the check for each request
			for {
				time.Sleep(interval)

				// Make ping request to the URL
				failed := false
				// Send the request
				resp, err := sendPing(probe.Socket)
				logger.Info().Str("context", "probe").Str("type", "ping").Msgf("%s - %s - %s - %d - %.3fms", probe.Name, probeHealth.Status, probe.Socket.Host, resp.StatusCode, resp.ResponseTime)

				if err != nil {
					// If error, mark as failed
					failed = true
					break
				} else if resp.ResponseTime > float64(10_000) {
					// If response time is greater than the 10 seconds timeout, mark as failed
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
							logger.Info().Str("context", "probe").Str("type", "ping").Msgf("Probe %s has failed, sending notification to the configured channel(s)", probe.Name)
							for _, notification := range config.Notifications {
								notifier.SendNotification(notification, "Probe "+probe.Name+" is now in an incident state")
							}
						} else {
							// Else, just log
							logger.Info().Str("context", "probe").Str("type", "ping").Msgf("Probe %s is failing, attempt %d of %d until it may be considered an incident", probe.Name, probeHealth.IncidentCount, probeHealth.IncidentThreshold)
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
							logger.Info().Str("context", "probe").Str("type", "ping").Msgf("Probe %s has recovered, sending notification to the configured channel(s)", probe.Name)
							for _, notification := range config.Notifications {
								notifier.SendNotification(notification, "Probe "+probe.Name+" is now in an healthy state")
							}
						} else {
							// Else, just log
							logger.Info().Str("context", "probe").Str("type", "ping").Msgf("Probe %s is recovering, attempt %d of %d until it may be considered a recovery", probe.Name, probeHealth.RecoveryCount, probeHealth.RecoveryThreshold)
						}
					}
				}
			}
		}(probe, probeHealth)
	}
}

func sendPing(socket loader.ConfigProbeSocket) (*PingResult, error) {
	// Create a new pinger
	pinger, err := probing.NewPinger(socket.Host)
	if err != nil {
		return nil, err
	}
	pinger.Count = 1
	pinger.Timeout = 10 * time.Second

	// Run ping
	err = pinger.Run()
	if err != nil {
		return nil, err
	}

	// Get ping statistics
	stats := pinger.Statistics()
	responseTime := float64(stats.MaxRtt.Milliseconds()) / 1_000

	return &PingResult{StatusCode: 200, ResponseTime: responseTime}, nil
}
