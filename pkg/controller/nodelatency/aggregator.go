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

package nodelatency

import (
	"sync"

	"k8s.io/klog/v2"

	"antrea.io/antrea/pkg/apis/controlplane"
)

const (
	uidIndex           = "uid"
	GroupNameIndexName = "groupName"
)

// NodeIPLatencyStats is a struct that holds the latency stats for each node.
type NodeIPLatencyStats struct {
	mutex sync.RWMutex

	// map[node name] and list of each node's latency stats.
	nodeLatencyStats map[string]NodeLatencyStat
}

// NodeLatencyStat is a struct that holds the latency stats for each node.
type NodeLatencyStat struct {
	// TODO: Add current node gateway ip to the struct?
	// NodeLatencyEntryList is a map[key is destination gateway ip] that holds the latency stats for destination nodes.
	NodeLatencyEntryList map[string]NodeLatencyEntry
}

// NodeLatencyEntry is a struct that holds the latency stats for each destination node.
type NodeLatencyEntry struct {
	// NodeName is the name of the destination node.
	LastSendTime int64
	// LastRecvTime is the time when the last packet was received from the destination node.
	LastRecvTime int64
	// LastMeasuredRTT is the last measured round-trip time between the source and destination nodes.
	LastMeasuredRTT int64
}

func NewNodeIPLatencyStats() *NodeIPLatencyStats {
	return &NodeIPLatencyStats{
		nodeLatencyStats: make(map[string]NodeLatencyStat),
	}
}

// Aggregator collects the node ip latency statistics from antrea-agents and caches them in memory. It provides methods
// for node ip latency API handlers to query them. It implements the following interfaces:
// - pkg/apiserver/registry/controlplane/nodeiplatencystat.nodeLatencyCollector
type Aggregator struct {
	// map[node name] and list of each node's latency stats.
	nodeLatencyStats *NodeIPLatencyStats
	// dataCh is the channel that buffers the NodeIPLatencyStat sent by antrea-agents.
	dataCh chan *controlplane.NodeIPLatencyStat
}

// NewAggregator creates a new Aggregator.
func NewAggregator() *Aggregator {
	aggregator := &Aggregator{
		nodeLatencyStats: NewNodeIPLatencyStats(),
		dataCh:           make(chan *controlplane.NodeIPLatencyStat, 1000),
	}

	return aggregator
}

// Collect collects the node ip latency stats asynchronously to avoid the competition for the lock and to save clients
// from pending on it.
func (a *Aggregator) Collect(summary *controlplane.NodeIPLatencyStat) {
	a.dataCh <- summary
}

// Get returns the latency stats for the provided node name.
func (a *Aggregator) Get(nodeName string) *controlplane.NodeIPLatencyStat {
	a.nodeLatencyStats.mutex.RLock()
	defer a.nodeLatencyStats.mutex.RUnlock()

	result := &controlplane.NodeIPLatencyStat{
		NodeName: nodeName,
	}
	entryList := a.nodeLatencyStats.nodeLatencyStats[nodeName]
	for gatewayIP, entry := range entryList.NodeLatencyEntryList {
		result.NodeIPLatencyList = append(result.NodeIPLatencyList, controlplane.NodeIPLatencyEntry{
			GatewayIP:       gatewayIP,
			LastSendTime:    entry.LastSendTime,
			LastRecvTime:    entry.LastRecvTime,
			LastMeasuredRTT: entry.LastMeasuredRTT,
		})
	}

	return result
}

// Run runs a loop that keeps taking node ip latency stat from the data channel and actually collecting them until the
// provided stop channel is closed.
func (a *Aggregator) Run(stopCh <-chan struct{}) {
	klog.Info("Starting node ip latency stat aggregator")
	defer klog.Info("Shutting down node ip latency stat aggregator")

	for {
		select {
		case summary := <-a.dataCh:
			a.doCollect(summary)
		case <-stopCh:
			return
		}
	}
}

func (a *Aggregator) doCollect(summary *controlplane.NodeIPLatencyStat) {
	a.nodeLatencyStats.mutex.Lock()
	defer a.nodeLatencyStats.mutex.Unlock()

	nodeName := summary.NodeName
	nodeLatencyStat, exists := a.nodeLatencyStats.nodeLatencyStats[nodeName]
	if !exists {
		nodeLatencyStat = NodeLatencyStat{
			NodeLatencyEntryList: make(map[string]NodeLatencyEntry),
		}
	}

	for _, entry := range summary.NodeIPLatencyList {
		nodeLatencyStat.NodeLatencyEntryList[entry.GatewayIP] = NodeLatencyEntry{
			LastSendTime:    entry.LastSendTime,
			LastRecvTime:    entry.LastRecvTime,
			LastMeasuredRTT: entry.LastMeasuredRTT,
		}
	}

	a.nodeLatencyStats.nodeLatencyStats[nodeName] = nodeLatencyStat
}
