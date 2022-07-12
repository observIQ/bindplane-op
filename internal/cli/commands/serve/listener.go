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
	"fmt"
	"io/ioutil"

	"github.com/observiq/bindplane-op/common"
)

func configureTLS(config *common.Server) (*tls.Config, error) {
	keyPair, err := tls.LoadX509KeyPair(config.Certificate, config.PrivateKey)
	if err != nil {
		return nil, fmt.Errorf("failed to load tls certificate: %w", err)
	}

	tlsConfig := tls.Config{
		Certificates: []tls.Certificate{keyPair},
		MinVersion:   tls.VersionTLS12,
	}

	if len(config.CertificateAuthority) > 0 {
		err := configureMutualTLS(&tlsConfig, config.CertificateAuthority)
		if err != nil {
			return nil, fmt.Errorf("failed to configure mTLS: %w", err)
		}
	}

	return &tlsConfig, nil
}

func configureMutualTLS(config *tls.Config, caFile []string) error {
	var caPool = x509.NewCertPool()

	for _, caFile := range caFile {
		ca, err := ioutil.ReadFile(caFile) // #nosec G304, user devices ca file path via a flag
		if err != nil {
			return fmt.Errorf("failed to read certificate authority file: %w", err)
		}

		if !caPool.AppendCertsFromPEM(ca) {
			return fmt.Errorf("failed to load certificate authority file: %s", caFile)
		}
	}

	config.ClientCAs = caPool

	if config.ClientCAs != nil {
		config.ClientAuth = tls.RequireAndVerifyClientCert
	}

	return nil
}
