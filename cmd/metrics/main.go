// packame main implements a web server that exposes the metrics coded in
// pkg/metrics.
//
// This program will look for the clouds.yaml entry corresponding to the
// environment variable OS_CLOUD, then expose the metrics on port 8080 on all
// interfaces.
package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gophercloud/utils/openstack/clientconfig"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"sigs.k8s.io/cluster-api-provider-openstack/pkg/metrics"
)

const addr = ":8080"

func main() {
	osCloud := os.Getenv("OS_CLOUD")
	if osCloud == "" {
		log.Fatal("OS_CLOUD not found in the environment. Exiting.")
	}

	cloud, err := clientconfig.GetCloudFromYAML(&clientconfig.ClientOpts{
		Cloud: osCloud,
	})
	if err != nil {
		log.Fatalf("parsing clouds.yaml: %v", err)
	}

	ctx := context.Background()

	services, err := metrics.New(ctx, *cloud, []byte(cloud.CACertFile), metrics.WithTTL(10*time.Second))
	if err != nil {
		log.Fatalf("creating the metrics: %v", err)
	}

	r := prometheus.NewRegistry()
	r.MustRegister(services...)

	h := promhttp.HandlerFor(r, promhttp.HandlerOpts{})

	log.Printf("Listening on %q", addr)
	http.Handle("/metrics", h)
	log.Fatal(http.ListenAndServe(addr, nil))
}
