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

package v1alpha2

import (
	"context"
	"time"

	v1alpha2 "antrea.io/antrea/pkg/apis/crd/v1alpha2"
	scheme "antrea.io/antrea/pkg/client/clientset/versioned/scheme"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	types "k8s.io/apimachinery/pkg/types"
	watch "k8s.io/apimachinery/pkg/watch"
	rest "k8s.io/client-go/rest"
)

// NodeLatencyMonitorsGetter has a method to return a NodeLatencyMonitorInterface.
// A group's client should implement this interface.
type NodeLatencyMonitorsGetter interface {
	NodeLatencyMonitors() NodeLatencyMonitorInterface
}

// NodeLatencyMonitorInterface has methods to work with NodeLatencyMonitor resources.
type NodeLatencyMonitorInterface interface {
	Create(ctx context.Context, nodeLatencyMonitor *v1alpha2.NodeLatencyMonitor, opts v1.CreateOptions) (*v1alpha2.NodeLatencyMonitor, error)
	Update(ctx context.Context, nodeLatencyMonitor *v1alpha2.NodeLatencyMonitor, opts v1.UpdateOptions) (*v1alpha2.NodeLatencyMonitor, error)
	Delete(ctx context.Context, name string, opts v1.DeleteOptions) error
	DeleteCollection(ctx context.Context, opts v1.DeleteOptions, listOpts v1.ListOptions) error
	Get(ctx context.Context, name string, opts v1.GetOptions) (*v1alpha2.NodeLatencyMonitor, error)
	List(ctx context.Context, opts v1.ListOptions) (*v1alpha2.NodeLatencyMonitorList, error)
	Watch(ctx context.Context, opts v1.ListOptions) (watch.Interface, error)
	Patch(ctx context.Context, name string, pt types.PatchType, data []byte, opts v1.PatchOptions, subresources ...string) (result *v1alpha2.NodeLatencyMonitor, err error)
	NodeLatencyMonitorExpansion
}

// nodeLatencyMonitors implements NodeLatencyMonitorInterface
type nodeLatencyMonitors struct {
	client rest.Interface
}

// newNodeLatencyMonitors returns a NodeLatencyMonitors
func newNodeLatencyMonitors(c *CrdV1alpha2Client) *nodeLatencyMonitors {
	return &nodeLatencyMonitors{
		client: c.RESTClient(),
	}
}

// Get takes name of the nodeLatencyMonitor, and returns the corresponding nodeLatencyMonitor object, and an error if there is any.
func (c *nodeLatencyMonitors) Get(ctx context.Context, name string, options v1.GetOptions) (result *v1alpha2.NodeLatencyMonitor, err error) {
	result = &v1alpha2.NodeLatencyMonitor{}
	err = c.client.Get().
		Resource("nodelatencymonitors").
		Name(name).
		VersionedParams(&options, scheme.ParameterCodec).
		Do(ctx).
		Into(result)
	return
}

// List takes label and field selectors, and returns the list of NodeLatencyMonitors that match those selectors.
func (c *nodeLatencyMonitors) List(ctx context.Context, opts v1.ListOptions) (result *v1alpha2.NodeLatencyMonitorList, err error) {
	var timeout time.Duration
	if opts.TimeoutSeconds != nil {
		timeout = time.Duration(*opts.TimeoutSeconds) * time.Second
	}
	result = &v1alpha2.NodeLatencyMonitorList{}
	err = c.client.Get().
		Resource("nodelatencymonitors").
		VersionedParams(&opts, scheme.ParameterCodec).
		Timeout(timeout).
		Do(ctx).
		Into(result)
	return
}

// Watch returns a watch.Interface that watches the requested nodeLatencyMonitors.
func (c *nodeLatencyMonitors) Watch(ctx context.Context, opts v1.ListOptions) (watch.Interface, error) {
	var timeout time.Duration
	if opts.TimeoutSeconds != nil {
		timeout = time.Duration(*opts.TimeoutSeconds) * time.Second
	}
	opts.Watch = true
	return c.client.Get().
		Resource("nodelatencymonitors").
		VersionedParams(&opts, scheme.ParameterCodec).
		Timeout(timeout).
		Watch(ctx)
}

// Create takes the representation of a nodeLatencyMonitor and creates it.  Returns the server's representation of the nodeLatencyMonitor, and an error, if there is any.
func (c *nodeLatencyMonitors) Create(ctx context.Context, nodeLatencyMonitor *v1alpha2.NodeLatencyMonitor, opts v1.CreateOptions) (result *v1alpha2.NodeLatencyMonitor, err error) {
	result = &v1alpha2.NodeLatencyMonitor{}
	err = c.client.Post().
		Resource("nodelatencymonitors").
		VersionedParams(&opts, scheme.ParameterCodec).
		Body(nodeLatencyMonitor).
		Do(ctx).
		Into(result)
	return
}

// Update takes the representation of a nodeLatencyMonitor and updates it. Returns the server's representation of the nodeLatencyMonitor, and an error, if there is any.
func (c *nodeLatencyMonitors) Update(ctx context.Context, nodeLatencyMonitor *v1alpha2.NodeLatencyMonitor, opts v1.UpdateOptions) (result *v1alpha2.NodeLatencyMonitor, err error) {
	result = &v1alpha2.NodeLatencyMonitor{}
	err = c.client.Put().
		Resource("nodelatencymonitors").
		Name(nodeLatencyMonitor.Name).
		VersionedParams(&opts, scheme.ParameterCodec).
		Body(nodeLatencyMonitor).
		Do(ctx).
		Into(result)
	return
}

// Delete takes name of the nodeLatencyMonitor and deletes it. Returns an error if one occurs.
func (c *nodeLatencyMonitors) Delete(ctx context.Context, name string, opts v1.DeleteOptions) error {
	return c.client.Delete().
		Resource("nodelatencymonitors").
		Name(name).
		Body(&opts).
		Do(ctx).
		Error()
}

// DeleteCollection deletes a collection of objects.
func (c *nodeLatencyMonitors) DeleteCollection(ctx context.Context, opts v1.DeleteOptions, listOpts v1.ListOptions) error {
	var timeout time.Duration
	if listOpts.TimeoutSeconds != nil {
		timeout = time.Duration(*listOpts.TimeoutSeconds) * time.Second
	}
	return c.client.Delete().
		Resource("nodelatencymonitors").
		VersionedParams(&listOpts, scheme.ParameterCodec).
		Timeout(timeout).
		Body(&opts).
		Do(ctx).
		Error()
}

// Patch applies the patch and returns the patched nodeLatencyMonitor.
func (c *nodeLatencyMonitors) Patch(ctx context.Context, name string, pt types.PatchType, data []byte, opts v1.PatchOptions, subresources ...string) (result *v1alpha2.NodeLatencyMonitor, err error) {
	result = &v1alpha2.NodeLatencyMonitor{}
	err = c.client.Patch(pt).
		Resource("nodelatencymonitors").
		Name(name).
		SubResource(subresources...).
		VersionedParams(&opts, scheme.ParameterCodec).
		Body(data).
		Do(ctx).
		Into(result)
	return
}
