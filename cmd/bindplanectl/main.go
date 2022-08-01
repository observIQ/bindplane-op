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

package main

import (
	"fmt"
	"os"

	"github.com/observiq/bindplane-op/common"
	"github.com/observiq/bindplane-op/internal/cli"
	"github.com/observiq/bindplane-op/internal/cli/commands"
	"github.com/observiq/bindplane-op/internal/cli/commands/apply"
	"github.com/observiq/bindplane-op/internal/cli/commands/copy"
	"github.com/observiq/bindplane-op/internal/cli/commands/delete"
	"github.com/observiq/bindplane-op/internal/cli/commands/get"
	"github.com/observiq/bindplane-op/internal/cli/commands/initialize"
	"github.com/observiq/bindplane-op/internal/cli/commands/install"
	"github.com/observiq/bindplane-op/internal/cli/commands/label"
	"github.com/observiq/bindplane-op/internal/cli/commands/profile"
	"github.com/observiq/bindplane-op/internal/cli/commands/validate"
	"github.com/observiq/bindplane-op/internal/cli/commands/version"
	"github.com/spf13/cobra"
)

func main() {
	home := commands.BindplaneHome()

	var h = profile.NewHelper(home)

	// We need to perform this before creating a new bindplane cli because bindplane cli
	// creates a new logger with a file in ~/.bindplane
	err := h.HomeFolderSetup()
	if err != nil {
		fmt.Printf("error while trying to set up BindPlane home directory %s, %s\n", home, err.Error())
		os.Exit(1)
	}

	// Initialize the BindPlane CLI
	bindplaneConfig := common.InitConfig(home)
	bindplane := cli.NewBindPlane(bindplaneConfig, os.Stdout)

	rootCmd := commands.Command(bindplane, "bindplanectl")

	// Client does not contain serve command
	rootCmd.AddCommand(
		apply.Command(bindplane),
		get.Command(bindplane),
		label.Command(bindplane),
		delete.Command(bindplane),
		profile.Command(h),
		version.Command(bindplane),
		initialize.Command(bindplane, h, initialize.ClientMode),
		install.Command(bindplane),
		validate.Command(bindplane),
		copy.Command(bindplane),
	)

	cobra.CheckErr(rootCmd.Execute())
	bindplane.Shutdown()
}
