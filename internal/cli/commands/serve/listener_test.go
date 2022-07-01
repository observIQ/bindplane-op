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

package serve

import (
	"crypto/tls"
	"crypto/x509"
	"io/ioutil"
	"testing"

	"github.com/observiq/bindplane-op/common"
	"github.com/stretchr/testify/require"
)

func Test_configureTLS(t *testing.T) {
	cases := []struct {
		name       string
		serverConf *common.Server
		expect     *tls.Config
		errSubStr  string
	}{
		{
			"tls",
			func() *common.Server {
				s := &common.Server{
					Common: common.Common{
						TLSConfig: common.TLSConfig{
							Certificate: "testdata/bindplane.crt",
							PrivateKey:  "testdata/bindplane.key",
						},
					},
				}

				return s
			}(),
			func() *tls.Config {
				pair, err := tls.LoadX509KeyPair("testdata/bindplane.crt", "testdata/bindplane.key")
				if err != nil {
					t.Fatalf("failed to load expected keypair: %v", err)
				}

				return &tls.Config{
					Certificates: []tls.Certificate{pair},
					MinVersion:   tls.VersionTLS12,
					ClientAuth:   tls.NoClientCert,
				}
			}(),
			"",
		},
		{
			"mtls",
			func() *common.Server {
				s := &common.Server{
					Common: common.Common{
						TLSConfig: common.TLSConfig{
							Certificate:          "testdata/bindplane.crt",
							PrivateKey:           "testdata/bindplane.key",
							CertificateAuthority: []string{"testdata/bindplane-ca.crt"},
						},
					},
				}

				return s
			}(),
			func() *tls.Config {
				pair, err := tls.LoadX509KeyPair("testdata/bindplane.crt", "testdata/bindplane.key")
				if err != nil {
					t.Fatalf("failed to load expected keypair: %v", err)
				}

				ca, err := ioutil.ReadFile("testdata/bindplane-ca.crt")
				if err != nil {
					t.Fatalf("failed to load expected ca certificate: %v", err)
				}

				caPool := x509.NewCertPool()
				if !caPool.AppendCertsFromPEM(ca) {
					t.Fatal("failed to append certificate file to capool")
				}

				return &tls.Config{
					Certificates: []tls.Certificate{pair},
					MinVersion:   tls.VersionTLS12,
					ClientCAs:    caPool,
					ClientAuth:   tls.RequireAndVerifyClientCert,
				}
			}(),
			"",
		},
		{
			"invalid-certificate-path",
			func() *common.Server {
				s := &common.Server{
					Common: common.Common{
						TLSConfig: common.TLSConfig{
							Certificate: "testdata/bindplane.crt.invalid",
							PrivateKey:  "testdata/bindplane.key",
						},
					},
				}

				return s
			}(),
			nil,
			"failed to load tls certificate",
		},
		{
			"invalid-private-key-path",
			func() *common.Server {
				s := &common.Server{
					Common: common.Common{
						TLSConfig: common.TLSConfig{
							Certificate: "testdata/bindplane.crt",
							PrivateKey:  "testdata/bindplane.key.invalid",
						},
					},
				}

				return s
			}(),
			nil,
			"failed to load tls certificate",
		},
		{
			"invalid-ca",
			func() *common.Server {
				s := &common.Server{
					Common: common.Common{
						TLSConfig: common.TLSConfig{
							Certificate:          "testdata/bindplane.crt",
							PrivateKey:           "testdata/bindplane.key",
							CertificateAuthority: []string{"testdata/bindplane-ca.crt.invalid"},
						},
					},
				}

				return s
			}(),
			nil,
			"failed to read certificate authority file",
		},
		{
			"malformed-ca",
			func() *common.Server {
				s := &common.Server{
					Common: common.Common{
						TLSConfig: common.TLSConfig{
							Certificate:          "testdata/bindplane.crt",
							PrivateKey:           "testdata/bindplane.key",
							CertificateAuthority: []string{"testdata/bindplane-ca.crt.malformed"},
						},
					},
				}

				return s
			}(),
			nil,
			"failed to load certificate authority file",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			output, err := configureTLS(tc.serverConf)

			if tc.errSubStr != "" {
				require.Error(t, err)
				require.Contains(t, err.Error(), tc.errSubStr)
				return
			}

			require.NoError(t, err)
			require.Equal(t, tc.expect.Certificates, output.Certificates)
			require.Equal(t, tc.expect.MinVersion, output.MinVersion)

			// mTLS tests
			require.Equal(t, tc.expect.ClientAuth, output.ClientAuth)
			if len(tc.serverConf.CertificateAuthority) != 0 {
				require.NotNil(t, tc.expect.ClientCAs)
				require.Equal(t, tc.expect.ClientCAs.Subjects(), output.ClientCAs.Subjects())
			}

		})
	}
}
