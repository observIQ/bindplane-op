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

package delete

import (
	"bytes"
	"context"
	"fmt"
	"io/ioutil"
	"path"
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/observiq/bindplane-op/client"
	"github.com/observiq/bindplane-op/internal/cli"
	"github.com/observiq/bindplane-op/model"
)

func TestDeleteCommand(t *testing.T) {
	bindplane := &cli.BindPlane{}
	cmd := Command(bindplane)
	t.Run("returns cobra command", func(t *testing.T) {
		assert.NotNil(t, cmd)
		assert.IsType(t, &cobra.Command{}, cmd)
	})
}

type mockClient struct {
	client.BindPlane
	mock.Mock
}

func (m *mockClient) DeletePipeline(ctx context.Context, name string) error {
	args := m.Called(ctx, name)
	return args.Error(0)
}

func (m *mockClient) DeleteExporter(ctx context.Context, name string) error {
	args := m.Called(ctx, name)
	return args.Error(0)
}

func (m *mockClient) DeleteReceiver(ctx context.Context, name string) error {
	args := m.Called(ctx, name)
	return args.Error(0)
}

func (m *mockClient) DeleteConfiguration(ctx context.Context, name string) error {
	args := m.Called(ctx, name)
	return args.Error(0)
}

func (m *mockClient) Delete(ctx context.Context, resources []*model.AnyResource) ([]*model.AnyResourceStatus, error) {
	args := m.Called(ctx, resources)
	return args.Get(0).([]*model.AnyResourceStatus), args.Error(1)
}

type deleteReturn struct {
	deleted []*model.AnyResourceStatus
	err     error
}

func TestDeleteFile(t *testing.T) {
	tests := []struct {
		description      string
		args             []string
		expectError      bool
		expectOutput     string
		mockReturn       *deleteReturn
		expectHelpCalled bool
	}{
		{
			description:      "prints help when file argument not given",
			args:             make([]string, 0),
			mockReturn:       &deleteReturn{nil, nil},
			expectError:      false,
			expectOutput:     "",
			expectHelpCalled: true,
		},
		{
			description:      "error when deleting a malformed yaml",
			args:             []string{"-f", path.Join("testfiles", "source-malformed-yaml.yaml")},
			mockReturn:       &deleteReturn{nil, nil},
			expectError:      true,
			expectOutput:     "",
			expectHelpCalled: false,
		},
		{
			description: "error when client delete fails",
			args:        []string{"-f", path.Join("testfiles", "combined.yaml")},
			mockReturn: &deleteReturn{
				deleted: nil,
				err:     fmt.Errorf("client.Delete error")},
			expectOutput:     "",
			expectError:      true,
			expectHelpCalled: false,
		},
	}

	for _, test := range tests {
		t.Run(test.description, func(t *testing.T) {
			// set up mock client
			client := &mockClient{}
			stub := cli.NewBindPlaneForTesting()
			stub.SetClient(client)

			// create command with stub
			cmd := Command(stub)

			// setup
			var helpCalled bool
			cmd.SetHelpFunc(func(c *cobra.Command, s []string) {
				helpCalled = true
			})

			buffer := bytes.NewBufferString("")
			cmd.SetOut(buffer)

			cmd.SetArgs(test.args)

			client.On("Delete", mock.Anything, mock.Anything).Return(test.mockReturn.deleted, test.mockReturn.err)

			// execute and read out
			cmdError := cmd.Execute()
			out, err := ioutil.ReadAll(buffer)
			assert.NoError(t, err)

			assert.Equal(t, test.expectHelpCalled, helpCalled)
			if test.expectOutput != "" {
				assert.Equal(t, test.expectOutput, string(out))
			}
			assert.Equal(t, test.expectError, cmdError != nil)
		})
	}
}

func TestDeleteSubCommands(t *testing.T) {
	var tests = []struct {
		resourceType string
		mockFuncName string
	}{
		{
			resourceType: "configuration",
			mockFuncName: "DeleteConfiguration",
		},
	}

	for _, test := range tests {
		runDeleteTests(t, test.resourceType, test.mockFuncName)
	}
}

func runDeleteTests(t *testing.T, resourceType string, mockFuncName string) {
	t.Run(fmt.Sprintf("%s: error with no name arg", resourceType), func(t *testing.T) {
		client := &mockClient{}
		stub := cli.NewBindPlaneForTesting()
		stub.SetClient(client)

		cmd := Command(stub)
		cmd.SetArgs([]string{resourceType})
		err := cmd.Execute()
		require.Error(t, err, "expected error with no <name> argument.")
	})

	t.Run(fmt.Sprintf("error when %s fails", mockFuncName), func(t *testing.T) {
		client := &mockClient{}
		stub := cli.NewBindPlaneForTesting()
		stub.SetClient(client)
		client.On(mockFuncName, mock.Anything, mock.Anything).Return(fmt.Errorf("unexpected error"))

		cmd := Command(stub)
		cmd.SetArgs([]string{resourceType, "name"})

		err := cmd.Execute()
		require.Error(t, err, fmt.Sprintf("expected error when %s returns an error", mockFuncName))
	})

	t.Run(fmt.Sprintf("%s: prints message on successful deletion", resourceType), func(t *testing.T) {
		client := &mockClient{}
		stub := cli.NewBindPlaneForTesting()
		stub.SetClient(client)
		client.On(mockFuncName, mock.Anything, mock.Anything).Return(nil)

		cmd := Command(stub)

		b := bytes.NewBufferString("")
		cmd.SetOut(b)
		cmd.SetArgs([]string{resourceType, "name"})

		cmd.Execute()
		out, err := ioutil.ReadAll(b)
		require.NoError(t, err, "error while trying to read bytes from cmd out")

		want := fmt.Sprintf("Successfully deleted %s 'name'\n", resourceType)
		require.Equal(t, want, string(out))
	})
}
