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
	"net"
	"os"
	"sync"
	"sync/atomic"
	"time"

	stv1aplpha1 "antrea.io/antrea/pkg/apis/stats/v1alpha1"
	"antrea.io/antrea/pkg/util/env"
	"golang.org/x/net/icmp"
	"golang.org/x/net/ipv4"
	"golang.org/x/net/ipv6"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	coreinformers "k8s.io/client-go/informers/core/v1"
	"k8s.io/client-go/tools/cache"
	"k8s.io/klog/v2"

	"antrea.io/antrea/pkg/agent"
	config "antrea.io/antrea/pkg/agent/config"
	"antrea.io/antrea/pkg/apis/crd/v1alpha1"
	crdinformers "antrea.io/antrea/pkg/client/informers/externalversions/crd/v1alpha1"
)

var (
	icmpSeq    uint32
	icmpEchoID = os.Getpid() & 0xffff
)

const (
	IPv4ProtocolICMPRaw = "ip4:icmp"
	IPv6ProtocolICMPRaw = "ip6:ipv6-icmp"
	IPProtocol          = "ip"
	ProtocolICMP        = 1
	ProtocolICMPv6      = 58
)

// getICMPSeq returns the next sequence number as uint16,
// wrapping around to 0 after reaching the maximum value of uint16.
func getICMPSeq() uint16 {
	// Increment the sequence number atomically and get the new value.
	// We use atomic.AddUint32 and pass 1 as the increment.
	// The returned value is the new value post-increment.
	newVal := atomic.AddUint32(&icmpSeq, 1)

	return uint16(newVal)
}

// NodeLatencyMonitor is a tool to monitor the latency of the Node.
type NodeLatencyMonitor struct {
	// latencyStore is the cache to store the latency of each Nodes.
	latencyStore *LatencyStore
	// latencyConfig is the config for the latency monitor.
	latencyConfig *LatencyConfig
	// latencyConfigChanged is the channel to notify the latency config changed.
	latencyConfigChanged chan struct{}
	// isIPv4Enabled is the flag to indicate whether the IPv4 is enabled.
	isIPv4Enabled bool
	// isIPv6Enabled is the flag to indicate whether the IPv6 is enabled.
	isIPv6Enabled bool

	// antreaClientProvider provides interfaces to get antreaClient, which will be used to report the statistics
	antreaClientProvider       agent.AntreaClientProvider
	nodeInformer               coreinformers.NodeInformer
	nodeLatencyMonitorInformer crdinformers.NodeLatencyMonitorInformer
}

// LatencyConfig is the config for the latency monitor.
type LatencyConfig struct {
	// Enable is the flag to enable the latency monitor.
	Enable bool
	// Interval is the interval time to ping all Nodes.
	Interval time.Duration
}

// NewNodeLatencyMonitor creates a new NodeLatencyMonitor.
func NewNodeLatencyMonitor(antreaClientProvider agent.AntreaClientProvider,
	nodeInformer coreinformers.NodeInformer,
	nlmInformer crdinformers.NodeLatencyMonitorInformer,
	nodeConfig *config.NodeConfig,
	trafficEncapMode config.TrafficEncapModeType) *NodeLatencyMonitor {
	m := &NodeLatencyMonitor{
		latencyStore:               NewLatencyStore(trafficEncapMode.IsNetworkPolicyOnly()),
		latencyConfig:              &LatencyConfig{Enable: false},
		latencyConfigChanged:       make(chan struct{}, 1),
		antreaClientProvider:       antreaClientProvider,
		nodeInformer:               nodeInformer,
		nodeLatencyMonitorInformer: nlmInformer,
	}

	m.isIPv4Enabled, _ = config.IsIPv4Enabled(nodeConfig, trafficEncapMode)
	m.isIPv6Enabled, _ = config.IsIPv6Enabled(nodeConfig, trafficEncapMode)

	nodeInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc:    m.onNodeAdd,
		UpdateFunc: m.onNodeUpdate,
		DeleteFunc: m.onNodeDelete,
	})

	nlmInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc:    m.onNodeLatencyMonitorAdd,
		UpdateFunc: m.onNodeLatencyMonitorUpdate,
		DeleteFunc: m.onNodeLatencyMonitorDelete,
	})

	return m
}

