package probers

import (
	"hyperjumptech/monika/internal/loader"
	HTTPProber "hyperjumptech/monika/internal/probers/http"
	PingProber "hyperjumptech/monika/internal/probers/ping"
)

func InitializeProbes(config *loader.Config) {
	// Create probes based on type
	HTTPProbes := make([]loader.ConfigProbe, 0)
	PingProbes := make([]loader.ConfigProbe, 0)

	// Filter probes based on type
	for _, probe := range config.Probes {
		if probe.Ping != (loader.ConfigProbePing{}) {
			PingProbes = append(PingProbes, probe)
		} else {
			HTTPProbes = append(HTTPProbes, probe)
		}
	}

	// Create new configuration to pass to probes
	HTTPConfig := loader.Config{
		Probes:        HTTPProbes,
		Notifications: config.Notifications,
	}
	HTTPProber.CreateProbes(&HTTPConfig)

	PingConfig := loader.Config{
		Probes:        PingProbes,
		Notifications: config.Notifications,
	}
	PingProber.CreateProbes(&PingConfig)
}
