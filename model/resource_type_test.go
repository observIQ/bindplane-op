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

package model

import (
	"strings"
	"testing"

	"github.com/observiq/bindplane-op/model/otel"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
)

func TestEvalCabinDestination(t *testing.T) {
	dt := fileResource[*DestinationType](t, "testfiles/destinationtype-cabin.yaml")
	d := fileResource[*Destination](t, "testfiles/destination-cabin.yaml")
	values := dt.evalOutput(&dt.Spec.Logs, d, func(e error) {
		require.NoError(t, e)
	})
	require.Len(t, values.Receivers, 0)
	require.Len(t, values.Processors, 1)
	require.Len(t, values.Exporters, 1)
	require.Len(t, values.Extensions, 0)

	// we expect observiq-cloud to be rendered
	_, ok := values.Exporters[0][otel.NewComponentID("observiq", "observiq-cloud__cabin-production-logs")]
	require.True(t, ok)

	exportersYaml, err := yaml.Marshal(values.Exporters)
	require.NoError(t, err)

	expectYaml := strings.TrimLeft(`
- observiq/observiq-cloud__cabin-production-logs:
    endpoint: https://nozzle.app.observiq.com
    secret_key: 2c088c5e-2afc-483b-be52-e2b657fcff08
    timeout: 10s
`, "\n")

	require.Equal(t, expectYaml, string(exportersYaml))
}

func TestEvalGoogleCloud(t *testing.T) {
	dt := fileResource[*DestinationType](t, "testfiles/destinationtype-googlecloud.yaml")
	d := fileResource[*Destination](t, "testfiles/destination-googlecloud.yaml")
	values := dt.eval(d, func(e error) {
		require.NoError(t, e)
	})
	require.Len(t, values[otel.Logs].Receivers, 0)
	require.Len(t, values[otel.Logs].Processors, 1)
	require.Len(t, values[otel.Logs].Exporters, 1)
	require.Len(t, values[otel.Logs].Extensions, 0)
	require.Len(t, values[otel.Metrics].Receivers, 0)
	require.Len(t, values[otel.Metrics].Processors, 2)
	require.Len(t, values[otel.Metrics].Exporters, 1)
	require.Len(t, values[otel.Metrics].Extensions, 0)
	require.Len(t, values[otel.Traces].Receivers, 0)
	require.Len(t, values[otel.Traces].Processors, 1)
	require.Len(t, values[otel.Traces].Exporters, 1)
	require.Len(t, values[otel.Traces].Extensions, 0)
}

func TestTelemetryTypes(t *testing.T) {
	macosSourceType := fileResource[*SourceType](t, "testfiles/sourcetype-macos.yaml")
	otlpSourceType := fileResource[*SourceType](t, "testfiles/sourcetype-otlp.yaml")

	tests := []struct {
		description string
		sourceType  *SourceType
		expect      []otel.PipelineType
	}{
		{
			description: "macos supports logs and metrics",
			sourceType:  macosSourceType,
			expect:      []otel.PipelineType{otel.Logs, otel.Metrics},
		},
		{
			description: "otlp supports logs, metrics, and traces",
			sourceType:  otlpSourceType,
			expect:      []otel.PipelineType{otel.Logs, otel.Metrics, otel.Traces},
		},
	}

	for _, test := range tests {
		t.Run(test.description, func(t *testing.T) {
			got := test.sourceType.Spec.TelemetryTypes()
			require.ElementsMatch(t, test.expect, got)
		},
		)
	}
}
