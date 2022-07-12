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

package profile

import (
	"os"
	"path"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/observiq/bindplane-op/common"
	"github.com/observiq/bindplane-op/model"
)

var (
	testConfigCommonConfig = common.Common{
		ServerURL: "https://remote-address.com",
		Host:      "192.168.64.1",
		Port:      "5000",
		Username:  "admin",
		Password:  "admin",
		TLSConfig: common.TLSConfig{
			Certificate:          "tls/bindplane.crt",
			PrivateKey:           "tls/bindplane.key",
			CertificateAuthority: []string{"tls/bindplane-authority", "tls/bindplane-authority2"},
		},
	}

	testProfile = model.NewProfileWithMetadata(model.Metadata{
		Name: "local",
	}, model.ProfileSpec{
		Server:  common.Server{},
		Client:  common.Client{},
		Command: common.Command{},
		Common:  testConfigCommonConfig,
	})

	testContext = model.NewContextWithMetadata(model.Metadata{}, model.ContextSpec{
		CurrentContext: "local",
	})
)

func TestNewHelper(t *testing.T) {
	h := NewHelper("")
	assert.NotNil(t, h)
	assert.IsType(t, &helper{}, h)
}

func TestDirectory(t *testing.T) {
	h := newTestHelper()
	defer cleanupTestFiles(h)

	homedir, _ := os.UserHomeDir()
	wantPath := path.Join(homedir, ".test-bindplane")
	dirPath := h.Directory()

	assert.Equal(t, wantPath, dirPath)
}

func TestMkDir(t *testing.T) {
	h := newTestHelper()
	defer cleanupTestFiles(h)

	h.mkDir()
	assert.DirExists(t, h.Directory())
}
