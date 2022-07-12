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
	"fmt"
	"net"

	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/sdk/resource"
	"go.opentelemetry.io/otel/sdk/trace"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// OpenTelemetryTracing is configuration for tracing to an Open Telemetry Collector
type OpenTelemetryTracing struct {
	Endpoint string `mapstructure:"endpoint,omitempty" yaml:"endpoint,omitempty"`
	TLS      struct {
		Insecure bool `mapstructure:"insecure,omitempty" yaml:"insecure,omitempty"`
	} `mapstructure:"tls,omitempty" yaml:"tls,omitempty"`
}

// NewOTLPExporter returns a new Open Telemetry TracerProvider.
func NewOTLPExporter(ctx context.Context, config OpenTelemetryTracing, resource *resource.Resource) (*trace.TracerProvider, error) {
	var dialOpts []grpc.DialOption

	// TODO(jsirianni): How to do we handle server side tls, mtls, etc?
	if config.TLS.Insecure {
		dialOpts = append(dialOpts, grpc.WithTransportCredentials(insecure.NewCredentials()))
	}

	_, _, err := net.SplitHostPort(config.Endpoint)
	if err != nil {
		return nil, fmt.Errorf("failed to parse gRPC endpoint: %v", err)
	}

	conn, err := grpc.DialContext(ctx, config.Endpoint, dialOpts...)
	if err != nil {
		return nil, fmt.Errorf("failed to dial otlp endpoint: %v", err)
	}

	exporter, err := otlptracegrpc.New(ctx, otlptracegrpc.WithGRPCConn(conn))
	if err != nil {
		return nil, fmt.Errorf("failed to create OTLP trace exporter: %v", err)
	}

	return trace.NewTracerProvider(
		trace.WithBatcher(exporter),
		trace.WithResource(resource),
	), nil
}
