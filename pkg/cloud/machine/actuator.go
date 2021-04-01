/*
Copyright 2018 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package machine

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/client-go/tools/record"
	"k8s.io/klog/v2"

	clconfig "github.com/coreos/container-linux-config-transpiler/config"
	"github.com/gophercloud/gophercloud/openstack/compute/v2/flavors"
	"github.com/gophercloud/gophercloud/openstack/compute/v2/servers"
	machinev1 "github.com/openshift/machine-api-operator/pkg/apis/machine/v1beta1"
	apierrors "github.com/openshift/machine-api-operator/pkg/controller/machine"
	"k8s.io/apimachinery/pkg/api/equality"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	tokenapi "k8s.io/cluster-bootstrap/token/api"
	tokenutil "k8s.io/cluster-bootstrap/token/util"
	openstackconfigv1 "sigs.k8s.io/cluster-api-provider-openstack/api/openstackproviderconfig/v1alpha1"
	infrav1 "sigs.k8s.io/cluster-api-provider-openstack/api/v1alpha4"
	"sigs.k8s.io/cluster-api-provider-openstack/pkg/cloud/services/compute"
	"sigs.k8s.io/cluster-api-provider-openstack/pkg/cloud/services/provider"
	"sigs.k8s.io/cluster-api/api/v1alpha4"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// Event Action Constants
const (
	createEventAction = "Create"
	updateEventAction = "Update"
	deleteEventAction = "Delete"
	noEventAction     = ""

	UserDataKey          = "userData"
	DisableTemplatingKey = "disableTemplating"
	PostprocessorKey     = "postprocessor"

	InstanceStatusAnnotationKey = "instance-status"

	OpenstackIdAnnotationKey = "openstack-resourceId"

	// MachineInstanceStateAnnotationName as annotation name for a machine instance state
	MachineInstanceStateAnnotationName = "machine.openshift.io/instance-state"

	// ErrorState is assigned to the machine if its instance has been destroyed
	ErrorState = "ERROR"

	// MachineRegionLabelName as annotation name for a machine region
	MachineRegionLabelName = "machine.openshift.io/region"

	// MachineAZLabelName as annotation name for a machine AZ
	MachineAZLabelName = "machine.openshift.io/zone"

	// MachineInstanceTypeLabelName as annotation name for a machine instance type
	MachineInstanceTypeLabelName = "machine.openshift.io/instance-type"
)

type OpenstackClient struct {
	params ActuatorParams
	scheme *runtime.Scheme
	client client.Client
	*DeploymentClient
	eventRecorder record.EventRecorder
}

func NewActuator(params ActuatorParams) (*OpenstackClient, error) {
	return &OpenstackClient{
		params:           params,
		client:           params.Client,
		scheme:           params.Scheme,
		DeploymentClient: NewDeploymentClient(),
		eventRecorder:    params.EventRecorder,
	}, nil
}

func (oc *OpenstackClient) Create(ctx context.Context, machine *machinev1.Machine) error {
	log := ctrl.LoggerFrom(ctx)

	// First check that provided labels are correct
	// TODO(mfedosin): stop sending the infrastructure request when we start to receive the cluster value
	clusterInfra, err := oc.params.ConfigClient.Infrastructures().Get(context.TODO(), "cluster", metav1.GetOptions{})
	if err != nil {
		return fmt.Errorf("Failed to retrieve cluster Infrastructure object: %v", err)
	}

	clusterInfraName := clusterInfra.Status.InfrastructureName
	clusterNameLabel := machine.Labels["machine.openshift.io/cluster-api-cluster"]

	if clusterNameLabel != clusterInfraName {
		klog.Errorf("machine.openshift.io/cluster-api-cluster label value is incorrect: %v, machine %v cannot join cluster %v", clusterNameLabel, machine.ObjectMeta.Name, clusterInfraName)
		verr := apierrors.InvalidMachineConfiguration("machine.openshift.io/cluster-api-cluster label value is incorrect: %v, machine %v cannot join cluster %v", clusterNameLabel, machine.ObjectMeta.Name, clusterInfraName)

		return oc.handleMachineError(machine, verr, createEventAction)
	}

	kubeClient := oc.params.KubeClient

	machineSpec, err := openstackconfigv1.MachineSpecFromProviderSpec(machine.Spec.ProviderSpec)
	if err != nil {
		return oc.handleMachineError(machine, apierrors.InvalidMachineConfiguration(
			"Cannot unmarshal providerSpec field: %v", err), createEventAction)
	}

	networks := make([]infrav1.NetworkParam, len(machineSpec.Networks))
	for _, n := range machineSpec.Networks {
		subnets := make([]infrav1.SubnetParam, len(n.Subnets))
		for _, s := range n.Subnets {
			subnets = append(subnets, infrav1.SubnetParam{
				UUID: s.UUID,
				Filter: infrav1.SubnetFilter{
					Name:            s.Filter.Name,
					Description:     s.Filter.Description,
					EnableDHCP:      s.Filter.EnableDHCP,
					NetworkID:       s.Filter.NetworkID,
					TenantID:        s.Filter.TenantID,
					ProjectID:       s.Filter.ProjectID,
					IPVersion:       s.Filter.IPVersion,
					GatewayIP:       s.Filter.GatewayIP,
					CIDR:            s.Filter.CIDR,
					IPv6AddressMode: s.Filter.IPv6AddressMode,
					IPv6RAMode:      s.Filter.IPv6RAMode,
					ID:              s.Filter.ID,
					SubnetPoolID:    s.Filter.SubnetPoolID,
					Limit:           s.Filter.Limit,
					Marker:          s.Filter.Marker,
					SortKey:         s.Filter.SortKey,
					SortDir:         s.Filter.SortDir,
					Tags:            s.Filter.Tags,
					TagsAny:         s.Filter.TagsAny,
					NotTags:         s.Filter.NotTags,
					NotTagsAny:      s.Filter.NotTagsAny,
				},
			})
		}
		networks = append(networks, infrav1.NetworkParam{
			UUID:    n.UUID,
			FixedIP: n.FixedIp,
			Filter: infrav1.Filter{
				Status:       n.Filter.Status,
				Name:         n.Filter.Name,
				Description:  n.Filter.Description,
				AdminStateUp: n.Filter.AdminStateUp,
				TenantID:     n.Filter.TenantID,
				ProjectID:    n.Filter.ProjectID,
				Shared:       n.Filter.Shared,
				ID:           n.Filter.ID,
				Marker:       n.Filter.Marker,
				Limit:        n.Filter.Limit,
				SortKey:      n.Filter.SortKey,
				SortDir:      n.Filter.SortDir,
				Tags:         n.Filter.Tags,
				TagsAny:      n.Filter.TagsAny,
				NotTags:      n.Filter.NotTags,
				NotTagsAny:   n.Filter.NotTagsAny,
			},
			Subnets: subnets,
		})

	}

	securityGroups := make([]infrav1.SecurityGroupParam, len(machineSpec.SecurityGroups))
	for _, sg := range machineSpec.SecurityGroups {
		securityGroups = append(securityGroups, infrav1.SecurityGroupParam{
			UUID: sg.UUID,
			Name: sg.Name,
			Filter: infrav1.SecurityGroupFilter{
				ID:          sg.Filter.ID,
				Name:        sg.Filter.Name,
				Description: sg.Filter.Description,
				TenantID:    sg.Filter.TenantID,
				ProjectID:   sg.Filter.ProjectID,
				Limit:       sg.Filter.Limit,
				Marker:      sg.Filter.Marker,
				SortKey:     sg.Filter.SortKey,
				SortDir:     sg.Filter.SortDir,
				Tags:        sg.Filter.Tags,
				TagsAny:     sg.Filter.TagsAny,
				NotTags:     sg.Filter.NotTags,
				NotTagsAny:  sg.Filter.NotTagsAny,
			},
		})
	}

	var rootVolume *infrav1.RootVolume
	if machineSpec.RootVolume != nil {
		rootVolume = &infrav1.RootVolume{
			SourceType: machineSpec.RootVolume.SourceType,
			SourceUUID: machineSpec.RootVolume.SourceUUID,
			DeviceType: machineSpec.RootVolume.DeviceType,
			Size:       machineSpec.RootVolume.Size,
		}
	}

	openStackMachine := &infrav1.OpenStackMachine{
		Spec: infrav1.OpenStackMachineSpec{
			ProviderID:     machine.Spec.ProviderID,
			CloudsSecret:   machineSpec.CloudsSecret,
			CloudName:      machineSpec.CloudName,
			Flavor:         machineSpec.Flavor,
			Image:          machineSpec.Image,
			SSHKeyName:     machineSpec.KeyName,
			Networks:       networks,
			Subnet:         machineSpec.PrimarySubnet,
			FloatingIP:     machineSpec.FloatingIP,
			SecurityGroups: securityGroups,
			UserDataSecret: machineSpec.UserDataSecret,
			Trunk:          machineSpec.Trunk,
			Tags:           machineSpec.Tags,
			ServerMetadata: machineSpec.ServerMetadata,
			ConfigDrive:    machineSpec.ConfigDrive,
			RootVolume:     rootVolume,
			ServerGroupID:  machineSpec.ServerGroupID,
		},
	}

	osProviderClient, clientOpts, err := provider.NewClientFromMachine(oc.client, openStackMachine)
	if err != nil {
		return err
	}

	computeService, err := compute.NewService(osProviderClient, clientOpts, log)
	if err != nil {
		return err
	}

	/*
		if err = oc.validateMachine(machine); err != nil {
			verr := apierrors.InvalidMachineConfiguration("Machine validation failed: %v", err)
			return oc.handleMachineError(machine, verr, createEventAction)
		}
	*/

	instance, err := computeService.InstanceExists(machine.Name)
	if err != nil {
		return err
	}
	if instance != nil {
		klog.Infof("Skipped creating a VM that already exists.\n")
		return nil
	}

	// Here we check whether we want to create a new instance or recreate the destroyed
	// one. If this is the second case, we have to return an error, because if we just
	// create an instance with the old name, because the CSR for it will not be approved
	// automatically.
	// See https://bugzilla.redhat.com/show_bug.cgi?id=1746369
	if machine.ObjectMeta.Annotations[InstanceStatusAnnotationKey] != "" {
		klog.Errorf("The instance has been destroyed for the machine %v, cannot recreate it.\n", machine.ObjectMeta.Name)
		verr := apierrors.InvalidMachineConfiguration("the instance has been destroyed for the machine %v, cannot recreate it.\n", machine.ObjectMeta.Name)

		return oc.handleMachineError(machine, verr, createEventAction)
	}

	// get machine startup script
	var ok bool
	var disableTemplating bool
	var postprocessor string
	var postprocess bool

	userData := []byte{}
	if machineSpec.UserDataSecret != nil {
		namespace := machineSpec.UserDataSecret.Namespace
		if namespace == "" {
			namespace = machine.Namespace
		}

		if machineSpec.UserDataSecret.Name == "" {
			return fmt.Errorf("UserDataSecret name must be provided")
		}

		userDataSecret, err := kubeClient.CoreV1().Secrets(namespace).Get(context.TODO(), machineSpec.UserDataSecret.Name, metav1.GetOptions{})
		if err != nil {
			return err
		}

		userData, ok = userDataSecret.Data[UserDataKey]
		if !ok {
			return fmt.Errorf("Machine's userdata secret %v in namespace %v did not contain key %v", machineSpec.UserDataSecret.Name, namespace, UserDataKey)
		}

		_, disableTemplating = userDataSecret.Data[DisableTemplatingKey]

		var p []byte
		p, postprocess = userDataSecret.Data[PostprocessorKey]

		postprocessor = string(p)
	}

	var userDataRendered string
	if len(userData) > 0 && !disableTemplating {
		// FIXME(mandre) Find the right way to check if machine is part of the control plane
		if machine.ObjectMeta.Name != "" {
			userDataRendered, err = masterStartupScript(machine, string(userData))
			if err != nil {
				return oc.handleMachineError(machine, apierrors.CreateMachine(
					"error creating Openstack instance: %v", err), createEventAction)
			}
		} else {
			klog.Info("Creating bootstrap token")
			token, err := oc.createBootstrapToken()
			if err != nil {
				return oc.handleMachineError(machine, apierrors.CreateMachine(
					"error creating Openstack instance: %v", err), createEventAction)
			}
			userDataRendered, err = nodeStartupScript(machine, token, string(userData))
			if err != nil {
				return oc.handleMachineError(machine, apierrors.CreateMachine(
					"error creating Openstack instance: %v", err), createEventAction)
			}
		}
	} else {
		userDataRendered = string(userData)
	}

	//Read the cluster name from the `machine`.
	clusterName := fmt.Sprintf("%s-%s", machine.Namespace, machine.Labels["machine.openshift.io/cluster-api-cluster"])

	if postprocess {
		switch postprocessor {
		// Postprocess with the Container Linux ct transpiler.
		case "ct":
			clcfg, ast, report := clconfig.Parse([]byte(userDataRendered))
			if len(report.Entries) > 0 {
				return fmt.Errorf("Postprocessor error: %s", report.String())
			}

			ignCfg, report := clconfig.Convert(clcfg, "openstack-metadata", ast)
			if len(report.Entries) > 0 {
				return fmt.Errorf("Postprocessor error: %s", report.String())
			}

			ud, err := json.Marshal(&ignCfg)
			if err != nil {
				return fmt.Errorf("Postprocessor error: %s", err)
			}

			userDataRendered = string(ud)

		default:
			return fmt.Errorf("Postprocessor error: unknown postprocessor: '%s'", postprocessor)
		}
	}

	instance, err = computeService.InstanceCreate(clusterName, &v1alpha4.Machine{}, openStackMachine, &infrav1.OpenStackCluster{}, string(userData))
	if err != nil {
		return oc.handleMachineError(machine, apierrors.CreateMachine(
			"error creating Openstack instance: %v", err), createEventAction)
	}

	if machineSpec.FloatingIP != "" {
		err := computeService.AssociateFloatingIP(instance.ID, machineSpec.FloatingIP)
		if err != nil {
			return oc.handleMachineError(machine, apierrors.CreateMachine(
				"Associate floatingIP err: %v", err), createEventAction)
		}

	}

	err = setMachineLabels(computeService, clientOpts.RegionName, machine, instance.ID)
	if err != nil {
		return nil
	}

	oc.eventRecorder.Eventf(machine, corev1.EventTypeNormal, "Created", "Created machine %v", machine.Name)
	return oc.updateAnnotation(machine, instance.ID, clusterInfraName)
}

