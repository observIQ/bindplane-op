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

package commands

import (
	"os"
	"path"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/observiq/bindplane-op/common"
	"github.com/observiq/bindplane-op/internal/cli"
	"github.com/observiq/bindplane-op/internal/cli/commands/profile"
	"github.com/observiq/bindplane-op/internal/cli/printer"
)

func TestGetAgents(t *testing.T) {
	t.Run("sets up a TablePrinter when no output option is given", func(t *testing.T) {
		bindplane := cli.NewBindPlane(common.InitConfig(""), os.Stdout)
		cmd := Command(bindplane, "bindplane")
		cmd.Execute()
		err := initViper(bindplane.Config)
		require.NoError(t, err)

		require.IsType(t, &printer.TablePrinter{}, bindplane.Printer())
	})

	t.Run("sets up a TablePrinter when output option is set to 'table'", func(t *testing.T) {
		bindplane := cli.NewBindPlane(common.InitConfig(""), os.Stdout)
		cmd := Command(bindplane, "bindplane")
		cmd.SetArgs([]string{"--output", "table"})
		cmd.Execute()
		err := initViper(bindplane.Config)
		require.NoError(t, err)

		require.IsType(t, &printer.TablePrinter{}, bindplane.Printer())
	})

	t.Run("sets up a JSONPrinter when output option is set to 'json'", func(t *testing.T) {
		bindplane := cli.NewBindPlane(common.InitConfig(""), os.Stdout)
		cmd := Command(bindplane, "bindplane")
		cmd.SetArgs([]string{"--output", "json"})
		cmd.Execute()
		err := initViper(bindplane.Config)
		require.NoError(t, err)

		require.IsType(t, &printer.JSONPrinter{}, bindplane.Printer())
	})

	t.Run("sets up a YamlPrinter when output option is set to 'yaml'", func(t *testing.T) {
		bindplane := cli.NewBindPlane(common.InitConfig(""), os.Stdout)
		cmd := Command(bindplane, "bindplane")
		cmd.SetArgs([]string{"--output", "yaml"})
		cmd.Execute()
		err := initViper(bindplane.Config)
		require.NoError(t, err)

		require.IsType(t, &printer.YamlPrinter{}, bindplane.Printer())
	})
}

func TestRootConfigFilePath(t *testing.T) {
	tests := []struct {
		name    string
		folder  string
		config  string
		profile string
		expect  string
		err     bool
	}{
		{
			name:   "defaults to current profile",
			folder: "profiles",
			expect: path.Join("profile", "testfiles", "profiles", "local.yaml"),
		},
		{
			name:   "config flag overrides current profile",
			folder: "profiles",
			config: "something.yaml",
			expect: "something.yaml",
		},
		{
			name:    "profile flag overrides current profile",
			folder:  "profiles",
			profile: "mindplane",
			expect:  path.Join("profile", "testfiles", "profiles", "mindplane.yaml"),
		},
		{
			name:    "config flag overrides profile flag overrides current profile",
			folder:  "profiles",
			config:  "something.yaml",
			profile: "mindplane",
			expect:  "something.yaml",
		},
		{
			name:    "profile flag value must exist",
			folder:  "profiles",
			profile: "does-not-exist",
			err:     true,
		},
		// ----------------------------------------------------------------------
		{
			name:   "defaults to current profile",
			folder: "does-not-exist",
			expect: "",
		},
		{
			name:   "config flag overrides current profile",
			folder: "does-not-exist",
			config: "something.yaml",
			expect: "something.yaml",
		},
		{
			name:    "profile flag overrides current profile",
			folder:  "does-not-exist",
			profile: "mindplane",
			expect:  "",
			err:     true,
		},
		{
			name:    "config flag overrides profile flag overrides current profile",
			folder:  "does-not-exist",
			config:  "something.yaml",
			profile: "mindplane",
			expect:  "something.yaml",
		},
		{
			name:    "profile flag value must exist",
			folder:  "does-not-exist",
			profile: "does-not-exist",
			err:     true,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			f := profile.LoadFolder(path.Join("profile", "testfiles", test.folder))
			path, err := configFilePath(f, test.config, test.profile)
			require.Equal(t, test.expect, path)
			if test.err {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}
