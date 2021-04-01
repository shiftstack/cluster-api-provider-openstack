package metrics

import (
	"fmt"

	"github.com/gophercloud/gophercloud"
	"github.com/gophercloud/gophercloud/openstack"
	computeVersions "github.com/gophercloud/gophercloud/openstack/compute/apiversions"
)

const (
	novaVersionID            = "v2.1"
	ErrNovaVersionIDNotFound = Error("Nova version ID '" + novaVersionID + "' not found")
)

type Error string

func (e Error) Error() string {
	return string(e)
}

func computeMetric(pClient *gophercloud.ProviderClient, region string) serviceVersion {
	return NewServiceVersion(
		"openstack_service_compute",
		"OpenStack Compute service version availability",

		func() (string, error) {
			client, err := openstack.NewComputeV2(pClient, gophercloud.EndpointOpts{
				Region: region,
			})
			if err != nil {
				return "", fmt.Errorf("unable to create Compute serviceClient: %w", err)
			}

			allPages, err := computeVersions.List(client).AllPages()
			if err != nil {
				return "", fmt.Errorf("unable to get Compute API versions: %w", err)
			}

			allVersions, err := computeVersions.ExtractAPIVersions(allPages)
			if err != nil {
				return "", fmt.Errorf("unable to extract Compute API versions: %w", err)
			}

			for _, version := range allVersions {
				if version.ID == novaVersionID {
					return version.Version, nil
				}
			}
			return "", ErrNovaVersionIDNotFound
		},
	)
}

func baremetalMetric(pClient *gophercloud.ProviderClient, region string) serviceAvailability {
	return NewServiceAvailability(
		"openstack_service_baremetal",
		"OpenStack Baremetal service availability",
		func() (bool, error) {
			_, err := openstack.NewBareMetalV1(pClient, gophercloud.EndpointOpts{
				Region: region,
			})
			if err != nil {
				if _, ok := err.(*gophercloud.ErrEndpointNotFound); ok {
					return false, nil
				}
				return false, fmt.Errorf("unable to create Baremetal serviceClient: %w", err)
			}
			return true, nil
		},
	)
}

func loadbalancerMetric(pClient *gophercloud.ProviderClient, region string) serviceAvailability {
	return NewServiceAvailability(
		"openstack_service_loadbalancer",
		"OpenStack Load-balancer service availability",

		func() (bool, error) {
			_, err := openstack.NewLoadBalancerV2(pClient, gophercloud.EndpointOpts{
				Region: region,
			})
			if err != nil {
				if _, ok := err.(*gophercloud.ErrEndpointNotFound); ok {
					return false, nil
				}
				return false, fmt.Errorf("unable to create Load-balancer serviceClient: %w", err)
			}
			return true, nil
		},
	)
}

func objectstoreMetric(pClient *gophercloud.ProviderClient, region string) serviceAvailability {
	return NewServiceAvailability(
		"openstack_service_objectstore",
		"OpenStack Object store service availability",

		func() (bool, error) {
			_, err := openstack.NewObjectStorageV1(pClient, gophercloud.EndpointOpts{
				Region: region,
			})
			if err != nil {
				if _, ok := err.(*gophercloud.ErrEndpointNotFound); ok {
					return false, nil
				}
				return false, fmt.Errorf("unable to create Object storage serviceClient: %w", err)
			}
			return true, nil
		},
	)
}

func dnsMetric(pClient *gophercloud.ProviderClient, region string) serviceAvailability {
	return NewServiceAvailability(
		"openstack_service_dns",
		"OpenStack DNS service availability",

		func() (bool, error) {
			_, err := openstack.NewDNSV2(pClient, gophercloud.EndpointOpts{
				Region: region,
			})
			if err != nil {
				if _, ok := err.(*gophercloud.ErrEndpointNotFound); ok {
					return false, nil
				}
				return false, fmt.Errorf("unable to create DNS serviceClient: %w", err)
			}
			return true, nil
		},
	)
}

func orchestrationMetric(pClient *gophercloud.ProviderClient, region string) serviceAvailability {
	return NewServiceAvailability(
		"openstack_service_orchestration",
		"OpenStack Orchestration service availability",

		func() (bool, error) {
			_, err := openstack.NewOrchestrationV1(pClient, gophercloud.EndpointOpts{
				Region: region,
			})
			if err != nil {
				if _, ok := err.(*gophercloud.ErrEndpointNotFound); ok {
					return false, nil
				}
				return false, fmt.Errorf("unable to create Orchestration serviceClient: %w", err)
			}
			return true, nil
		},
	)
}

func placementMetric(pClient *gophercloud.ProviderClient, region string) serviceAvailability {
	return NewServiceAvailability(
		"openstack_service_placement",
		"OpenStack Placement service availability",

		func() (bool, error) {
			_, err := openstack.NewOrchestrationV1(pClient, gophercloud.EndpointOpts{
				Region: region,
			})
			if err != nil {
				if _, ok := err.(*gophercloud.ErrEndpointNotFound); ok {
					return false, nil
				}
				return false, fmt.Errorf("unable to create Orchestration serviceClient: %w", err)
			}
			return true, nil
		},
	)
}

func sharedfilesystemMetric(pClient *gophercloud.ProviderClient, region string) serviceAvailability {
	return NewServiceAvailability(
		"openstack_service_sharedfilesystem",
		"OpenStack Shared file system service availability",

		func() (bool, error) {
			_, err := openstack.NewSharedFileSystemV2(pClient, gophercloud.EndpointOpts{
				Region: region,
			})
			if err != nil {
				if _, ok := err.(*gophercloud.ErrEndpointNotFound); ok {
					return false, nil
				}
				return false, fmt.Errorf("unable to create Shared file system serviceClient: %w", err)
			}
			return true, nil
		},
	)
}