// onNodeAdd is the event handler for adding Node.
func (m *NodeLatencyMonitor) onNodeAdd(obj interface{}) {
	node := obj.(*corev1.Node)
	m.latencyStore.addNode(node)

	klog.InfoS("Node added", "Node", klog.KObj(node))
}

// onNodeUpdate is the event handler for updating Node.
func (m *NodeLatencyMonitor) onNodeUpdate(oldObj, newObj interface{}) {
	node := newObj.(*corev1.Node)
	m.latencyStore.updateNode(node)

	klog.InfoS("Node updated", "Node", klog.KObj(node))
}

// onNodeDelete is the event handler for deleting Node.
func (m *NodeLatencyMonitor) onNodeDelete(obj interface{}) {
	node, ok := obj.(*corev1.Node)
	if !ok {
		deletedState, ok := obj.(cache.DeletedFinalStateUnknown)
		if !ok {
			klog.ErrorS(nil, "Received unexpected object", "obj", obj)
			return
		}
		node, ok = deletedState.Obj.(*corev1.Node)
		if !ok {
			klog.ErrorS(nil, "DeletedFinalStateUnknown contains non-Node object", "obj", deletedState.Obj)
			return
		}
	}

	m.latencyStore.deleteNode(node)
}

// onNodeLatencyMonitorAdd is the event handler for adding NodeLatencyMonitor.
func (m *NodeLatencyMonitor) onNodeLatencyMonitorAdd(obj interface{}) {
	nlm := obj.(*v1alpha1.NodeLatencyMonitor)
	klog.InfoS("NodeLatencyMonitor added", "NodeLatencyMonitor", klog.KObj(nlm))

	m.updateLatencyConfig(nlm)
}

// onNodeLatencyMonitorUpdate is the event handler for updating NodeLatencyMonitor.
func (m *NodeLatencyMonitor) onNodeLatencyMonitorUpdate(oldObj, newObj interface{}) {
	oldNLM := oldObj.(*v1alpha1.NodeLatencyMonitor)
	newNLM := newObj.(*v1alpha1.NodeLatencyMonitor)
	klog.InfoS("NodeLatencyMonitor updated", "NodeLatencyMonitor", klog.KObj(newNLM))

	if oldNLM.GetGeneration() == newNLM.GetGeneration() {
		return
	}

	m.updateLatencyConfig(newNLM)
}

// updateLatencyConfig updates the latency config based on the NodeLatencyMonitor CRD.
func (m *NodeLatencyMonitor) updateLatencyConfig(nlm *v1alpha1.NodeLatencyMonitor) {
	pingInterval := time.Duration(nlm.Spec.PingIntervalSeconds) * time.Second

	m.latencyConfig = &LatencyConfig{
		Enable:   true,
		Interval: pingInterval,
	}

	m.latencyConfigChanged <- struct{}{}
}

// onNodeLatencyMonitorDelete is the event handler for deleting NodeLatencyMonitor.
func (m *NodeLatencyMonitor) onNodeLatencyMonitorDelete(obj interface{}) {
	m.latencyConfig = &LatencyConfig{Enable: false}
	klog.InfoS("NodeLatencyMonitor deleted")

	m.latencyConfigChanged <- struct{}{}
}

// sendPing sends an ICMP message to the target IP address.
func (m *NodeLatencyMonitor) sendPing(socket net.PacketConn, addr net.IP) error {
	var requestType icmp.Type

	ip := &net.IPAddr{IP: addr}

	if addr.To4() == nil {
		requestType = ipv6.ICMPTypeEchoRequest
	} else {
		requestType = ipv4.ICMPTypeEcho
	}

	timeStart := time.Now()
	seqID := getICMPSeq()
	body := &icmp.Echo{
		ID:   icmpEchoID,
		Seq:  int(seqID),
		Data: []byte(timeStart.Format(time.RFC3339Nano)),
	}
	msg := icmp.Message{
		Type: requestType,
		Code: 0,
		Body: body,
	}
	klog.InfoS("Sending ICMP message", "IP", ip, "SeqID", seqID, "body", body)

	// Serialize the ICMP message
	msgBytes, err := msg.Marshal(nil)
	if err != nil {
		return err
	}

	// Send the ICMP message
	_, err = socket.WriteTo(msgBytes, ip)
	if err != nil {
		return err
	}

	// Create or update the latency store
	mutator := func(entry *NodeIPLatencyEntry) {
		entry.LastSendTime = timeStart
	}
	m.latencyStore.SetNodeIPLatencyEntry(addr.String(), mutator)

	return nil
}

