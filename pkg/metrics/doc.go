// package metrics exposes Prometheus collections for OpenStack-specific
// metrics. This package is meant to be consumed by machine-api-operator.
//
// Metrics are gathered and cached for 23 hours by default.
//
// If gathering a metric generates an error, the corresponding metric value is
// set to "-1".
package metrics