func (oc *OpenstackClient) Delete(ctx context.Context, machine *machinev1.Machine) error {
	return nil
}

func (oc *OpenstackClient) Update(ctx context.Context, machine *machinev1.Machine) error {
	return nil
}

func (oc *OpenstackClient) Exists(ctx context.Context, machine *machinev1.Machine) (bool, error) {
	return false, nil
}

// If the OpenstackClient has a client for updating Machine objects, this will set
// the appropriate reason/message on the Machine.Status. If not, such as during
// cluster installation, it will operate as a no-op. It also returns the
// original error for convenience, so callers can do "return handleMachineError(...)".
func (oc *OpenstackClient) handleMachineError(machine *machinev1.Machine, err *apierrors.MachineError, eventAction string) error {
	if eventAction != noEventAction {
		oc.eventRecorder.Eventf(machine, corev1.EventTypeWarning, "Failed"+eventAction, "%v", err.Reason)
	}
	if oc.client != nil {
		reason := err.Reason
		message := err.Message
		machine.Status.ErrorReason = &reason
		machine.Status.ErrorMessage = &message

		// Set state label to indicate that this machine is broken
		if machine.ObjectMeta.Annotations == nil {
			machine.ObjectMeta.Annotations = make(map[string]string)
		}
		machine.ObjectMeta.Annotations[MachineInstanceStateAnnotationName] = ErrorState

		if err := oc.client.Update(context.TODO(), machine); err != nil {
			return fmt.Errorf("unable to update machine status: %v", err)
		}
	}

	klog.Errorf("Machine error %s: %v", machine.Name, err.Message)
	return err
}