// recvPing receives an ICMP message from the target IP address.
func (m *NodeLatencyMonitor) recvPing(socket net.PacketConn, isIPv4 bool) {
	// We only expect small packets, if we receive a larger packet, we will drop the extra data.
	readBuffer := make([]byte, 128)
	for {
		n, peer, err := socket.ReadFrom(readBuffer)
		if err != nil {
			// When the socket is closed in the Run method, this error will be logged, which is not ideal.
			// In the future, we may try setting a ReadDeadline on the socket before each ReadFrom and using
			// a channel to signal that the loop should terminate.
			klog.ErrorS(err, "Failed to read ICMP message")
			return
		}

		destIP := peer.String()

		// Parse the ICMP message
		var msg *icmp.Message
		if isIPv4 {
			msg, err = icmp.ParseMessage(ProtocolICMP, readBuffer[:n])
			if err != nil {
				klog.ErrorS(err, "Failed to parse ICMP message")
				continue
			}
			if msg.Type != ipv4.ICMPTypeEchoReply {
				klog.InfoS("Failed to match ICMPTypeEchoReply", "Msg", msg)
				continue
			}
		} else {
			msg, err = icmp.ParseMessage(ProtocolICMPv6, readBuffer)
			if err != nil {
				klog.ErrorS(err, "Failed to parse ICMP message")
				continue
			}
			if msg.Type != ipv6.ICMPTypeEchoReply {
				klog.InfoS("Failed to match ICMPTypeEchoReply", "Msg", msg)
				continue
			}
		}

		echo, ok := msg.Body.(*icmp.Echo)
		if !ok {
			klog.ErrorS(nil, "Failed to assert type as *icmp.Echo")
			continue
		}

		klog.InfoS("Recv ICMP message", "IP", destIP, "Msg", msg)

		// Parse the time from the ICMP data
		sentTime, err := time.Parse(time.RFC3339Nano, string(echo.Data))
		if err != nil {
			klog.ErrorS(err, "Failed to parse time from ICMP data")
			continue
		}

		// Calculate the round-trip time
		end := time.Now()
		rtt := end.Sub(sentTime)
		klog.InfoS("Updating latency entry for Node IP", "IP", destIP, "lastSendTime", sentTime, "lastRecvTime", end, "RTT", rtt)

		// Update the latency store
		mutator := func(entry *NodeIPLatencyEntry) {
			entry.LastSendTime = sentTime
			entry.LastRecvTime = end
			entry.LastMeasuredRTT = rtt
		}
		m.latencyStore.SetNodeIPLatencyEntry(destIP, mutator)
	}
}

// pingAll sends ICMP messages to all the Nodes.
func (m *NodeLatencyMonitor) pingAll(ipv4Socket, ipv6Socket net.PacketConn) {
	klog.InfoS("Pinging all Nodes")
	nodeIPs := m.latencyStore.ListNodeIPs()
	for _, toIP := range nodeIPs {
		if toIP.To4() != nil && ipv4Socket != nil {
			if err := m.sendPing(ipv4Socket, toIP); err != nil {
				klog.ErrorS(nil, "Cannot send ICMP message to Node IP because socket is not initialized for IPv4", "IP", toIP)
			}
		} else if toIP.To16() != nil && ipv6Socket != nil {
			if err := m.sendPing(ipv6Socket, toIP); err != nil {
				klog.ErrorS(nil, "Cannot send ICMP message to Node IP because socket is not initialized for IPv6", "IP", toIP)
			}
		} else {
			klog.ErrorS(nil, "Cannot send ICMP message to Node IP because socket is not initialized for IP family", "IP", toIP)
		}
	}
	klog.InfoS("Done pinging all Nodes")
}

// GetSummary returns the latency summary of the given Node IP.
func (m *NodeLatencyMonitor) GetSummary() *stv1aplpha1.NodeIPLatencyStat {
	nodeName, _ := env.GetNodeName()
	return &stv1aplpha1.NodeIPLatencyStat{
		ObjectMeta: metav1.ObjectMeta{
			Name: nodeName,
		},
		NodeIPLatencyList: m.latencyStore.ConvertList(),
	}
}

