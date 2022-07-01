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

package initialize

import (
	"fmt"

	"github.com/AlecAivazis/survey/v2"
	"github.com/observiq/bindplane-op/internal/cli"
	"github.com/observiq/bindplane-op/internal/cli/commands/profile"
	"github.com/observiq/bindplane-op/model"
	"github.com/spf13/cobra"
)

// ClientCommand provides the implementation for "bindplane init client"
func ClientCommand(bindplane *cli.BindPlane, h profile.Helper) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "client",
		Aliases: []string{"cli"},
		Short:   "Initializes a new client installation",
		Long:    ``,
		RunE: func(cmd *cobra.Command, args []string) error {
			return modifyProfile(bindplane, h, func(spec *model.ProfileSpec) error {
				fmt.Printf("Please provide some basic configuration to initialize the client:\n")
				results := newClientInitOptions(spec)
				err := results.complete()
				if err != nil {
					return err
				}
				results.apply(spec)
				return nil
			})
		},
	}
	return cmd
}

// ----------------------------------------------------------------------
// survey for interactive prompts

type clientInitOptions struct {
	ServerURL string `survey:"server-url"`
	Username  string `survey:"username"`
	Password  string `survey:"password"`
}

func newClientInitOptions(spec *model.ProfileSpec) *clientInitOptions {
	c := &clientInitOptions{
		ServerURL: spec.BindPlaneURL(),
		Username:  spec.Username,
		Password:  spec.Password,
	}
	// fill with defaults
	if c.Username == "" {
		c.Username = "admin"
	}
	return c
}

func (s *clientInitOptions) apply(spec *model.ProfileSpec) {
	spec.ServerURL = s.ServerURL
	spec.Username = s.Username
	spec.Password = s.Password
}

func (s *clientInitOptions) clientQuestions() questions {
	return questions{
		{
			beforeText: "URL of the BindPlane OP server",
			Question: survey.Question{
				Name: "server-url",
				Prompt: &survey.Input{
					Message: "Server URL",
					Default: s.ServerURL,
				},
				Validate: survey.Required,
			},
		},
		{
			beforeText: "Login to access the BindPlane OP server",
			Question: survey.Question{
				Name: "username",
				Prompt: &survey.Input{
					Message: "Username",
					Default: s.Username,
				},
				Validate: survey.Required,
			},
		},
		{
			Question: survey.Question{
				Name: "password",
				Prompt: &survey.Password{
					Message: "Password",
				},
				Validate: survey.Required,
			},
		},
	}
}

func (s *clientInitOptions) complete() error {
	if err := s.clientQuestions().ask(s); err != nil {
		return err
	}
	fmt.Println("\nInitialization complete!\nRun \"bindplane version\" to test the login.")
	return nil
}
