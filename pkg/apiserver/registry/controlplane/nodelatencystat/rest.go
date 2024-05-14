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

// import (
// 	"context"

// 	"antrea.io/antrea/pkg/apis/controlplane"
// 	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
// 	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
// 	"k8s.io/apimachinery/pkg/runtime"
// 	"k8s.io/apiserver/pkg/registry/rest"
// )

// // nodeLatencyCollector is the interface required by the handler.
// type nodeLatencyCollector interface {
// 	Collect(summary *controlplane.NodeIPLatencyStat)
// 	Get(name string) *controlplane.NodeIPLatencyStat
// 	List() []*controlplane.NodeIPLatencyStat
// }

// type REST struct {
// 	nodeLatencyCollector nodeLatencyCollector
// }

// var (
// 	_ rest.Scoper               = &REST{}
// 	_ rest.Getter               = &REST{}
// 	_ rest.SingularNameProvider = &REST{}
// 	_ rest.Creater              = &REST{}
// )

// // NewREST returns a REST object that will work against API services.
// func NewREST(c nodeLatencyCollector) *REST {
// 	return &REST{c}
// }

// func (r *REST) New() runtime.Object {
// 	return &controlplane.NodeIPLatencyStat{}
// }

// func (r *REST) Destroy() {
// }

// func (r *REST) Create(ctx context.Context, obj runtime.Object, createValidation rest.ValidateObjectFunc, options *v1.CreateOptions) (runtime.Object, error) {
// 	// Try to store the NodeIPLatencyStat in the store.
// 	summary := obj.(*controlplane.NodeIPLatencyStat)
// 	r.nodeLatencyCollector.Collect(summary)
// 	// a valid runtime.Object must be returned, otherwise the client would throw error.
// 	return &controlplane.NodeIPLatencyStat{}, nil
// }

// func (r *REST) Get(ctx context.Context, name string, options *metav1.GetOptions) (runtime.Object, error) {
// 	// Try to retrieve the NodeIPLatencyStat from the store.
// 	entry := r.nodeLatencyCollector.Get(name)
// 	return entry, nil
// }

// func (r *REST) NamespaceScoped() bool {
// 	return false
// }

// func (r *REST) GetSingularName() string {
// 	return "nodeiplatencystat"
// }
