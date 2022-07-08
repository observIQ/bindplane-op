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

package client

import (
	"crypto/tls"
	"crypto/x509"
	"errors"
	"fmt"
	"io/ioutil"
)

// tlsClient takes file paths for a certificate, private key, and certificate authority.
// Returns a *tls.Config. All parameters are optional. If mutual TLS is desired, all parameters
// must be passed.
func tlsClient(cert, key string, caCertFile []string, skipVerify bool) (*tls.Config, error) {
	tlsConfig := &tls.Config{
		MinVersion:         tls.VersionTLS13,
		InsecureSkipVerify: skipVerify,
	}

	// CA certificate can be used to trust private certificates
	if len(caCertFile) > 0 {
		var caPool = x509.NewCertPool()

		for _, caCertFile := range caCertFile {
			ca, err := ioutil.ReadFile(caCertFile) // #nosec G304, user defines ca file path via a flag
			if err != nil {
				return nil, fmt.Errorf("failed to read certificate authority file: %w", err)
			}

			if !caPool.AppendCertsFromPEM(ca) {
				return nil, errors.New("failed to append certificate authority to root ca pool")
			}
		}

		tlsConfig.RootCAs = caPool
	}

	// Client key pair can be used for mutual TLS
	if cert != "" && key != "" {
		keypair, err := tls.LoadX509KeyPair(cert, key)
		if err != nil {
			return nil, fmt.Errorf("failed to load tls certificate: %w", err)
		}
		tlsConfig.Certificates = []tls.Certificate{keypair}
	}

	return tlsConfig, nil
}
