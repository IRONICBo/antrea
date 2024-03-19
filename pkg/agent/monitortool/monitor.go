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

package monitortool

import (
	"context"
	"sync"
	"time"

	"antrea.io/antrea/pkg/agent/nodeip"
)

// MonitorTool is a tool to monitor the latency of the node.
type MonitorTool struct {
	nodeTracker  nodeip.Checker
	latencyStore *LatencyStore

	interval time.Duration
	timeout  time.Duration
}

// TODO: Maybe we need to implement it like podStore.
// Simple nodeStore struct
type NodeStore struct {
	mutex     sync.RWMutex
	PingItems map[string]PingItem
}

// TODO: NodeInternalIP/NodeExternalIP
// We only need to store the NodeInternalIP of the node.
// In first step, we only use the nodeip tracker to get the node internal/external IP.
type PingItem struct {
	// Name is the name of the node.
	Name string
	// IP is the IP of the node.
	IPs []string
}

func NewNodeStore() *NodeStore {
	return &NodeStore{
		PingItems: make(map[string]PingItem),
	}
}

func (n *NodeStore) Clear() {
	n.mutex.Lock()
	defer n.mutex.Unlock()
	n.PingItems = make(map[string]PingItem)
}

func (n *NodeStore) AddPingItem(name string, ips []string) {
	n.mutex.Lock()
	defer n.mutex.Unlock()
	n.PingItems[name] = PingItem{
		Name: name,
		IPs:  ips,
	}
}

func (n *NodeStore) GetPingItem(name string) (PingItem, bool) {
	n.mutex.RLock()
	defer n.mutex.RUnlock()
	item, found := n.PingItems[name]
	return item, found
}

func (n *NodeStore) DeletePingItem(name string) {
	n.mutex.Lock()
	defer n.mutex.Unlock()
	delete(n.PingItems, name)
}

func (n *NodeStore) ListPingItems() []PingItem {
	n.mutex.RLock()
	defer n.mutex.RUnlock()
	items := make([]PingItem, 0, len(n.PingItems))
	for _, item := range n.PingItems {
		items = append(items, item)
	}
	return items
}

func (n *NodeStore) UpdatePingItem(name string, ips []string) {
	n.mutex.Lock()
	defer n.mutex.Unlock()
	n.PingItems[name] = PingItem{
		Name: name,
		IPs:  ips,
	}
}

func NewNodeLatencyMonitor(nodeIPCheck nodeip.Checker, interval, timeout time.Duration) *MonitorTool {
	return &MonitorTool{
		nodeTracker:  nodeIPCheck,
		latencyStore: NewLatencyStore(),
		interval:     interval,
		timeout:      timeout,
	}
}

func (m *MonitorTool) pingAll(nodes map[string]string) {
	// TODO: Add ping limiter

	ctx, cancel := context.WithTimeout(context.Background(), m.timeout)
	defer cancel()

	for ip, name := range nodes {
		ok, rtt := m.pingNode(ctx)
		if ok {
		}
	}
}

func (m *MonitorTool) pingNode(ctx context.Context, string, ip string) (bool, time.Duration) {

	return false, 0
}

func (m *MonitorTool) Run(stopCh <-chan struct{}) {

}
