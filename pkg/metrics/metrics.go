package metrics

import (
	"context"
	"time"

	"github.com/gophercloud/utils/openstack/clientconfig"
	"github.com/prometheus/client_golang/prometheus"
	"sigs.k8s.io/cluster-api-provider-openstack/pkg/cloud/openstack/clients"
)

type configuration struct {
	ttl time.Duration
}

func defaultOptions() configuration {
	return configuration{
		ttl: 23 * time.Hour,
	}
}

type option func(*configuration)

// WithTTL is a functional option to customize the TTL of the metrics data. Defaults to 23 hours.
//
// Use:
//
//     metrics.New(ctx, cloud, metrics.WithTTL(time.Minute))
func WithTTL(ttl time.Duration) option {
	return func(c *configuration) {
		c.ttl = ttl
	}
}

// func New returns the collectors to be added to a Prometheus Register.
//
// Exposed metrics:
//
//     openstack_service_compute {version}
//     openstack_service_baremetal
//     openstack_service_cloudformation
//     openstack_service_dns
//     openstack_service_loadbalancer
//     openstack_service_objectstore
//     openstack_service_orchestration
//     openstack_service_placement
//     openstack_service_sharedfilesystem
//
// `openstack_service_compute` exposes Nova's maximum microversion for ID 2.1.
// The other metrics expose 1 if the relative endpoint is listed in Keystone, 0
// otherwise.
func New(ctx context.Context, cloud clientconfig.Cloud, cert []byte, options ...option) ([]prometheus.Collector, error) {

	config := defaultOptions()
	for _, apply := range options {
		apply(&config)
	}

	provider, err := clients.GetProviderClient(cloud, cert)
	if err != nil {
		return nil, err
	}

	serviceCompute := computeMetric(provider, cloud.RegionName)
	serviceBaremetal := baremetalMetric(provider, cloud.RegionName)
	serviceLoadbalancer := loadbalancerMetric(provider, cloud.RegionName)
	serviceObjectstore := objectstoreMetric(provider, cloud.RegionName)
	serviceDNS := dnsMetric(provider, cloud.RegionName)
	serviceOrchestration := orchestrationMetric(provider, cloud.RegionName)
	servicePlacement := placementMetric(provider, cloud.RegionName)
	serviceSharedfilesystem := sharedfilesystemMetric(provider, cloud.RegionName)

	go keepFresh(
		ctx,
		config.ttl,
		serviceCompute,
		serviceBaremetal,
		serviceLoadbalancer,
		serviceObjectstore,
		serviceDNS,
		serviceOrchestration,
		servicePlacement,
		serviceSharedfilesystem,
	)

	return []prometheus.Collector{
		serviceCompute,
		serviceBaremetal,
		serviceLoadbalancer,
		serviceObjectstore,
		serviceDNS,
		serviceOrchestration,
		servicePlacement,
		serviceSharedfilesystem,
	}, nil
}
