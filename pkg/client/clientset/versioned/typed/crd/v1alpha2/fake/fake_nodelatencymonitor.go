// Copyright 2024 Antrea Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// Code generated by client-gen. DO NOT EDIT.

package fake

import (
	"context"

	v1alpha2 "antrea.io/antrea/pkg/apis/crd/v1alpha2"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	labels "k8s.io/apimachinery/pkg/labels"
	schema "k8s.io/apimachinery/pkg/runtime/schema"
	types "k8s.io/apimachinery/pkg/types"
	watch "k8s.io/apimachinery/pkg/watch"
	testing "k8s.io/client-go/testing"
)

// FakeNodeLatencyMonitors implements NodeLatencyMonitorInterface
type FakeNodeLatencyMonitors struct {
	Fake *FakeCrdV1alpha2
}

var nodelatencymonitorsResource = schema.GroupVersionResource{Group: "crd.antrea.io", Version: "v1alpha2", Resource: "nodelatencymonitors"}

var nodelatencymonitorsKind = schema.GroupVersionKind{Group: "crd.antrea.io", Version: "v1alpha2", Kind: "NodeLatencyMonitor"}

// Get takes name of the nodeLatencyMonitor, and returns the corresponding nodeLatencyMonitor object, and an error if there is any.
func (c *FakeNodeLatencyMonitors) Get(ctx context.Context, name string, options v1.GetOptions) (result *v1alpha2.NodeLatencyMonitor, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewRootGetAction(nodelatencymonitorsResource, name), &v1alpha2.NodeLatencyMonitor{})
	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha2.NodeLatencyMonitor), err
}

// List takes label and field selectors, and returns the list of NodeLatencyMonitors that match those selectors.
func (c *FakeNodeLatencyMonitors) List(ctx context.Context, opts v1.ListOptions) (result *v1alpha2.NodeLatencyMonitorList, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewRootListAction(nodelatencymonitorsResource, nodelatencymonitorsKind, opts), &v1alpha2.NodeLatencyMonitorList{})
	if obj == nil {
		return nil, err
	}

	label, _, _ := testing.ExtractFromListOptions(opts)
	if label == nil {
		label = labels.Everything()
	}
	list := &v1alpha2.NodeLatencyMonitorList{ListMeta: obj.(*v1alpha2.NodeLatencyMonitorList).ListMeta}
	for _, item := range obj.(*v1alpha2.NodeLatencyMonitorList).Items {
		if label.Matches(labels.Set(item.Labels)) {
			list.Items = append(list.Items, item)
		}
	}
	return list, err
}

// Watch returns a watch.Interface that watches the requested nodeLatencyMonitors.
func (c *FakeNodeLatencyMonitors) Watch(ctx context.Context, opts v1.ListOptions) (watch.Interface, error) {
	return c.Fake.
		InvokesWatch(testing.NewRootWatchAction(nodelatencymonitorsResource, opts))
}

// Create takes the representation of a nodeLatencyMonitor and creates it.  Returns the server's representation of the nodeLatencyMonitor, and an error, if there is any.
func (c *FakeNodeLatencyMonitors) Create(ctx context.Context, nodeLatencyMonitor *v1alpha2.NodeLatencyMonitor, opts v1.CreateOptions) (result *v1alpha2.NodeLatencyMonitor, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewRootCreateAction(nodelatencymonitorsResource, nodeLatencyMonitor), &v1alpha2.NodeLatencyMonitor{})
	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha2.NodeLatencyMonitor), err
}

// Update takes the representation of a nodeLatencyMonitor and updates it. Returns the server's representation of the nodeLatencyMonitor, and an error, if there is any.
func (c *FakeNodeLatencyMonitors) Update(ctx context.Context, nodeLatencyMonitor *v1alpha2.NodeLatencyMonitor, opts v1.UpdateOptions) (result *v1alpha2.NodeLatencyMonitor, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewRootUpdateAction(nodelatencymonitorsResource, nodeLatencyMonitor), &v1alpha2.NodeLatencyMonitor{})
	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha2.NodeLatencyMonitor), err
}

// Delete takes name of the nodeLatencyMonitor and deletes it. Returns an error if one occurs.
func (c *FakeNodeLatencyMonitors) Delete(ctx context.Context, name string, opts v1.DeleteOptions) error {
	_, err := c.Fake.
		Invokes(testing.NewRootDeleteActionWithOptions(nodelatencymonitorsResource, name, opts), &v1alpha2.NodeLatencyMonitor{})
	return err
}

// DeleteCollection deletes a collection of objects.
func (c *FakeNodeLatencyMonitors) DeleteCollection(ctx context.Context, opts v1.DeleteOptions, listOpts v1.ListOptions) error {
	action := testing.NewRootDeleteCollectionAction(nodelatencymonitorsResource, listOpts)

	_, err := c.Fake.Invokes(action, &v1alpha2.NodeLatencyMonitorList{})
	return err
}

// Patch applies the patch and returns the patched nodeLatencyMonitor.
func (c *FakeNodeLatencyMonitors) Patch(ctx context.Context, name string, pt types.PatchType, data []byte, opts v1.PatchOptions, subresources ...string) (result *v1alpha2.NodeLatencyMonitor, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewRootPatchSubresourceAction(nodelatencymonitorsResource, name, pt, data, subresources...), &v1alpha2.NodeLatencyMonitor{})
	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha2.NodeLatencyMonitor), err
}
