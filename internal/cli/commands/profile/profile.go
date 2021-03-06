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

package profile

import (
	"github.com/spf13/cobra"
)

// Command returns the BindPlane profile cobra command.
func Command(h Helper) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "profile",
		Short: "Profile commands.",
		Long:  "Profile commands for managing BindPlane application configuration",
	}

	cmd.AddCommand(
		GetCommand(h),
		SetCommand(h),
		DeleteCommand(h),
		ListCommand(h),
		UseCommand(h),
		CurrentCommand(h),
	)

	return cmd
}
