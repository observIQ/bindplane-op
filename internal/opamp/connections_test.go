// Copyright  observIQ, Inc.
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

package opamp

import (
	"context"
	"net"
	"testing"

	"github.com/open-telemetry/opamp-go/protobufs"
	opamp "github.com/open-telemetry/opamp-go/server/types"
	"github.com/stretchr/testify/require"
)

type testAddr struct {
	address string
}

var _ net.Addr = (*testAddr)(nil)

func (a *testAddr) Network() string {
	return "tcp"
}

func (a *testAddr) String() string {
	return a.address
}

type testConnection struct {
	agentID string
	addr    testAddr
}

var _ opamp.Connection = (*testConnection)(nil)

func (c *testConnection) RemoteAddr() net.Addr {
	return &c.addr
}

func (c *testConnection) Send(ctx context.Context, message *protobufs.ServerToAgent) error {
	return nil
}

func TestConnect(t *testing.T) {
	agentID := "1"
	c := newConnections()
	conn := testConnection{agentID: agentID}
	c.connect(&conn, agentID)
	require.Equal(t, []string{agentID}, c.agentIDs(), "should have agentID 1 connected")
	require.Equal(t, &conn, c.connection(agentID), "should be able to lookup connection by agentID")
	require.Equal(t, agentID, c.agentID(&conn), "should be able to lookup agentID by connection")
}

func TestDisconnect(t *testing.T) {
	agentID := "1"
	c := newConnections()
	conn := testConnection{agentID: agentID}
	c.connect(&conn, agentID)
	require.Equal(t, []string{agentID}, c.agentIDs(), "should have agentID 1 connected")
	c.disconnect(&conn)
	require.Equal(t, []string{}, c.agentIDs(), "should have no connections")
	require.Equal(t, nil, c.connection(agentID), "should have no connection by agentID")
	require.Equal(t, "", c.agentID(&conn), "should have no agentID by connection")
}

func TestConnected(t *testing.T) {
	c := newConnections()
	c.connect(&testConnection{agentID: "1"}, "1")
	require.Equal(t, []string{"1"}, c.agentIDs(), "should have agentID 1 connected")
	require.True(t, c.connected("1"), "should have agentID 1 connected")
}
