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

package trace

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestNewOTLPExporter(t *testing.T) {
	cases := []struct {
		name   string
		config OpenTelemetryTracing
		errStr string
	}{
		{
			"missing-credentials",
			OpenTelemetryTracing{
				Endpoint: "localhost:5555",
			},
			"no transport security set",
		},
		{
			"plain-text",
			func() OpenTelemetryTracing {
				x := OpenTelemetryTracing{
					Endpoint: "localhost:5555",
				}
				x.TLS.Insecure = true
				return x
			}(),
			"",
		},
		{
			"invalid-endpoint",
			func() OpenTelemetryTracing {
				x := OpenTelemetryTracing{
					Endpoint: "31",
				}
				x.TLS.Insecure = true
				return x
			}(),
			"failed to parse gRPC endpoint",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), time.Second)
			defer cancel()

			out, err := NewOTLPExporter(ctx, tc.config, nil)

			if tc.errStr != "" {
				require.Error(t, err)
				require.Contains(t, err.Error(), tc.errStr)
				return
			}

			require.NoError(t, err)
			require.NotNil(t, out)
		})
	}
}