/*
func (oc *OpenstackClient) validateMachine(machine *machinev1.Machine) error {
	machineSpec, err := openstackconfigv1.MachineSpecFromProviderSpec(machine.Spec.ProviderSpec)
	if err != nil {
		return fmt.Errorf("\nError getting the machine spec from the provider spec: %v", err)
	}

	machineService, err := clients.NewInstanceServiceFromMachine(oc.params.KubeClient, machine)
	if err != nil {
		return fmt.Errorf("\nError getting a new instance service from the machine: %v", err)
	}

	// TODO(mfedosin): add more validations here

	// Validate that image exists when not booting from volume
	if machineSpec.RootVolume == nil {
		err = machineService.DoesImageExist(machineSpec.Image)
		if err != nil {
			return err
		}
	}

	// Validate that flavor exists
	err = machineService.DoesFlavorExist(machineSpec.Flavor)
	if err != nil {
		return err
	}

	// Validate that Availability Zone exists
	err = machineService.DoesAvailabilityZoneExist(machineSpec.AvailabilityZone)
	if err != nil {
		return err
	}

	return nil
}
*/

// setMachineLabels set labels describing the machine
func setMachineLabels(computeClient *compute.Service, regionName string, machine *machinev1.Machine, instanceID string) error {
	if machine.Labels[MachineRegionLabelName] != "" && machine.Labels[MachineAZLabelName] != "" && machine.Labels[MachineInstanceTypeLabelName] != "" {
		return nil
	}

	var serverMetadata struct {
		// AZ contains name of the server's availability zone
		AZ string `json:"OS-EXT-AZ:availability_zone"`

		// Flavor refers to a JSON object, which itself indicates the hardware
		// configuration of the deployed server.
		Flavor map[string]interface{} `json:"flavor"`

		// Status contains the current operational status of the server,
		// such as IN_PROGRESS or ACTIVE.
		Status string `json:"status"`
	}

	compute := computeClient.HackGetComputeClient()
	err := servers.Get(compute, instanceID).ExtractInto(&serverMetadata)
	if err != nil {
		return err
	}

	if machine.Labels == nil {
		machine.Labels = make(map[string]string)
	}

	// Set the region
	machine.Labels[MachineRegionLabelName] = regionName

	// Set the availability zone
	machine.Labels[MachineAZLabelName] = serverMetadata.AZ

	// Set the flavor name
	flavor, err := flavors.Get(compute, serverMetadata.Flavor["id"].(string)).Extract()
	if err != nil {
		return err
	}
	machine.Labels[MachineInstanceTypeLabelName] = flavor.Name

	return nil
}

