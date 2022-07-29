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
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
)

func validateResource[T Resource](t *testing.T, name string) T {
	return fileResource[T](t, filepath.Join("testfiles", "validate", name))
}
func testResource[T Resource](t *testing.T, name string) T {
	return fileResource[T](t, filepath.Join("testfiles", name))
}
func fileResource[T Resource](t *testing.T, path string) T {
	resources, err := ResourcesFromFile(path)
	require.NoError(t, err)

	parsed, err := ParseResources(resources)
	require.NoError(t, err)
	require.Len(t, parsed, 1)

	resource, ok := parsed[0].(T)
	require.True(t, ok)
	return resource
}

type testResourceStore struct {
	sources          map[string]*Source
	sourceTypes      map[string]*SourceType
	processors       map[string]*Processor
	processorTypes   map[string]*ProcessorType
	destinations     map[string]*Destination
	destinationTypes map[string]*DestinationType
}

func newTestResourceStore() *testResourceStore {
	return &testResourceStore{
		sources:          map[string]*Source{},
		sourceTypes:      map[string]*SourceType{},
		processors:       map[string]*Processor{},
		processorTypes:   map[string]*ProcessorType{},
		destinations:     map[string]*Destination{},
		destinationTypes: map[string]*DestinationType{},
	}
}

var _ ResourceStore = (*testResourceStore)(nil)

func (s *testResourceStore) Source(name string) (*Source, error) {
	return s.sources[name], nil
}
func (s *testResourceStore) SourceType(name string) (*SourceType, error) {
	return s.sourceTypes[name], nil
}
func (s *testResourceStore) Processor(name string) (*Processor, error) {
	return s.processors[name], nil
}
func (s *testResourceStore) ProcessorType(name string) (*ProcessorType, error) {
	return s.processorTypes[name], nil
}
func (s *testResourceStore) Destination(name string) (*Destination, error) {
	return s.destinations[name], nil
}
func (s *testResourceStore) DestinationType(name string) (*DestinationType, error) {
	return s.destinationTypes[name], nil
}

func TestParseConfiguration(t *testing.T) {
	path := filepath.Join("testfiles", "configuration-raw.yaml")
	bytes, err := os.ReadFile(path)
	require.NoError(t, err, "failed to read the testfile")
	var configuration Configuration
	err = yaml.Unmarshal(bytes, &configuration)
	require.NoError(t, err)
	require.Equal(t, "cabin-production-configuration", configuration.Metadata.Name)
	require.Equal(t, "receivers:", strings.Split(configuration.Spec.Raw, "\n")[0])
}

func TestEvalConfiguration(t *testing.T) {
	store := newTestResourceStore()

	macos := testResource[*SourceType](t, "sourcetype-macos.yaml")
	store.sourceTypes[macos.Name()] = macos

	cabin := testResource[*Destination](t, "destination-cabin.yaml")
	store.destinations[cabin.Name()] = cabin

	cabinType := testResource[*DestinationType](t, "destinationtype-cabin.yaml")
	store.destinationTypes[cabinType.Name()] = cabinType

	configuration := testResource[*Configuration](t, "configuration-macos-sources.yaml")
	result, err := configuration.Render(context.TODO(), store)
	require.NoError(t, err)

	expect := strings.TrimLeft(`
receivers:
    plugin/MacOS__source0__journald:
        plugin:
            name: journald
    plugin/MacOS__source0__macos:
        parameters:
            - name: enable_system_log
              value: false
            - name: system_log_path
              value: /var/log/system.log
            - name: enable_install_log
              value: true
            - name: install_log_path
              value: /var/log/install.log
            - name: start_at
              value: end
        plugin:
            name: macos
    plugin/MacOS__source1__journald:
        plugin:
            name: journald
    plugin/MacOS__source1__macos:
        parameters:
            - name: enable_system_log
              value: true
            - name: system_log_path
              value: /var/log/system.log
            - name: enable_install_log
              value: true
            - name: install_log_path
              value: /var/log/install.log
            - name: start_at
              value: end
        plugin:
            name: macos
processors:
    batch/observiq-cloud__cabin-production-logs: null
exporters:
    observiq/observiq-cloud__cabin-production-logs:
        endpoint: https://nozzle.app.observiq.com
        secret_key: 2c088c5e-2afc-483b-be52-e2b657fcff08
        timeout: 10s
service:
    pipelines:
        logs/MacOS__source0__cabin-production-logs:
            receivers:
                - plugin/MacOS__source0__macos
                - plugin/MacOS__source0__journald
            processors:
                - batch/observiq-cloud__cabin-production-logs
            exporters:
                - observiq/observiq-cloud__cabin-production-logs
        logs/MacOS__source1__cabin-production-logs:
            receivers:
                - plugin/MacOS__source1__macos
                - plugin/MacOS__source1__journald
            processors:
                - batch/observiq-cloud__cabin-production-logs
            exporters:
                - observiq/observiq-cloud__cabin-production-logs
`, "\n")

	require.Equal(t, expect, result)
}

