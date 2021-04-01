package metrics

import (
	"bufio"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

type versionResults struct {
	returnValue string
	returnError error
}

func (s *versionResults) get() (string, error) {
	return s.returnValue, s.returnError
}

func newTestServiceVersion(metricName string) (serviceVersion, *versionResults) {
	res := new(versionResults)
	return NewServiceVersion(
		metricName,
		"This is not a helpful description.",
		res.get,
	), res
}

func getMetrics(t *testing.T, u, metricName string) []string {
	res, err := http.Get(u)
	if err != nil {
		t.Fatal(err)
	}
	defer res.Body.Close()

	var lines []string
	scanner := bufio.NewScanner(res.Body)
	for scanner.Scan() {
		if line := scanner.Text(); strings.HasPrefix(line, metricName) {
			lines = append(lines, line)
		}
	}
	if err := scanner.Err(); err != nil {
		t.Fatal(err)
	}

	return lines
}

func TestServiceVersionRefresh(t *testing.T) {
	const metricName = "my_example_metric"

	s, expectedVersion := newTestServiceVersion(metricName)

	r := prometheus.NewRegistry()
	r.MustRegister(s)

	ts := httptest.NewServer(promhttp.HandlerFor(r, promhttp.HandlerOpts{}))
	defer ts.Close()

	t.Run("returns the correct value", func(t *testing.T) {
		expectedVersion.returnValue = "1.99"
		s.Refresh()

		metrics := getMetrics(t, ts.URL, metricName)
		if have := len(metrics); have != 1 {
			t.Fatalf("expected one value for the metric, found %d", have)
		}

		want := fmt.Sprintf("%s{version=%q} %d", metricName, expectedVersion.returnValue, 1)
		if have := metrics[0]; have != want {
			t.Errorf("expected '%s', found %s", want, have)
		}
	})

	t.Run("correctly updates the value", func(t *testing.T) {
		expectedVersion.returnValue = "6.478392"
		s.Refresh()

		metrics := getMetrics(t, ts.URL, metricName)
		if have := len(metrics); have != 1 {
			t.Fatalf("expected one value for the metric, found %d", have)
		}

		want := fmt.Sprintf("%s{version=%q} %d", metricName, expectedVersion.returnValue, 1)
		if have := metrics[0]; have != want {
			t.Errorf("expected '%s', found %s", want, have)
		}
	})

	t.Run("outputs -1 in case of error", func(t *testing.T) {
		expectedVersion.returnError = bufio.ErrNegativeCount
		s.Refresh()

		metrics := getMetrics(t, ts.URL, metricName)
		if have := len(metrics); have != 1 {
			t.Fatalf("expected one value for the metric, found %d", have)
		}

		want := fmt.Sprintf("%s{version=%q} %d", metricName, "error", -1)
		if have := metrics[0]; have != want {
			t.Errorf("expected '%s', found %s", want, have)
		}
	})
}

type availabilityResults struct {
	returnValue bool
	returnError error
}

func (s *availabilityResults) get() (bool, error) {
	return s.returnValue, s.returnError
}

func newTestServiceAvailability(metricName string) (serviceAvailability, *availabilityResults) {
	res := new(availabilityResults)
	return NewServiceAvailability(
		metricName,
		"This is not a helpful description.",
		res.get,
	), res
}

func TestServiceAvailabilityRefresh(t *testing.T) {
	const metricName = "my_example_metric"

	s, expectedAvailability := newTestServiceAvailability(metricName)

	r := prometheus.NewRegistry()
	r.MustRegister(s)

	ts := httptest.NewServer(promhttp.HandlerFor(r, promhttp.HandlerOpts{}))
	defer ts.Close()

	t.Run("returns the correct value", func(t *testing.T) {
		expectedAvailability.returnValue = false
		s.Refresh()

		metrics := getMetrics(t, ts.URL, metricName)
		if have := len(metrics); have != 1 {
			t.Fatalf("expected one value for the metric, found %d", have)
		}

		want := fmt.Sprintf("%s %s", metricName, "0")
		if have := metrics[0]; have != want {
			t.Errorf("expected '%s', found %s", want, have)
		}
	})

	t.Run("correctly updates the value", func(t *testing.T) {
		expectedAvailability.returnValue = true
		s.Refresh()

		metrics := getMetrics(t, ts.URL, metricName)
		if have := len(metrics); have != 1 {
			t.Fatalf("expected one value for the metric, found %d", have)
		}

		want := fmt.Sprintf("%s %s", metricName, "1")
		if have := metrics[0]; have != want {
			t.Errorf("expected '%s', found %s", want, have)
		}
	})

	t.Run("outputs -1 in case of error", func(t *testing.T) {
		expectedAvailability.returnError = bufio.ErrNegativeCount
		s.Refresh()

		metrics := getMetrics(t, ts.URL, metricName)
		if have := len(metrics); have != 1 {
			t.Fatalf("expected one value for the metric, found %d", have)
		}

		want := fmt.Sprintf("%s %s", metricName, "-1")
		if have := metrics[0]; have != want {
			t.Errorf("expected '%s', found %s", want, have)
		}
	})
}
