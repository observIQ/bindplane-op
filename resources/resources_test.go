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

package resources

import (
	"io/ioutil"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/observiq/bindplane-op/model"
)

// tests confirm that all resources are valid

func fileResource[T model.Resource](t *testing.T, path string) T {
	resources, err := model.ResourcesFromFile(path)
	require.NoError(t, err)

	parsed, err := model.ParseResources(resources)
	require.NoError(t, err)
	require.Len(t, parsed, 1)

	configuration, ok := parsed[0].(T)
	require.True(t, ok)
	return configuration
}

func resourcePaths(t *testing.T, folder string) []string {
	files, err := ioutil.ReadDir(folder)
	require.NoError(t, err)

	result := make([]string, len(files))
	for i, file := range files {
		result[i] = filepath.Join(folder, file.Name())
	}

	return result
}

func TestValidateSourceTypes(t *testing.T) {
	paths := resourcePaths(t, "source-types")
	for _, path := range paths {
		t.Run(path, func(t *testing.T) {
			resource := fileResource[*model.SourceType](t, path)
			err := resource.Validate()
			require.NoError(t, err)
		})
	}
}

func TestValidateDestinationTypes(t *testing.T) {
	paths := resourcePaths(t, "destination-types")
	for _, path := range paths {
		t.Run(path, func(t *testing.T) {
			resource := fileResource[*model.DestinationType](t, path)
			err := resource.Validate()
			require.NoError(t, err)
		})
	}
}