func TestEvalConfiguration2(t *testing.T) {
	store := newTestResourceStore()

	macos := testResource[*SourceType](t, "sourcetype-macos.yaml")
	store.sourceTypes[macos.Name()] = macos

	googleCloudType := testResource[*DestinationType](t, "destinationtype-googlecloud.yaml")
	store.destinationTypes[googleCloudType.Name()] = googleCloudType

	configuration := testResource[*Configuration](t, "configuration-macos-googlecloud.yaml")
	result, err := configuration.Render(context.TODO(), store)
	require.NoError(t, err)

	expect := strings.TrimLeft(`
receivers:
    hostmetrics/MacOS__source0:
        collection_interval: 1m
        scrapers:
            load: null
    hostmetrics/MacOS__source1:
        collection_interval: 1m
        scrapers:
            load: null
    plugin/MacOS__source0__journald:
        plugin:
            name: journald
    plugin/MacOS__source0__macos:
        parameters:
            - name: enable_system_log
              value: false
            - name: system_log_path
              value: /var/log/system.log
            - name: enable_install_log
              value: true
            - name: install_log_path
              value: /var/log/install.log
            - name: start_at
              value: end
        plugin:
            name: macos
    plugin/MacOS__source1__journald:
        plugin:
            name: journald
    plugin/MacOS__source1__macos:
        parameters:
            - name: enable_system_log
              value: true
            - name: system_log_path
              value: /var/log/system.log
            - name: enable_install_log
              value: true
            - name: install_log_path
              value: /var/log/install.log
            - name: start_at
              value: end
        plugin:
            name: macos
processors:
    batch/googlecloud__destination0: null
    normalizesums/googlecloud__destination0: null
exporters:
    googlecloud/googlecloud__destination0: null
service:
    pipelines:
        logs/MacOS__source0__destination0:
            receivers:
                - plugin/MacOS__source0__macos
                - plugin/MacOS__source0__journald
            processors:
                - batch/googlecloud__destination0
            exporters:
                - googlecloud/googlecloud__destination0
        logs/MacOS__source1__destination0:
            receivers:
                - plugin/MacOS__source1__macos
                - plugin/MacOS__source1__journald
            processors:
                - batch/googlecloud__destination0
            exporters:
                - googlecloud/googlecloud__destination0
        metrics/MacOS__source0__destination0:
            receivers:
                - hostmetrics/MacOS__source0
            processors:
                - normalizesums/googlecloud__destination0
                - batch/googlecloud__destination0
            exporters:
                - googlecloud/googlecloud__destination0
        metrics/MacOS__source1__destination0:
            receivers:
                - hostmetrics/MacOS__source1
            processors:
                - normalizesums/googlecloud__destination0
                - batch/googlecloud__destination0
            exporters:
                - googlecloud/googlecloud__destination0
`, "\n")

	require.Equal(t, expect, result)
}

func TestEvalConfiguration3(t *testing.T) {
	store := newTestResourceStore()

	otlp := testResource[*SourceType](t, "sourcetype-otlp.yaml")
	store.sourceTypes[otlp.Name()] = otlp

	googleCloudType := testResource[*DestinationType](t, "destinationtype-otlp.yaml")
	store.destinationTypes[googleCloudType.Name()] = googleCloudType

	configuration := testResource[*Configuration](t, "configuration-otlp.yaml")
	result, err := configuration.Render(context.TODO(), store)
	require.NoError(t, err)

	expect := strings.TrimLeft(`
receivers:
    otlp/otlp__source0:
        protocols:
            grpc: null
            http: null
processors:
    batch/otlp__destination0: null
exporters:
    otlp/otlp__destination0:
        endpoint: otelcol:4317
service:
    pipelines:
        logs/otlp__source0__destination0:
            receivers:
                - otlp/otlp__source0
            processors:
                - batch/otlp__destination0
            exporters:
                - otlp/otlp__destination0
        metrics/otlp__source0__destination0:
            receivers:
                - otlp/otlp__source0
            processors:
                - batch/otlp__destination0
            exporters:
                - otlp/otlp__destination0
        traces/otlp__source0__destination0:
            receivers:
                - otlp/otlp__source0
            processors:
                - batch/otlp__destination0
            exporters:
                - otlp/otlp__destination0
`, "\n")

	require.Equal(t, expect, result)
}

