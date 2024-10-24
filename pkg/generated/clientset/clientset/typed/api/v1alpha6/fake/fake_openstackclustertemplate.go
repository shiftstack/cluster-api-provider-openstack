/*
Copyright 2024 The Kubernetes Authors.

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

// Code generated by client-gen. DO NOT EDIT.

package fake

import (
	"context"
	json "encoding/json"
	"fmt"

	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	labels "k8s.io/apimachinery/pkg/labels"
	types "k8s.io/apimachinery/pkg/types"
	watch "k8s.io/apimachinery/pkg/watch"
	testing "k8s.io/client-go/testing"
	v1alpha6 "sigs.k8s.io/cluster-api-provider-openstack/api/v1alpha6"
	apiv1alpha6 "sigs.k8s.io/cluster-api-provider-openstack/pkg/generated/applyconfiguration/api/v1alpha6"
)

// FakeOpenStackClusterTemplates implements OpenStackClusterTemplateInterface
type FakeOpenStackClusterTemplates struct {
	Fake *FakeInfrastructureV1alpha6
	ns   string
}

var openstackclustertemplatesResource = v1alpha6.SchemeGroupVersion.WithResource("openstackclustertemplates")

var openstackclustertemplatesKind = v1alpha6.SchemeGroupVersion.WithKind("OpenStackClusterTemplate")

// Get takes name of the openStackClusterTemplate, and returns the corresponding openStackClusterTemplate object, and an error if there is any.
func (c *FakeOpenStackClusterTemplates) Get(ctx context.Context, name string, options v1.GetOptions) (result *v1alpha6.OpenStackClusterTemplate, err error) {
	emptyResult := &v1alpha6.OpenStackClusterTemplate{}
	obj, err := c.Fake.
		Invokes(testing.NewGetActionWithOptions(openstackclustertemplatesResource, c.ns, name, options), emptyResult)

	if obj == nil {
		return emptyResult, err
	}
	return obj.(*v1alpha6.OpenStackClusterTemplate), err
}

// List takes label and field selectors, and returns the list of OpenStackClusterTemplates that match those selectors.
func (c *FakeOpenStackClusterTemplates) List(ctx context.Context, opts v1.ListOptions) (result *v1alpha6.OpenStackClusterTemplateList, err error) {
	emptyResult := &v1alpha6.OpenStackClusterTemplateList{}
	obj, err := c.Fake.
		Invokes(testing.NewListActionWithOptions(openstackclustertemplatesResource, openstackclustertemplatesKind, c.ns, opts), emptyResult)

	if obj == nil {
		return emptyResult, err
	}

	label, _, _ := testing.ExtractFromListOptions(opts)
	if label == nil {
		label = labels.Everything()
	}
	list := &v1alpha6.OpenStackClusterTemplateList{ListMeta: obj.(*v1alpha6.OpenStackClusterTemplateList).ListMeta}
	for _, item := range obj.(*v1alpha6.OpenStackClusterTemplateList).Items {
		if label.Matches(labels.Set(item.Labels)) {
			list.Items = append(list.Items, item)
		}
	}
	return list, err
}

// Watch returns a watch.Interface that watches the requested openStackClusterTemplates.
func (c *FakeOpenStackClusterTemplates) Watch(ctx context.Context, opts v1.ListOptions) (watch.Interface, error) {
	return c.Fake.
		InvokesWatch(testing.NewWatchActionWithOptions(openstackclustertemplatesResource, c.ns, opts))

}

// Create takes the representation of a openStackClusterTemplate and creates it.  Returns the server's representation of the openStackClusterTemplate, and an error, if there is any.
func (c *FakeOpenStackClusterTemplates) Create(ctx context.Context, openStackClusterTemplate *v1alpha6.OpenStackClusterTemplate, opts v1.CreateOptions) (result *v1alpha6.OpenStackClusterTemplate, err error) {
	emptyResult := &v1alpha6.OpenStackClusterTemplate{}
	obj, err := c.Fake.
		Invokes(testing.NewCreateActionWithOptions(openstackclustertemplatesResource, c.ns, openStackClusterTemplate, opts), emptyResult)

	if obj == nil {
		return emptyResult, err
	}
	return obj.(*v1alpha6.OpenStackClusterTemplate), err
}

// Update takes the representation of a openStackClusterTemplate and updates it. Returns the server's representation of the openStackClusterTemplate, and an error, if there is any.
func (c *FakeOpenStackClusterTemplates) Update(ctx context.Context, openStackClusterTemplate *v1alpha6.OpenStackClusterTemplate, opts v1.UpdateOptions) (result *v1alpha6.OpenStackClusterTemplate, err error) {
	emptyResult := &v1alpha6.OpenStackClusterTemplate{}
	obj, err := c.Fake.
		Invokes(testing.NewUpdateActionWithOptions(openstackclustertemplatesResource, c.ns, openStackClusterTemplate, opts), emptyResult)

	if obj == nil {
		return emptyResult, err
	}
	return obj.(*v1alpha6.OpenStackClusterTemplate), err
}

// Delete takes name of the openStackClusterTemplate and deletes it. Returns an error if one occurs.
func (c *FakeOpenStackClusterTemplates) Delete(ctx context.Context, name string, opts v1.DeleteOptions) error {
	_, err := c.Fake.
		Invokes(testing.NewDeleteActionWithOptions(openstackclustertemplatesResource, c.ns, name, opts), &v1alpha6.OpenStackClusterTemplate{})

	return err
}

// DeleteCollection deletes a collection of objects.
func (c *FakeOpenStackClusterTemplates) DeleteCollection(ctx context.Context, opts v1.DeleteOptions, listOpts v1.ListOptions) error {
	action := testing.NewDeleteCollectionActionWithOptions(openstackclustertemplatesResource, c.ns, opts, listOpts)

	_, err := c.Fake.Invokes(action, &v1alpha6.OpenStackClusterTemplateList{})
	return err
}

// Patch applies the patch and returns the patched openStackClusterTemplate.
func (c *FakeOpenStackClusterTemplates) Patch(ctx context.Context, name string, pt types.PatchType, data []byte, opts v1.PatchOptions, subresources ...string) (result *v1alpha6.OpenStackClusterTemplate, err error) {
	emptyResult := &v1alpha6.OpenStackClusterTemplate{}
	obj, err := c.Fake.
		Invokes(testing.NewPatchSubresourceActionWithOptions(openstackclustertemplatesResource, c.ns, name, pt, data, opts, subresources...), emptyResult)

	if obj == nil {
		return emptyResult, err
	}
	return obj.(*v1alpha6.OpenStackClusterTemplate), err
}

// Apply takes the given apply declarative configuration, applies it and returns the applied openStackClusterTemplate.
func (c *FakeOpenStackClusterTemplates) Apply(ctx context.Context, openStackClusterTemplate *apiv1alpha6.OpenStackClusterTemplateApplyConfiguration, opts v1.ApplyOptions) (result *v1alpha6.OpenStackClusterTemplate, err error) {
	if openStackClusterTemplate == nil {
		return nil, fmt.Errorf("openStackClusterTemplate provided to Apply must not be nil")
	}
	data, err := json.Marshal(openStackClusterTemplate)
	if err != nil {
		return nil, err
	}
	name := openStackClusterTemplate.Name
	if name == nil {
		return nil, fmt.Errorf("openStackClusterTemplate.Name must be provided to Apply")
	}
	emptyResult := &v1alpha6.OpenStackClusterTemplate{}
	obj, err := c.Fake.
		Invokes(testing.NewPatchSubresourceActionWithOptions(openstackclustertemplatesResource, c.ns, *name, types.ApplyPatchType, data, opts.ToPatchOptions()), emptyResult)

	if obj == nil {
		return emptyResult, err
	}
	return obj.(*v1alpha6.OpenStackClusterTemplate), err
}
