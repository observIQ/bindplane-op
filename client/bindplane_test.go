// Copyright  observIQ, Inc
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

package client

import (
	"fmt"
	"testing"
	"time"

	"github.com/observiq/bindplane-op/common"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func TestQueryOption(t *testing.T) {
	cases := []struct {
		name    string
		optFunc func() []QueryOption
		expect  queryOptions
	}{
		// WithSelector
		{
			name: "selector",
			optFunc: func() []QueryOption {
				return []QueryOption{WithSelector("dev")}
			},
			expect: queryOptions{
				selector: "dev",
			},
		},
		{
			name: "selector-nop",
			optFunc: func() []QueryOption {
				return []QueryOption{WithSelector("")}
			},
			expect: queryOptions{},
		},
		// WithQuery
		{
			name: "query",
			optFunc: func() []QueryOption {
				return []QueryOption{WithQuery("test query")}
			},
			expect: queryOptions{
				query: "test query",
			},
		},
		{
			name: "query-nop",
			optFunc: func() []QueryOption {
				return []QueryOption{WithQuery("")}
			},
			expect: queryOptions{},
		},
		// WithOffset
		{
			name: "offset",
			optFunc: func() []QueryOption {
				return []QueryOption{WithOffset(1000)}
			},
			expect: queryOptions{
				offset: 1000,
			},
		},
		{
			name: "offset-nop",
			optFunc: func() []QueryOption {
				return []QueryOption{WithOffset(0)}
			},
			expect: queryOptions{},
		},
		// WithLimit
		{
			name: "limit",
			optFunc: func() []QueryOption {
				return []QueryOption{WithLimit(12)}
			},
			expect: queryOptions{
				limit: 12,
			},
		},
		{
			name: "limit-nop",
			optFunc: func() []QueryOption {
				return []QueryOption{WithLimit(0)}
			},
			expect: queryOptions{},
		},
		// WithSort
		{
			name: "sort",
			optFunc: func() []QueryOption {
				return []QueryOption{WithSort("hostname")}
			},
			expect: queryOptions{
				sort: "hostname",
			},
		},
		{
			name: "sort-nop",
			optFunc: func() []QueryOption {
				return []QueryOption{WithSort("")}
			},
			expect: queryOptions{},
		},
		// Multiple Options
		{
			name: "multi",
			optFunc: func() []QueryOption {
				return []QueryOption{
					WithSelector("select"),
					WithQuery("host=dev"),
					WithOffset(111),
					WithLimit(22),
					WithSort("hostname"),
				}
			},
			expect: queryOptions{
				selector: "select",
				query:    "host=dev",
				offset:   111,
				limit:    22,
				sort:     "hostname",
			},
		},
		{
			name: "multi-nop",
			optFunc: func() []QueryOption {
				return []QueryOption{
					WithSelector(""),
					WithQuery(""),
					WithOffset(0),
					WithLimit(0),
					WithSort(""),
				}
			},
			expect: queryOptions{},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			out := makeQueryOptions(tc.optFunc())
			require.Equal(t, tc.expect, out)
		})
	}
}

func TestNewBindPlane(t *testing.T) {
	cases := []struct {
		name      string
		client    *common.Client
		logger    *zap.Logger
		expect    *BindplaneClient
		expectErr string
	}{
		{
			name:   "default",
			client: &common.InitConfig("").Client,
			logger: zap.NewNop(),
			expect: &BindplaneClient{
				config: &common.Client{
					Common: common.Common{},
				},
			},
		},
		{
			name: "fields",
			client: &common.Client{
				Common: common.Common{
					Host:     "10.99.1.5",
					Port:     "2000",
					Username: "devel",
				},
			},
			logger: zap.NewNop(),
			expect: &BindplaneClient{
				config: &common.Client{
					Common: common.Common{
						Host:     "10.99.1.5",
						Port:     "2000",
						Username: "devel",
					},
				},
			},
		},
		{
			name: "tls",
			client: &common.Client{
				Common: common.Common{
					TLSConfig: common.TLSConfig{
						Certificate: "../internal/cli/commands/serve/testdata/bindplane.crt",
						PrivateKey:  "../internal/cli/commands/serve/testdata/bindplane.key",
						CertificateAuthority: []string{
							"../internal/cli/commands/serve/testdata/bindplane-ca.crt",
						},
					},
				},
			},
			logger: zap.NewNop(),
			expect: &BindplaneClient{
				config: &common.Client{
					Common: common.Common{
						TLSConfig: common.TLSConfig{
							Certificate: "../internal/cli/commands/serve/testdata/bindplane.crt",
							PrivateKey:  "../internal/cli/commands/serve/testdata/bindplane.key",
							CertificateAuthority: []string{
								"../internal/cli/commands/serve/testdata/bindplane-ca.crt",
							},
						},
					},
				},
			},
		},
		{
			name: "tls-invalid-cert-path",
			client: &common.Client{
				Common: common.Common{
					TLSConfig: common.TLSConfig{
						Certificate: "../internal/cli/commands/serve/testdata/bindplane.crt.invalid",
						PrivateKey:  "../internal/cli/commands/serve/testdata/bindplane.key",
					},
				},
			},
			logger:    zap.NewNop(),
			expectErr: "failed to configure TLS client: failed to load tls certificate: open ../internal/cli/commands/serve/testdata/bindplane.crt.invalid",
		},
		{
			name: "tls-invalid-key-path",
			client: &common.Client{
				Common: common.Common{
					TLSConfig: common.TLSConfig{
						Certificate: "../internal/cli/commands/serve/testdata/bindplane.crt",
						PrivateKey:  "../internal/cli/commands/serve/testdata/bindplane.key.invalid",
					},
				},
			},
			logger:    zap.NewNop(),
			expectErr: "failed to configure TLS client: failed to load tls certificate: open ../internal/cli/commands/serve/testdata/bindplane.key.invalid",
		},
		{
			name: "tls-invalid-ca-path",
			client: &common.Client{
				Common: common.Common{
					TLSConfig: common.TLSConfig{
						Certificate: "../internal/cli/commands/serve/testdata/bindplane.crt",
						PrivateKey:  "../internal/cli/commands/serve/testdata/bindplane.key",
						CertificateAuthority: []string{
							"../internal/cli/commands/serve/testdata/bindplane-ca.crt.invalid",
						},
					},
				},
			},
			logger:    zap.NewNop(),
			expectErr: "failed to configure TLS client: failed to read certificate authority file: open ../internal/cli/commands/serve/testdata/bindplane-ca.crt.invalid",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			out, err := NewBindPlane(tc.client, tc.logger)
			if tc.expectErr != "" {
				require.Error(t, err)
				require.Contains(t, err.Error(), tc.expectErr)
				return
			}
			require.NoError(t, err)

			require.NotNil(t, out)
			require.NotNil(t, out.client)
			require.NotNil(t, out.config)
			require.NotNil(t, out.Logger)
			require.Equal(t, tc.expect.config, out.config)
			require.Equal(t, time.Second*20, out.client.GetClient().Timeout)

			if tc.client.Username != "" {
				require.Equal(t, tc.client.Username, out.client.UserInfo.Username)
			}
			if tc.client.Password != "" {
				require.Equal(t, tc.client.Password, out.client.UserInfo.Password)
			}

			base := fmt.Sprintf("%s/v1", tc.client.BindPlaneURL())
			require.Equal(t, base, out.client.BaseURL)
		})
	}
}
