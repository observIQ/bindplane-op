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

	texporter "github.com/GoogleCloudPlatform/opentelemetry-operations-go/exporter/trace"
	"go.opentelemetry.io/otel/sdk/resource"
	"go.opentelemetry.io/otel/sdk/trace"
	"google.golang.org/api/option"
)

// GoogleCloudTracing is configuration for tracing to Google Cloud Monitoring
type GoogleCloudTracing struct {
	ProjectID       string `mapstructure:"projectID,omitempty" yaml:"projectID,omitempty"`
	CredentialsFile string `mapstructure:"credentialsFile,omitempty" yaml:"credentialsFile,omitempty"`
}

// NewGoogleCloudExporter returns a new Google Cloud TracerProvider.
func NewGoogleCloudExporter(ctx context.Context, config GoogleCloudTracing, resource *resource.Resource) (*trace.TracerProvider, error) {
	exporter, err := texporter.New(
		texporter.WithContext(ctx),
		texporter.WithProjectID(config.ProjectID),
		texporter.WithTraceClientOptions([]option.ClientOption{
			option.WithCredentialsFile(config.CredentialsFile),
		}),
	)
	if err != nil {
		return nil, err
	}

	return trace.NewTracerProvider(
		trace.WithBatcher(exporter),
		trace.WithResource(resource),
	), nil
}
