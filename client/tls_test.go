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
	"crypto/tls"
	"crypto/x509"
	"io/ioutil"
	"testing"

	"github.com/stretchr/testify/require"
)

func Test_tlsClient(t *testing.T) {
	cases := []struct {
		name      string
		cert      string
		key       string
		ca        []string
		expect    *tls.Config
		expectErr bool
	}{
		{
			"tls",
			"../internal/cli/commands/serve/testdata/bindplane.crt",
			"../internal/cli/commands/serve/testdata/bindplane.key",
			[]string{},
			&tls.Config{
				Certificates: func() []tls.Certificate {
					pair, err := tls.LoadX509KeyPair(
						"../internal/cli/commands/serve/testdata/bindplane.crt",
						"../internal/cli/commands/serve/testdata/bindplane.key",
					)
					if err != nil {
						t.Errorf("setup failed: %v", err)
						t.FailNow()
					}
					return []tls.Certificate{pair}
				}(),
			},
			false,
		},
		{
			"mutual-tls",
			"../internal/cli/commands/serve/testdata/bindplane.crt",
			"../internal/cli/commands/serve/testdata/bindplane.key",
			[]string{
				"../internal/cli/commands/serve/testdata/bindplane-ca.crt",
			},
			&tls.Config{
				Certificates: func() []tls.Certificate {
					t, _ := tls.LoadX509KeyPair(
						"../internal/cli/commands/serve/testdata/bindplane.crt",
						"../internal/cli/commands/serve/testdata/bindplane.key",
					)
					return []tls.Certificate{t}
				}(),
				RootCAs: func() *x509.CertPool {
					path := "../internal/cli/commands/serve/testdata/bindplane-ca.crt"
					ca, err := ioutil.ReadFile(path)
					if err != nil {
						t.Errorf("setup failed: %v", err)
						t.FailNow()
					}
					var pool = x509.NewCertPool()
					pool.AppendCertsFromPEM(ca)
					return pool
				}(),
			},
			false,
		},
		{
			"mutual-tls-invalid-ca",
			"../internal/cli/commands/serve/testdata/bindplane.crt",
			"../internal/cli/commands/serve/testdata/bindplane.key",
			[]string{
				// tls.go will never be a valid x509 pem file
				"tls.go",
			},
			nil,
			true,
		},
		{
			"tls-invalid-cert-path",
			"../internal/cli/commands/serve/testdata/bindplane.crt.invalid",
			"../internal/cli/commands/serve/testdata/bindplane.key",
			[]string{
				"../internal/cli/commands/serve/testdata/bindplane-ca.crt",
			},
			nil,
			true,
		},
		{
			"tls-invalid-key-path",
			"../internal/cli/commands/serve/testdata/bindplane.crt",
			"../internal/cli/commands/serve/testdata/bindplane.key.invalid",
			[]string{
				"../internal/cli/commands/serve/testdata/bindplane-ca.crt",
			},
			nil,
			true,
		},
		{
			"tls-invalid-ca-path",
			"../internal/cli/commands/serve/testdata/bindplane.crt",
			"../internal/cli/commands/serve/testdata/bindplane.key",
			[]string{
				"../internal/cli/commands/serve/testdata/bindplane-ca.crt.invalid",
			},
			nil,
			true,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			out, err := tlsClient(tc.cert, tc.key, tc.ca)
			if tc.expectErr {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
			require.NotNil(t, out)
			require.Equal(t, tc.expect.Certificates, out.Certificates)
			if len(tc.ca) > 0 {
				require.NotNil(t, out.RootCAs)
			}
			require.Equal(t, tls.VersionTLS13, int(out.MinVersion))
		})
	}
}
