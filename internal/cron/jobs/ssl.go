// In ssl/check.go
package ssl

import (
	"crypto/tls"
	"fmt"
	"hyperjumptech/monika/internal/loader"
	"hyperjumptech/monika/internal/logger"
	"hyperjumptech/monika/internal/notification"
	"net"
	"net/url"
	"time"
)

func Check(conf *loader.Config) {
	logger := logger.GetLogger()

	// Check if config is loaded
	// If there's no config, skip the job
	if conf == nil || conf.Probes == nil {
		logger.Info().Str("context", "cron").Str("type", "ssl").Msg("No configuration found. Skipping...")
		return
	}

	// Filter probes based on type
	// We only want to check SSL of HTTPS probes
	HTTProbes := make([]loader.ConfigProbe, 0)
	for _, probe := range conf.Probes {
		if probe.Ping.Uri == "" {
			HTTProbes = append(HTTProbes, probe)
		}

		continue
	}

	// If no HTTP probes found, skip the job
	if len(HTTProbes) == 0 {
		logger.Info().Str("context", "cron").Str("type", "ssl").Msg("No HTTP probes found. Skipping...")
		return
	}

	// Check SSL of all HTTP probes
	for _, probe := range HTTProbes {
		for _, request := range probe.Requests {
			parsedURL, err := url.Parse(request.URL)
			if err != nil {
				logger.Warn().Err(err).Str("context", "cron").Str("type", "ssl").Msgf("Failed to parse URL %s. Skipping...", request.URL)
				continue
			}

			if parsedURL.Scheme != "https" {
				logger.Info().Str("context", "cron").Str("type", "ssl").Msgf("URL %s is not HTTPS. Skipping...", request.URL)
				continue
			}

			hostname := parsedURL.Hostname()
			if hostname == "" {
				logger.Warn().Str("context", "cron").Str("type", "ssl").Msgf("Failed to get hostname from URL %s. Skipping...", request.URL)
				continue
			}

			// Default to port 443 if not specified
			port := parsedURL.Port()
			if port == "" {
				port = "443"
			}

			serverAddr := fmt.Sprintf("%s:%s", hostname, port)
			logger.Info().Str("context", "cron").Str("type", "ssl").Msgf("Checking SSL of probe %s at %s", probe.Name, serverAddr)

			// Set up proper TLS config
			tlsConfig := &tls.Config{
				ServerName: hostname,
				MinVersion: tls.VersionTLS12,
			}

			// Connect with timeout
			dialer := &net.Dialer{
				Timeout: 10 * time.Second,
			}

			conn, err := tls.DialWithDialer(dialer, "tcp", serverAddr, tlsConfig)
			if err != nil {
				logger.Warn().Err(err).Str("context", "cron").Str("type", "ssl").Msgf("Failed to establish SSL connection to %s", serverAddr)
				continue
			}
			defer conn.Close()

			// Certificate validation
			certs := conn.ConnectionState().PeerCertificates
			if len(certs) == 0 {
				logger.Warn().Str("context", "cron").Str("type", "ssl").Msgf("No certificates found for %s", serverAddr)
				continue
			}

			cert := certs[0]

			// Check expiry
			now := time.Now()
			expiresIn := cert.NotAfter.Sub(now)

			if now.After(cert.NotAfter) {
				logger.Warn().
					Str("context", "cron").
					Str("type", "ssl").
					Msgf("SSL certificate for %s is expired, expired at %s", hostname, cert.NotAfter)

				// Send notification
				for _, n := range conf.Notifications {
					message := fmt.Sprintf("SSL certificate for %s is expired, expired at %s", hostname, cert.NotAfter)
					notification.SendNotification(n, message)
				}
			} else if expiresIn == 30*24*time.Hour || expiresIn < 14*24*time.Hour || expiresIn < 7*24*time.Hour {
				// Warn if certificate expires in equal to 30 days, equal to 14 days or equal to 7 days
				logger.Warn().
					Str("context", "cron").
					Str("type", "ssl").
					Msgf("SSL certificate for %s expires soon, expired at %s", hostname, cert.NotAfter)

				// Send notification
				for _, n := range conf.Notifications {
					message := fmt.Sprintf("SSL certificate for %s expires soon, expired at %s", hostname, cert.NotAfter)
					notification.SendNotification(n, message)
				}
			} else {
				logger.Info().
					Str("context", "cron").
					Str("type", "ssl").
					Msgf("SSL certificate for %s is valid", hostname)
			}
		}
	}
}
