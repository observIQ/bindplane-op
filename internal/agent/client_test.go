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

package agent

import (
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func newTestAgentVersionsServer(t *testing.T, file string) *httptest.Server {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, err := ioutil.ReadFile(filepath.Join("testdata", file))
		require.NoError(t, err)

		w.Header().Add("Content-type", "application/json")
		_, err = w.Write(body)
		require.NoError(t, err)
	}))
	return server
}

func TestLatestVersion(t *testing.T) {
	server := newTestAgentVersionsServer(t, "client_latest_version.json")
	defer server.Close()

	client := NewClient(ClientSettings{AgentVersionsURL: server.URL})
	version, err := client.LatestVersion()
	require.NoError(t, err)
	require.Equal(t, "2.0.6", version.Version)
	require.Equal(t, "https://storage.googleapis.com/observiq-cloud/observiq-agent/2.0.6/darwin-arm64/installer/observiq-agent-installer.sh", version.Downloads["darwin-arm64"][Installer])
}

func TestVersion(t *testing.T) {
	server := newTestAgentVersionsServer(t, "client_latest_version.json")
	defer server.Close()

	client := NewClient(ClientSettings{AgentVersionsURL: server.URL})
	version, err := client.Version("2.0.6")
	require.NoError(t, err)
	require.Equal(t, "2.0.6", version.Version)
	require.Equal(t, "https://storage.googleapis.com/observiq-cloud/observiq-agent/2.0.6/darwin-arm64/installer/observiq-agent-installer.sh", version.Downloads["darwin-arm64"][Installer])
}
