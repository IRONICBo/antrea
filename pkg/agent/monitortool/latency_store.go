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
	"sync"
	"time"

	corev1 "k8s.io/api/core/v1"
)

type LatencyStore struct {
	// Maybe we need to use small lock for the map
	mutex         sync.RWMutex
	connectionMap map[string]*Connection
	nodeInfo      map[string]*corev1.Node
}

// TODO1: use LRU cache to store the latency of the connection?
// TODO2: we only support ipv4 now
type Connection struct {
	// The source IP of the connection
	FromIP string
	// The destination IP of the connection
	ToIP string
	// The latency of the connection
	Latency time.Duration
	// The status of the connection
	Status bool
	// The last time the connection was updated
	LastUpdated time.Time
	// The time the connection was created.
	CreatedAt time.Time
}

func NewLatencyStore() *LatencyStore {
	return &LatencyStore{
		connectionMap: make(map[string]*Connection),
	}
}

func (l *LatencyStore) AddConnToMap(connKey string, conn *Connection) {
	l.mutex.Lock()
	defer l.mutex.Unlock()
	l.connectionMap[connKey] = conn
}

func (l *LatencyStore) GetConnByKey(connKey string) (*Connection, bool) {
	l.mutex.RLock()
	defer l.mutex.RUnlock()
	conn, found := l.connectionMap[connKey]
	return conn, found
}

func (l *LatencyStore) DeleteConnByKey(connKey string) {
	l.mutex.Lock()
	defer l.mutex.Unlock()
	delete(l.connectionMap, connKey)
}

func (l *LatencyStore) UpdateConnByKey(connKey string, conn *Connection) {
	l.mutex.Lock()
	defer l.mutex.Unlock()
	l.connectionMap[connKey] = conn
}
