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
	"testing"

	"github.com/stretchr/testify/assert"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"antrea.io/antrea/pkg/apis/controlplane"
)

func TestREST(t *testing.T) {
	r := NewREST(nil)
	assert.Equal(t, &controlplane.NodeIPLatencyStat{}, r.New())
	assert.False(t, r.NamespaceScoped())
}

type fakeCollector struct {
	gotSummary *controlplane.NodeIPLatencyStat
}

func (f *fakeCollector) Collect(summary *controlplane.NodeIPLatencyStat) {
	f.gotSummary = summary
}

func (f *fakeCollector) Get(name string) (*controlplane.NodeIPLatencyStat, error) {
	return nil, nil
}

func TestRESTCreate(t *testing.T) {
	collector := &fakeCollector{}
	r := NewREST(collector)

	summary := &controlplane.NodeIPLatencyStat{
		ObjectMeta: v1.ObjectMeta{
			Name: "foo",
		},
	}
	actualObj, err := r.Create(context.TODO(), summary, nil, &v1.CreateOptions{})
	assert.NoError(t, err)
	// Empty struct is returned on success.
	assert.Equal(t, &controlplane.NodeIPLatencyStat{}, actualObj)
	assert.Equal(t, summary, collector.gotSummary)
}