func (oc *OpenstackClient) createBootstrapToken() (string, error) {
	token, err := tokenutil.GenerateBootstrapToken()
	if err != nil {
		return "", err
	}

	// XXX(mdbooth): Expiration comes from a cmdline arg
	//expiration := time.Now().UTC().Add(options.TokenTTL)
	expiration := time.Now().UTC().Add(60 * time.Minute)

	tokenSecret, err := generateTokenSecret(token, expiration)
	if err != nil {
		panic(fmt.Sprintf("unable to create token. there might be a bug somwhere: %v", err))
	}

	err = oc.client.Create(context.TODO(), tokenSecret)
	if err != nil {
		return "", err
	}

	return tokenutil.TokenFromIDAndSecret(
		string(tokenSecret.Data[tokenapi.BootstrapTokenIDKey]),
		string(tokenSecret.Data[tokenapi.BootstrapTokenSecretKey]),
	), nil
}

func (oc *OpenstackClient) updateAnnotation(computeService *compute.Service, machine *machinev1.Machine, instanceID string, clusterInfraName string) error {
	statusCopy := *machine.Status.DeepCopy()

	if machine.ObjectMeta.Annotations == nil {
		machine.ObjectMeta.Annotations = make(map[string]string)
	}
	machine.ObjectMeta.Annotations[OpenstackIdAnnotationKey] = instanceID
	instance, _ := computeService.InstanceExists(machine.Name)
	mapAddr, err := getIPsFromInstance(instance)
	if err != nil {
		return err
	}

	primaryIP, err := oc.getPrimaryMachineIP(mapAddr, machine, clusterInfraName)
	if err != nil {
		return err
	}
	klog.Infof("Found the primary address for the machine %v: %v", machine.Name, primaryIP)

	machine.ObjectMeta.Annotations[OpenstackIPAnnotationKey] = primaryIP
	machine.ObjectMeta.Annotations[MachineInstanceStateAnnotationName] = string(instance.State)

	if err := oc.client.Update(context.TODO(), machine); err != nil {
		return err
	}

	networkAddresses := []corev1.NodeAddress{}
	networkAddresses = append(networkAddresses, corev1.NodeAddress{
		Type:    corev1.NodeInternalIP,
		Address: primaryIP,
	})

	networkAddresses = append(networkAddresses, corev1.NodeAddress{
		Type:    corev1.NodeHostName,
		Address: machine.Name,
	})

	networkAddresses = append(networkAddresses, corev1.NodeAddress{
		Type:    corev1.NodeInternalDNS,
		Address: machine.Name,
	})

	machineCopy := machine.DeepCopy()
	machineCopy.Status.Addresses = networkAddresses

	if !equality.Semantic.DeepEqual(machine.Status.Addresses, machineCopy.Status.Addresses) {
		if err := oc.client.Status().Update(context.TODO(), machineCopy); err != nil {
			return err
		}
	}

	machine.Status = statusCopy
	return oc.updateInstanceStatus(machine)
}
