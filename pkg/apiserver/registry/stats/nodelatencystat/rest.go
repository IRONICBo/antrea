// Copyright 2024 Antrea Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package nodelatencystat

import (
	"context"
	"fmt"

	statsv1alpha1 "antrea.io/antrea/pkg/apis/stats/v1alpha1"
	"k8s.io/apimachinery/pkg/api/meta"
	metatable "k8s.io/apimachinery/pkg/api/meta/table"
	"k8s.io/apimachinery/pkg/apis/meta/internalversion"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apiserver/pkg/registry/rest"
	"k8s.io/client-go/tools/cache"
	"k8s.io/klog/v2"
)

// nodeLatencyCollector is the interface required by the handler.
type nodeLatencyCollector interface {
	Collect(summary *statsv1alpha1.NodeIPLatencyStat)
	Get(name string) *statsv1alpha1.NodeIPLatencyStat
	List() []statsv1alpha1.NodeIPLatencyStat
}

type REST struct {
	nodeLatencyCollector nodeLatencyCollector
}

// nodeLatencyCollectorImpl implements nodeLatencyCollector.
type nodeLatencyCollectorImpl struct {
	nodeLatencyStats cache.Indexer
	dataCh           chan *statsv1alpha1.NodeIPLatencyStat
}

func (n *nodeLatencyCollectorImpl) Collect(summary *statsv1alpha1.NodeIPLatencyStat) {
	n.dataCh <- summary
}

func (n *nodeLatencyCollectorImpl) doCollect(summary *statsv1alpha1.NodeIPLatencyStat) {
	n.nodeLatencyStats.Update(summary)
}

func (n *nodeLatencyCollectorImpl) Get(name string) *statsv1alpha1.NodeIPLatencyStat {
	obj, exists, err := n.nodeLatencyStats.GetByKey(name)
	if err != nil || !exists {
		return nil
	}
	return obj.(*statsv1alpha1.NodeIPLatencyStat)
}

func (n *nodeLatencyCollectorImpl) List() []statsv1alpha1.NodeIPLatencyStat {
	objs := n.nodeLatencyStats.List()
	entries := make([]statsv1alpha1.NodeIPLatencyStat, len(objs))
	for i := range objs {
		entries[i] = *(objs[i].(*statsv1alpha1.NodeIPLatencyStat))
	}
	return entries
}

var (
	_ rest.Storage              = &REST{}
	_ rest.Scoper               = &REST{}
	_ rest.Getter               = &REST{}
	_ rest.Lister               = &REST{}
	_ rest.SingularNameProvider = &REST{}
)

const (
	uidIndex           = "uid"
	GroupNameIndexName = "groupName"
)

// uidIndexFunc is an index function that indexes based on an object's UID.
func uidIndexFunc(obj interface{}) ([]string, error) {
	meta, err := meta.Accessor(obj)
	if err != nil {
		return []string{""}, fmt.Errorf("object has no meta: %v", err)
	}
	return []string{string(meta.GetUID())}, nil
}

// NewREST returns a REST object that will work against API services.
func NewREST() *REST {
	nodeLatencyCollector := &nodeLatencyCollectorImpl{
		nodeLatencyStats: cache.NewIndexer(cache.MetaNamespaceKeyFunc, cache.Indexers{cache.NamespaceIndex: cache.MetaNamespaceIndexFunc, uidIndex: uidIndexFunc}),
		dataCh:           make(chan *statsv1alpha1.NodeIPLatencyStat, 1000),
	}

	go func() {
		for summary := range nodeLatencyCollector.dataCh {
			// Store the NodeIPLatencyStat in the store.
			nodeLatencyCollector.doCollect(summary)
		}
	}()

	return &REST{
		nodeLatencyCollector: &nodeLatencyCollectorImpl{
			nodeLatencyStats: cache.NewIndexer(cache.MetaNamespaceKeyFunc, cache.Indexers{cache.NamespaceIndex: cache.MetaNamespaceIndexFunc, uidIndex: uidIndexFunc}),
			dataCh:           make(chan *statsv1alpha1.NodeIPLatencyStat, 1000),
		},
	}
}

func (r *REST) New() runtime.Object {
	return &statsv1alpha1.NodeIPLatencyStat{}
}

func (r *REST) Destroy() {
}

func (r *REST) Create(ctx context.Context, obj runtime.Object, createValidation rest.ValidateObjectFunc, options *metav1.CreateOptions) (runtime.Object, error) {
	// Try to store the NodeIPLatencyStat in the store.
	summary := obj.(*statsv1alpha1.NodeIPLatencyStat)
	r.nodeLatencyCollector.Collect(summary)
	klog.InfoS("NodeIPLatencyStat created", "name", summary.Name)
	// a valid runtime.Object must be returned, otherwise the client would throw error.
	return &statsv1alpha1.NodeIPLatencyStat{}, nil
}

func (r *REST) Get(ctx context.Context, name string, options *metav1.GetOptions) (runtime.Object, error) {
	// Try to retrieve the NodeIPLatencyStat from the store.
	entry := r.nodeLatencyCollector.Get(name)
	return entry, nil
}

func (r *REST) NewList() runtime.Object {
	return &statsv1alpha1.NodeIPLatencyStatList{}
}

func (r *REST) List(ctx context.Context, options *internalversion.ListOptions) (runtime.Object, error) {
	// Try to retrieve the NodeIPLatencyStat from the store.
	entries := r.nodeLatencyCollector.List()
	klog.InfoS("NodeIPLatencyStat list", "entries", entries)
	return &statsv1alpha1.NodeIPLatencyStatList{Items: entries}, nil
}

func (r *REST) ConvertToTable(ctx context.Context, obj runtime.Object, tableOptions runtime.Object) (*metav1.Table, error) {
	// Convert the NodeIPLatencyStat to a Table object.
	table := &metav1.Table{
		ColumnDefinitions: []metav1.TableColumnDefinition{
			{Name: "SourceNodeName", Type: "string", Format: "name", Description: "Source node name."},
			{Name: "NodeIPLatencyList", Type: "array", Format: "string", Description: "Node IP latency list."},
		},
	}
	if m, err := meta.ListAccessor(obj); err == nil {
		table.ResourceVersion = m.GetResourceVersion()
		table.Continue = m.GetContinue()
		table.RemainingItemCount = m.GetRemainingItemCount()
	} else {
		if m, err := meta.CommonAccessor(obj); err == nil {
			table.ResourceVersion = m.GetResourceVersion()
		}
	}
	var err error
	table.Rows, err = metatable.MetaToTableRow(obj, func(obj runtime.Object, m metav1.Object, name, age string) ([]interface{}, error) {
		summary := obj.(*statsv1alpha1.NodeIPLatencyStat)
		return []interface{}{name, summary.NodeIPLatencyList}, nil
	})
	return table, err
}

func (r *REST) NamespaceScoped() bool {
	return false
}

func (r *REST) GetSingularName() string {
	return "nodeiplatencystat"
}