func TestEvalConfiguration4(t *testing.T) {
	store := newTestResourceStore()

	postgresql := testResource[*SourceType](t, "sourcetype-postgresql.yaml")
	store.sourceTypes[postgresql.Name()] = postgresql

	googleCloudType := testResource[*DestinationType](t, "destinationtype-googlecloud.yaml")
	store.destinationTypes[googleCloudType.Name()] = googleCloudType

	configuration := testResource[*Configuration](t, "configuration-postgresql-googlecloud.yaml")
	result, err := configuration.Render(context.TODO(), store)
	require.NoError(t, err)

	expect := strings.TrimLeft(`
receivers:
    plugin/postgresql__source0__postgresql:
        parameters:
            postgresql_log_path:
                - /var/log/postgresql/postgresql*.log
                - /var/lib/pgsql/data/log/postgresql*.log
                - /var/lib/pgsql/*/data/log/postgresql*.log
            start_at: end
        path: $OIQ_OTEL_COLLECTOR_HOME/plugins/postgresql_logs.yaml
processors:
    batch/googlecloud__destination0: null
    normalizesums/googlecloud__destination0: null
    resourceattributetransposer/postgresql__source0:
        operations:
            - from: host.name
              to: agent
    resourcedetection/postgresql__source0:
        detectors:
            - system
        system:
            hostname_sources:
                - os
exporters:
    googlecloud/googlecloud__destination0: null
service:
    pipelines:
        logs/postgresql__source0__destination0:
            receivers:
                - plugin/postgresql__source0__postgresql
            processors:
                - batch/googlecloud__destination0
            exporters:
                - googlecloud/googlecloud__destination0
`, "\n")

	require.Equal(t, expect, result)
}

func TestEvalConfiguration5(t *testing.T) {
	store := newTestResourceStore()

	postgresql := testResource[*SourceType](t, "sourcetype-macos.yaml")
	store.sourceTypes[postgresql.Name()] = postgresql

	googleCloudType := testResource[*DestinationType](t, "destinationtype-googlecloud.yaml")
	store.destinationTypes[googleCloudType.Name()] = googleCloudType

	googleCloud := testResource[*Destination](t, "destination-googlecloud.yaml")
	store.destinations[googleCloud.Name()] = googleCloud

	resourceAttributeTransposerType := testResource[*ProcessorType](t, "processortype-resourceattributetransposer.yaml")
	store.processorTypes[resourceAttributeTransposerType.Name()] = resourceAttributeTransposerType

	configuration := testResource[*Configuration](t, "configuration-macos-processors.yaml")
	result, err := configuration.Render(context.TODO(), store)
	require.NoError(t, err)

	expect := strings.TrimLeft(`
receivers:
    hostmetrics/MacOS__source0:
        collection_interval: 1m
        scrapers:
            load: null
    plugin/MacOS__source0__journald:
        plugin:
            name: journald
    plugin/MacOS__source0__macos:
        parameters:
            - name: enable_system_log
              value: false
            - name: system_log_path
              value: /var/log/system.log
            - name: enable_install_log
              value: true
            - name: install_log_path
              value: /var/log/install.log
            - name: start_at
              value: end
        plugin:
            name: macos
processors:
    batch/googlecloud__googlecloud: null
    normalizesums/googlecloud__googlecloud: null
    resourceattributetransposer/resource-attribute-transposer__MacOS__source0__processor0:
        operations:
            - from: from.attribute
              to: to.attribute
    resourceattributetransposer/resource-attribute-transposer__MacOS__source0__processor1:
        operations:
            - from: from.attribute2
              to: to.attribute2
exporters:
    googlecloud/googlecloud__googlecloud: null
service:
    pipelines:
        logs/MacOS__source0__googlecloud:
            receivers:
                - plugin/MacOS__source0__macos
                - plugin/MacOS__source0__journald
            processors:
                - resourceattributetransposer/resource-attribute-transposer__MacOS__source0__processor0
                - resourceattributetransposer/resource-attribute-transposer__MacOS__source0__processor1
                - batch/googlecloud__googlecloud
            exporters:
                - googlecloud/googlecloud__googlecloud
        metrics/MacOS__source0__googlecloud:
            receivers:
                - hostmetrics/MacOS__source0
            processors:
                - resourceattributetransposer/resource-attribute-transposer__MacOS__source0__processor0
                - resourceattributetransposer/resource-attribute-transposer__MacOS__source0__processor1
                - normalizesums/googlecloud__googlecloud
                - batch/googlecloud__googlecloud
            exporters:
                - googlecloud/googlecloud__googlecloud
`, "\n")

	require.Equal(t, expect, result)
}

