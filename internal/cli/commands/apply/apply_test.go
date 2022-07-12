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

package apply

import (
	"bytes"
	"context"
	"fmt"
	"io/ioutil"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/observiq/bindplane-op/client"
	"github.com/observiq/bindplane-op/internal/cli"
	"github.com/observiq/bindplane-op/model"
)

type mockClient struct {
	client.BindPlane
	mock.Mock
}

func (s *mockClient) Agents(ctx context.Context, options ...client.QueryOption) ([]*model.Agent, error) {
	args := s.Called(ctx, options)
	return nil, args.Error(1)
}

func (s *mockClient) Apply(ctx context.Context, r []*model.AnyResource) ([]*model.AnyResourceStatus, error) {
	args := s.Called(ctx, r)
	result, _ := args.Get(0).([]*model.AnyResourceStatus)
	return result, args.Error(1)
}

func TestApply(t *testing.T) {
	destinationStatus := &model.AnyResourceStatus{
		Resource: model.AnyResource{ResourceMeta: model.ResourceMeta{Metadata: model.Metadata{Name: "resource-1"}, Kind: model.KindDestination}},
		Status:   model.StatusConfigured,
	}
	sourceStatus := &model.AnyResourceStatus{
		Resource: model.AnyResource{ResourceMeta: model.ResourceMeta{Metadata: model.Metadata{Name: "resource-2"}, Kind: model.KindSource}},
		Status:   model.StatusCreated,
	}
	configurationStatus := &model.AnyResourceStatus{
		Resource: model.AnyResource{ResourceMeta: model.ResourceMeta{Metadata: model.Metadata{Name: "resource-3"}, Kind: model.KindConfiguration}},
		Status:   model.StatusUnchanged,
	}

	client := &mockClient{}
	resourceStatuses := []*model.AnyResourceStatus{
		destinationStatus,
		sourceStatus,
		configurationStatus,
	}
	client.On("Apply", mock.Anything, mock.Anything).Return(resourceStatuses, nil)
	stub := &cli.BindPlane{
		Config: nil,
	}
	stub.SetClient(client)

	t.Run("no args", func(t *testing.T) {
		apply := Command(stub)
		err := apply.Execute()
		require.NoError(t, err)
	})

	t.Run("file doesn't exist", func(t *testing.T) {
		apply := Command(stub)
		apply.SetArgs([]string{"-f", ".does-not-exist.yaml"})
		err := apply.Execute()
		require.Error(t, err)
	})

	t.Run("error when client.Apply fails", func(t *testing.T) {
		errClient := &mockClient{}
		errClient.On("Apply", mock.Anything, mock.Anything).Return(nil, fmt.Errorf("unexpected error"))
		stub := &cli.BindPlane{
			Config: nil,
		}
		stub.SetClient(errClient)

		apply := Command(stub)
		apply.SetArgs([]string{"-f", "testfiles/macos.yaml"})

		err := apply.Execute()
		require.Error(t, err, "expects an error when client.Apply() fails.")
	})

	t.Run("applies yaml", func(t *testing.T) {
		apply := Command(stub)
		apply.SetArgs([]string{"-f", "testfiles/macos.yaml"})

		b := bytes.NewBufferString("")
		apply.SetOut(b)

		err := apply.Execute()
		require.NoError(t, err)
	})

	t.Run("can apply using an arg instead of flag", func(t *testing.T) {
		apply := Command(stub)
		apply.SetArgs([]string{"testfiles/macos.yaml"})

		b := bytes.NewBufferString("")
		apply.SetOut(b)

		err := apply.Execute()
		require.NoError(t, err)
	})

	t.Run("apply output", func(t *testing.T) {
		want := `Destination resource-1 configured
Source resource-2 created
Configuration resource-3 unchanged
`
		apply := Command(stub)
		apply.SetArgs([]string{"testfiles/macos.yaml"})

		b := bytes.NewBufferString("")
		apply.SetOut(b)

		err := apply.Execute()
		require.NoError(t, err)

		out, err := ioutil.ReadAll(b)
		require.NoError(t, err, "require no error while reading buffer string")
		assert.Equal(t, want, string(out))
	})

	var macosYaml = `apiVersion: bindplane.observiq.com/v1beta
kind: Source
metadata:
    name: macOS
spec:
  plugin:
    name: macos
  parameters:
    - name: name
      value: macOS
    - name: version
      value: "0.0.2"
    - name: start_at
      value: end
    - name: enable_system_log
      value: true
    - name: enable_install_log
      value: true
`
	var malformedYaml = `apiVersion: bindplane.observiq.com/v1beta
kind Destination
metadata:
  name: cabin-production-logs
spec:
  plugin:
	name: [cabin_output]
  parameters:
  - name: endpoint
	value: https://nozzle.app.observiq.com
  - name: secret_key
	value: 2c088c5e-2afc-483b-be52-e2b657fcff08
  - name: 10
	value: 10s
`

	t.Run("can apply using '-' argument from stdin", func(t *testing.T) {
		apply := Command(stub)
		apply.SetArgs([]string{"-"})

		in := bytes.NewBufferString("")
		in.Write([]byte(macosYaml))
		apply.SetIn(in)

		out := bytes.NewBufferString("")
		apply.SetOut(out)

		err := apply.Execute()
		require.NoError(t, err)
	})

	t.Run("error using apply '-' argument from malfromed stdin", func(t *testing.T) {
		apply := Command(stub)
		apply.SetArgs([]string{"-"})

		in := bytes.NewBufferString("")
		in.Write([]byte(malformedYaml))
		apply.SetIn(in)

		out := bytes.NewBufferString("")
		apply.SetOut(out)

		err := apply.Execute()
		require.Error(t, err)
	})

	// TODO(jsirianni) decided if this is needed? https://github.com/observiq/bindplane/issues/246
	// t.Run("applies malformed spec yaml", func(t *testing.T) {
	// 	apply := Command(stub)
	// 	apply.SetArgs([]string{"-f", "testfiles/cabin-malformed-spec.yaml"})

	// 	b := bytes.NewBufferString("")
	// 	apply.SetOut(b)

	// 	err := apply.Execute()
	// 	require.Error(t, err)
	// })
}
