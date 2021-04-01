package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
)

type serviceVersion struct {
	*prometheus.GaugeVec
	Version func() (string, error)
}

// Refresh gathers fresh values and updates the Prometheus gauge. Sets the
// metric value to -1 in case gathering the metric produced a non-nil error.
func (s serviceVersion) Refresh() {
	s.Reset()

	version, err := s.Version()
	if err == nil {
		s.With(map[string]string{"version": version}).Set(1)
	} else {
		s.With(map[string]string{"version": "error"}).Set(-1)
	}
}

// NewServiceVersion constructs a type that embeds a gauge with labels and a
// function for getting fresh values.
// The returned value implements prometheus.Collector.
func NewServiceVersion(name, help string, versionsFunc func() (string, error)) serviceVersion {
	return serviceVersion{
		GaugeVec: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Name: name,
			Help: help,
		}, []string{"version"}),

		Version: versionsFunc,
	}
}

// serviceAvailability embeds a gauge and a method for getting fresh values.
type serviceAvailability struct {
	prometheus.Gauge
	HasEndpoint func() (bool, error)
}

// Refresh gathers fresh values and updates the Prometheus gauge. Sets the
// metric value to -1 in case gathering the metric produced a non-nil error.
func (s serviceAvailability) Refresh() {
	hasEndpoint, err := s.HasEndpoint()
	if err != nil {
		s.Set(-1)
		return
	}

	if hasEndpoint {
		s.Set(1)
	} else {
		s.Set(0)
	}
}

// NewServiceAvailability constructs a type that embeds a gauge and a function
// for getting fresh values.
// The returned value implements prometheus.Collector.
func NewServiceAvailability(name, help string, availabilityFunc func() (bool, error)) serviceAvailability {
	return serviceAvailability{
		Gauge: prometheus.NewGauge(prometheus.GaugeOpts{
			Name: name,
			Help: help,
		}),

		HasEndpoint: availabilityFunc,
	}
}