func TestEvalConfigurationFailsMissingResource(t *testing.T) {
	store := newTestResourceStore()

	postgresql := testResource[*SourceType](t, "sourcetype-macos.yaml")
	store.sourceTypes[postgresql.Name()] = postgresql

	googleCloudType := testResource[*DestinationType](t, "destinationtype-googlecloud.yaml")
	store.destinationTypes[googleCloudType.Name()] = googleCloudType

	googleCloud := testResource[*Destination](t, "destination-googlecloud.yaml")
	store.destinations[googleCloud.Name()] = googleCloud

	resourceAttributeTransposerType := testResource[*ProcessorType](t, "processortype-resourceattributetransposer.yaml")
	store.processorTypes[resourceAttributeTransposerType.Name()] = resourceAttributeTransposerType

	configuration := testResource[*Configuration](t, "configuration-macos-processors.yaml")

	tests := []struct {
		name            string
		deleteResources func()
		expectError     string
		expect          string
	}{
		{
			name:            "deletes sourceType",
			deleteResources: func() { delete(store.sourceTypes, postgresql.Name()) },
			expectError:     "1 error occurred:\n\t* unknown SourceType: MacOS\n\n",
		},
		{
			name:            "deletes googleCloudType",
			deleteResources: func() { delete(store.destinationTypes, googleCloudType.Name()) },
			expectError:     "1 error occurred:\n\t* unknown DestinationType: googlecloud\n\n",
		},
		{
			name:            "deletes destination",
			deleteResources: func() { delete(store.destinations, googleCloud.Name()) },
			expectError:     "1 error occurred:\n\t* unknown Destination: googlecloud\n\n",
		},
		{
			name:            "deletes processorType",
			deleteResources: func() { delete(store.processorTypes, resourceAttributeTransposerType.Name()) },
			expectError:     "2 errors occurred:\n\t* unknown ProcessorType: resource-attribute-transposer\n\t* unknown ProcessorType: resource-attribute-transposer\n\n",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			// before rendering, delete resources that we reference
			test.deleteResources()

			_, err := configuration.Render(context.TODO(), store)
			require.Error(t, err)
			require.Equal(t, test.expectError, err.Error())

			// reset for next iteration
			store.sourceTypes[postgresql.Name()] = postgresql
			store.destinationTypes[googleCloudType.Name()] = googleCloudType
			store.destinations[googleCloud.Name()] = googleCloud
			store.processorTypes[resourceAttributeTransposerType.Name()] = resourceAttributeTransposerType
		})
	}
}

func TestDuplicate(t *testing.T) {
	duplicateName := "duplicate-config"

	configuration := testResource[*Configuration](t, "configuration-macos-googlecloud.yaml")
	require.NotNil(t, configuration)

	new := configuration.Duplicate(duplicateName)
	require.NotNil(t, new)

	t.Run("equal sources, destinations", func(t *testing.T) {
		require.Equal(t, configuration.Spec.Sources, new.Spec.Sources)
		require.Equal(t, configuration.Spec.Destinations, new.Spec.Destinations)
	})

	t.Run("replace name, id, and match labels", func(t *testing.T) {
		// Set the duplicate name
		require.Equal(t, new.Metadata.Name, duplicateName)

		// Set a new ID
		require.NotEqual(t, new.Metadata.ID, configuration.Metadata.ID)

		// Set the configuration matchLabel
		require.Contains(t, new.Spec.Selector.MatchLabels, "configuration")
		require.Equal(t, new.Spec.Selector.MatchLabels["configuration"], duplicateName)
	})
}
