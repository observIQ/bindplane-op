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
	"fmt"
	"io/fs"
	"os"
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/observiq/bindplane-op/common"
	"github.com/observiq/bindplane-op/internal/cli/flags"
	"github.com/observiq/bindplane-op/model"
)

func TestSetCommand(t *testing.T) {
	h := newTestHelper()

	initializeTestFiles(t, h)
	defer cleanupTestFiles(h)

	t.Run("returns cobra command", func(t *testing.T) {
		s := SetCommand(h)
		assert.IsType(t, &cobra.Command{}, s)
	})

	t.Run("error with no name argument", func(t *testing.T) {
		s := SetCommand(h)
		err := s.Execute()
		assert.Error(t, err)
	})

	var setParamsTests = []struct {
		name  string
		flag  string
		value interface{}
		want  model.Profile
	}{
		{
			name:  "port",
			flag:  "--port",
			value: "8000",
			want: *model.NewProfileWithMetadata(model.Metadata{Name: "port"}, model.ProfileSpec{
				Common: common.Common{Port: "8000"},
			}),
		},
		{
			name:  "host",
			flag:  "--host",
			value: "host",
			want: *model.NewProfileWithMetadata(model.Metadata{Name: "host"}, model.ProfileSpec{
				Common: common.Common{Host: "host"},
			}),
		},
		{
			name:  "server-url",
			flag:  "--server-url",
			value: "http://www.test.com",
			want: func() model.Profile {
				c := model.NewProfileWithMetadata(model.Metadata{Name: "server-url"}, model.ProfileSpec{
					Common: common.Common{
						ServerURL: "http://www.test.com",
					},
				})
				return *c
			}(),
		},
		{
			name:  "username",
			flag:  "--username",
			value: "username",
			want: *model.NewProfileWithMetadata(model.Metadata{Name: "username"}, model.ProfileSpec{
				Common: common.Common{Username: "username"},
			})},
		{
			name:  "offline",
			flag:  "--offline",
			value: true,
			want: *model.NewProfileWithMetadata(model.Metadata{Name: "offline"}, model.ProfileSpec{
				Server: common.Server{Offline: true},
			})},
		{
			name:  "remote-url",
			flag:  "--remote-url",
			value: "http://localhost:3001",
			want: *model.NewProfileWithMetadata(model.Metadata{Name: "remote-url"}, model.ProfileSpec{
				Server: common.Server{RemoteURL: "http://localhost:3001"},
			})},
		{
			name:  "secret-key",
			flag:  "--secret-key",
			value: "5ce40143-61d7-43cb-bd81-051453f05dfe",
			want: *model.NewProfileWithMetadata(model.Metadata{Name: "secret-key"}, model.ProfileSpec{
				Server: common.Server{SecretKey: "5ce40143-61d7-43cb-bd81-051453f05dfe"},
			})},
		{
			name:  "storage-file-path",
			flag:  "--storage-file-path",
			value: "/path/to/file",
			want: *model.NewProfileWithMetadata(model.Metadata{Name: "storage-file-path"}, model.ProfileSpec{
				Server: common.Server{StorageFilePath: "/path/to/file"},
			})},
		{
			name:  "downloads-folder-path",
			flag:  "--downloads-folder-path",
			value: "/path/to/downloads",
			want: *model.NewProfileWithMetadata(model.Metadata{Name: "downloads-folder-path"}, model.ProfileSpec{
				Server: common.Server{DownloadsFolderPath: "/path/to/downloads"},
			})},
		{
			name:  "agents-service-url",
			flag:  "--agents-service-url",
			value: "https://agents.remote.com",
			want: *model.NewProfileWithMetadata(model.Metadata{Name: "agents-service-url"}, model.ProfileSpec{
				Server: common.Server{AgentsServiceURL: "https://agents.remote.com"},
			})},
		{
			name:  "disable-downloads-cache",
			flag:  "--disable-downloads-cache",
			value: true,
			want: *model.NewProfileWithMetadata(model.Metadata{Name: "disable-downloads-cache"}, model.ProfileSpec{
				Server: common.Server{DisableDownloadsCache: true},
			})},
		{
			name:  "password",
			flag:  "--password",
			value: "p$ssword!1",
			want: *model.NewProfileWithMetadata(model.Metadata{Name: "password"}, model.ProfileSpec{
				Common: common.Common{Password: "p$ssword!1"},
			})},
		{
			name:  "tls-cert",
			flag:  "--tls-cert",
			value: "/opt/bindplane/tls/bindplane.crt",
			want: *model.NewProfileWithMetadata(model.Metadata{Name: "tls-cert"}, model.ProfileSpec{
				Common: common.Common{
					TLSConfig: common.TLSConfig{Certificate: "/opt/bindplane/tls/bindplane.crt"},
				},
			})},
		{
			name:  "tls-key",
			flag:  "--tls-key",
			value: "/opt/bindplane/tls/bindplane.key",
			want: *model.NewProfileWithMetadata(model.Metadata{Name: "tls-key"}, model.ProfileSpec{
				Common: common.Common{
					TLSConfig: common.TLSConfig{PrivateKey: "/opt/bindplane/tls/bindplane.key"},
				},
			})},
		{
			name:  "tls-ca",
			flag:  "--tls-ca",
			value: "/opt/bindplane/tls/bindplane.key,/opt/bindplane/tls/bindplane2.key",
			want: *model.NewProfileWithMetadata(model.Metadata{Name: "tls-ca"}, model.ProfileSpec{
				Common: common.Common{
					TLSConfig: common.TLSConfig{CertificateAuthority: []string{"/opt/bindplane/tls/bindplane.key", "/opt/bindplane/tls/bindplane2.key"}},
				},
			})},
	}

	for _, test := range setParamsTests {
		h := newTestHelper()
		initializeTestFiles(t, h)
		defer cleanupTestFiles(h)

		t.Run(fmt.Sprintf("able to set flag: %s", test.flag), func(t *testing.T) {
			// This is a little weird, but "config set" checks values from the
			// global persistent flags, so it needs to "inherit" the flags to properly simulate
			// the real set.go
			c := Command(h)
			flags.Global(c)
			flags.Serve(c)

			args := []string{
				"set",
				test.name,
				fmt.Sprintf("%v=%v", test.flag, test.value),
			}

			c.SetArgs(args)
			err := c.Execute()
			require.NoError(t, err)

			profile, err := h.Folder().ReadProfile(test.name)
			assert.Equal(t, &test.want, profile)
		})
	}

	t.Run("able to overwrite existing config value", func(t *testing.T) {
		initializeTestFiles(t, h)
		defer cleanupTestFiles(h)

		localConfig, err := h.Folder().ReadProfile("local")
		require.NoError(t, err)

		assert.Equal(t, testProfile, localConfig)

		c := Command(h)
		flags.Global(c)
		flags.Serve(c)

		for _, test := range setParamsTests {
			args := []string{
				"set",
				"local",
				fmt.Sprintf("%v=%v", test.flag, test.value),
			}

			c.SetArgs(args)
			c.Execute()
		}

		want := model.NewProfileWithMetadata(model.Metadata{Name: "local"}, model.ProfileSpec{
			Common: common.Common{
				Port:      "8000",
				Host:      "host",
				Username:  "username",
				Password:  "p$ssword!1",
				ServerURL: "http://www.test.com",
				TLSConfig: common.TLSConfig{
					Certificate:          "/opt/bindplane/tls/bindplane.crt",
					PrivateKey:           "/opt/bindplane/tls/bindplane.key",
					CertificateAuthority: []string{"/opt/bindplane/tls/bindplane.key", "/opt/bindplane/tls/bindplane2.key"},
				},
			},
			Server: common.Server{
				StorageFilePath:       "/path/to/file",
				SecretKey:             "5ce40143-61d7-43cb-bd81-051453f05dfe",
				RemoteURL:             "http://localhost:3001",
				Offline:               true,
				AgentsServiceURL:      "https://agents.remote.com",
				DownloadsFolderPath:   "/path/to/downloads",
				DisableDownloadsCache: true,
			},
		})

		localConfig, err = h.Folder().ReadProfile("local")
		require.NoError(t, err)

		assert.Equal(t, want, localConfig)
	})

	t.Run("creates a profile when a context name is specified that isn't present", func(t *testing.T) {
		initializeTestFiles(t, h)
		defer cleanupTestFiles(h)

		// This is a little weird, but "config set" checks values from the
		// global persistent flags, so it needs to "inherit" the flags to properly simulate
		// the real set.go
		c := Command(h)
		flags.Global(c)
		flags.Serve(c)

		newName := "new"

		newProfile, err := h.Folder().ReadProfile(newName)
		require.Error(t, err)

		assert.Nil(t, newProfile)

		args := []string{
			"set",
			newName,
			fmt.Sprintf("%v=%v", "--host", "localhost-newhostname"),
		}

		c.SetArgs(args)
		c.Execute()

		newProfile, err = h.Folder().ReadProfile(newName)
		require.NoError(t, err)

		assert.NotNil(t, newProfile)
	})

	t.Run("returns error when writeConfigYaml fails", func(t *testing.T) {
		initializeTestFiles(t, h)
		defer cleanupTestFiles(h)

		// This is a little weird, but "config set" checks values from the
		// global persistent flags, so it needs to "inherit" the flags to properly simulate
		// the real set.go
		c := Command(h)
		flags.Global(c)
		flags.Serve(c)

		os.Chmod(h.Folder().ProfilesFolderPath(), fs.FileMode(os.O_RDONLY))
		defer func() {
			os.Chmod(h.Folder().ProfilesFolderPath(), 0750)
		}()

		args := []string{"set", "local", "--host", "5000"}
		c.SetArgs(args)

		err := c.Execute()
		assert.Error(t, err)
	})
}
