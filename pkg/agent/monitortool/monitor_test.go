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
	"bytes"
	"net"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

type fakePacketConn struct {
	addr   net.IPAddr
	buffer *bytes.Buffer
}

var _ net.PacketConn = (*fakePacketConn)(nil)

func (pc *fakePacketConn) ReadFrom(p []byte) (int, net.Addr, error) {
	// Make a copy from buffer
	copyBuffer := bytes.NewBuffer(pc.buffer.Bytes())
	n, _ := copyBuffer.Read(p)
	return n, &pc.addr, nil
}

func (pc *fakePacketConn) WriteTo(p []byte, addr net.Addr) (int, error) {
	return pc.buffer.Write(p)
}

func (pc *fakePacketConn) Close() error {
	return nil
}

func (pc *fakePacketConn) LocalAddr() net.Addr {
	return &pc.addr
}

func (pc *fakePacketConn) SetDeadline(t time.Time) error {
	return nil
}

func (pc *fakePacketConn) SetReadDeadline(t time.Time) error {
	return nil
}

func (pc *fakePacketConn) SetWriteDeadline(t time.Time) error {
	return nil
}

func TestNodeLatencyMonitor_sendPing(t *testing.T) {
	nodeLatencyMonitor := &NodeLatencyMonitor{
		latencyStore: NewLatencyStore(false),
	}
	tests := []struct {
		packetConn net.PacketConn
		addr       net.IP
	}{
		{
			packetConn: &fakePacketConn{
				addr:   net.IPAddr{IP: net.ParseIP("127.0.0.1")},
				buffer: bytes.NewBuffer([]byte{}),
			},
			addr: net.ParseIP("127.0.0.1"),
		},
		{
			packetConn: &fakePacketConn{
				addr:   net.IPAddr{IP: net.ParseIP("::1")},
				buffer: bytes.NewBuffer([]byte{}),
			},
			addr: net.ParseIP("::1"),
		},
	}

	for _, tt := range tests {
		err := nodeLatencyMonitor.sendPing(tt.packetConn, tt.addr)
		assert.Nil(t, err)
	}
}

func TestNodeLatencyMonitor_recvPing(t *testing.T) {
	nodeLatencyMonitor := &NodeLatencyMonitor{
		latencyStore: NewLatencyStore(false),
	}
	tests := []struct {
		packetConn net.PacketConn
		addr       net.IP
		isIPv4     bool
	}{
		{
			packetConn: &fakePacketConn{
				addr:   net.IPAddr{IP: net.ParseIP("127.0.0.1")},
				buffer: bytes.NewBuffer([]byte{}),
			},
			addr:   net.ParseIP("127.0.0.1"),
			isIPv4: true,
		},
		{
			packetConn: &fakePacketConn{
				addr:   net.IPAddr{IP: net.ParseIP("::1")},
				buffer: bytes.NewBuffer([]byte{}),
			},
			addr:   net.ParseIP("::1"),
			isIPv4: false,
		},
	}
	for _, tt := range tests {
		err := nodeLatencyMonitor.sendPing(tt.packetConn, tt.addr)
		assert.Nil(t, err)
		go nodeLatencyMonitor.recvPing(tt.packetConn, tt.isIPv4)
		res, ok := nodeLatencyMonitor.latencyStore.getNodeIPLatencyEntry(tt.addr.String())
		assert.NotNil(t, res.LastRecvTime)
		assert.True(t, ok)
	}
}