func (m *NodeLatencyMonitor) report() {
	summary := m.GetSummary()
	if summary == nil {
		klog.InfoS("Latency summary is nil")
		return
	}
	antreaClient, err := m.antreaClientProvider.GetAntreaClient()
	if err != nil {
		klog.ErrorS(err, "Failed to get Antrea client")
		return
	}
	if _, err := antreaClient.StatsV1alpha1().NodeIPLatencyStats().Create(context.TODO(), summary, metav1.CreateOptions{}); err != nil {
		klog.ErrorS(err, "Failed to update NodeIPLatencyStats")
	}
}

// Run starts the NodeLatencyMonitor.
func (m *NodeLatencyMonitor) Run(stopCh <-chan struct{}) {
	go m.nodeLatencyMonitorInformer.Informer().Run(stopCh)
	go m.nodeInformer.Informer().Run(stopCh)
	go m.monitorLoop(stopCh)

	<-stopCh
}

// monitorLoop is the main loop to monitor the latency of the Node.
func (m *NodeLatencyMonitor) monitorLoop(stopCh <-chan struct{}) {
	klog.InfoS("NodeLatencyMonitor is running")
	// Low level goroutine to handle ping loop
	var ticker *time.Ticker
	var tickerCh <-chan time.Time
	var ipv4Socket, ipv6Socket net.PacketConn
	var err error

	defer func() {
		if ipv4Socket != nil {
			ipv4Socket.Close()
		}
		if ipv6Socket != nil {
			ipv6Socket.Close()
		}
		if ticker != nil {
			ticker.Stop()
		}
	}()

	// Update current ticker based on the latencyConfig
	updateTicker := func(interval time.Duration) {
		if ticker != nil {
			ticker.Stop() // Stop the current ticker
		}
		ticker = time.NewTicker(interval)
		tickerCh = ticker.C
	}

	klog.InfoS("NodeLatencyMonitor is running2")
	wg := sync.WaitGroup{}
	// Start the pingAll goroutine
	for {
		select {
		case <-tickerCh:
			// Try to send pingAll signal
			m.pingAll(ipv4Socket, ipv6Socket)
			m.report()
		case <-stopCh:
			return
		case <-m.latencyConfigChanged:
			// Start or stop the pingAll goroutine based on the latencyConfig
			if m.latencyConfig.Enable {
				// latencyConfig changed
				updateTicker(m.latencyConfig.Interval)

				// If the recvPing socket is closed,
				// recreate it if it is closed(CRD is deleted).
				if ipv4Socket == nil && m.isIPv4Enabled {
					// Create a new socket for IPv4 when it is IPv4-only
					ipv4Socket, err = icmp.ListenPacket(IPv4ProtocolICMPRaw, "0.0.0.0")
					if err != nil {
						klog.ErrorS(err, "Failed to create ICMP socket for IPv4")
						return
					}
					wg.Add(1)
					go func() {
						defer wg.Done()
						m.recvPing(ipv4Socket, true)
					}()
				}
				if ipv6Socket == nil && m.isIPv6Enabled {
					// Create a new socket for IPv6 when it is IPv6-only
					ipv6Socket, err = icmp.ListenPacket(IPv6ProtocolICMPRaw, "::")
					if err != nil {
						klog.ErrorS(err, "Failed to create ICMP socket for IPv6")
						return
					}
					wg.Add(1)
					go func() {
						defer wg.Done()
						m.recvPing(ipv6Socket, false)
					}()
				}
			} else {
				// latencyConfig deleted
				if ticker != nil {
					ticker.Stop()
					ticker = nil
				}

				// We close the sockets as a signal to recvPing that it needs to stop.
				// Note that at that point, we are guaranteed that there is no ongoing Write
				// to the socket, because pingAll runs in the same goroutine as this code.
				if ipv4Socket != nil {
					ipv4Socket.Close()
				}
				if ipv6Socket != nil {
					ipv6Socket.Close()
				}

				// After closing the sockets, wait for the recvPing goroutines to return
				wg.Wait()
				ipv4Socket = nil
				ipv6Socket = nil
			}
		}
	}
}
