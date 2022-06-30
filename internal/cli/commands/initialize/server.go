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
	"os"

	"github.com/AlecAivazis/survey/v2"
	"github.com/AlecAivazis/survey/v2/terminal"
	"github.com/google/uuid"
	"github.com/observiq/bindplane/common"
	"github.com/observiq/bindplane/internal/cli"
	"github.com/observiq/bindplane/internal/cli/commands/profile"
	"github.com/observiq/bindplane/model"
	"github.com/spf13/cobra"
)

// ServerCommand provides the implementation for "bindplane init server"
func ServerCommand(bindplane *cli.BindPlane, h profile.Helper) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "server",
		Aliases: []string{"serve"},
		Short:   "Initializes a new server installation",
		Long:    ``,
		RunE: func(cmd *cobra.Command, args []string) error {
			return modifyProfile(bindplane, h, func(spec *model.ProfileSpec) error {
				fmt.Printf("Please provide some basic configuration to initialize the server:\n")
				results := newServerInitOptions(spec)
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

type serverInitOptions struct {
	spec           *model.ProfileSpec
	Host           string `survey:"host"`
	Port           string `survey:"port"`
	ServerURL      string `survey:"server-url"`
	RemoteURL      string `survey:"remote-url"`
	SecretKey      string `survey:"secret-key"`
	SessionsSecret string `survey:"sessions-secret"`
	Username       string `survey:"username"`
	Password       string
}

func newServerInitOptions(spec *model.ProfileSpec) *serverInitOptions {
	// fill any values already specified by the spec
	s := &serverInitOptions{
		spec:      spec,
		Host:      spec.Host,
		Port:      spec.Port,
		ServerURL: spec.Server.ServerURL,
		RemoteURL: spec.Server.RemoteURL,
		Username:  spec.Username,
		Password:  spec.Password,
		SecretKey: spec.Server.SecretKey,
	}

	// fill with defaults
	if s.Host == "" {
		s.Host = "127.0.0.1"
	}
	if s.Port == "" {
		s.Port = "3001"
	}
	if s.Username == "" {
		s.Username = "admin"
	}
	if s.SecretKey == "" {
		s.SecretKey = uuid.NewString()
	}
	if s.SessionsSecret == "" {
		s.SessionsSecret = uuid.NewString()
	}

	return s
}

func (s *serverInitOptions) apply(spec *model.ProfileSpec) {
	spec.Host = s.Host
	spec.Port = s.Port
	spec.Username = s.Username
	// blank password to preserve existing
	if s.Password != "" {
		spec.Password = s.Password
	}
	spec.ServerURL = s.ServerURL
	spec.Server.RemoteURL = s.RemoteURL
	spec.Server.SecretKey = s.SecretKey
	spec.Server.SessionsSecret = s.SessionsSecret
}

func (s *serverInitOptions) serverHostQuestions() questions {
	// Note: defaults are based on the existing values, which are created in newSurveyResults if necessary.
	return questions{
		{
			beforeText: "The IP address the BindPlane server should listen on.\nSet to 0.0.0.0 to listen on all IP addresses.",
			Question: survey.Question{
				Name: "host",
				Prompt: &survey.Input{
					Message: "Server Host",
					Help:    "Bind Address for the HTTP server or 0.0.0.0 to bind to all network interfaces",
					Default: s.Host,
				},
				Validate: survey.Required,
			},
		},
		{
			beforeText: "The TCP port BindPlane should bind to.\nAll communication to the BindPlane server (HTTP, GraphQL, WebSocket) will use this port.",
			Question: survey.Question{
				Name: "port",
				Prompt: &survey.Input{
					Message: "Server Port",
					Help:    "Port for the HTTP server",
					Default: s.Port,
				},
				Validate: survey.Required,
			},
		},
	}
}

func (s *serverInitOptions) serverQuestions() questions {
	return questions{
		{
			beforeText: "The full HTTP URL used for communication between client and server.\nUse the IP address or hostname of the server, starting with http:// for plain text or https:// for TLS.",
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
			beforeText: "The full WebSocket URL used for remote management of agents.\nUse the IP address or hostname of the server, starting with ws:// for plain text or wss:// for TLS.",
			Question: survey.Question{
				Name: "remote-url",
				Prompt: &survey.Input{
					Message: "Remote URL",
					Default: s.RemoteURL,
				},
				Validate: survey.Required,
			},
		},
		{
			beforeText: "Choose a secret key to be used for authentication between server and agents.",
			Question: survey.Question{
				Name: "secret-key",
				Prompt: &survey.Input{
					Message: "Secret Key",
					Default: s.SecretKey,
				},
				Validate: survey.Required,
			},
		},
		{
			beforeText: "Choose a secret key to be used to encode user session cookies.  Must be a uuid.",
			Question: survey.Question{
				Name: "sessions-secret",
				Prompt: &survey.Input{
					Message: "Sessions Secret",
					Default: s.SessionsSecret,
				},
				Validate: validateUUID,
			},
		},
		{
			beforeText: "Specify a username and password to restrict access to the server.",
			Question: survey.Question{
				Name: "username",
				Prompt: &survey.Input{
					Message: "Username",
					Default: s.Username,
				},
				Validate: survey.Required,
			},
		},
	}
}

func (s *serverInitOptions) complete() error {
	if err := s.serverHostQuestions().ask(s); err != nil {
		return err
	}

	s.regenerateURLs()

	if err := s.serverQuestions().ask(s); err != nil {
		return err
	}

	err := s.completePassword()
	if err != nil {
		return err
	}

	fmt.Println("\nInitialization complete!\nRestart the BindPlane server to reload the configuration.")

	return nil
}

func (s *serverInitOptions) regenerateURLs() {
	c := common.Server{
		Common: common.Common{
			Host: s.Host,
			Port: s.Port,
		},
	}

	hostChanged := s.Host != s.spec.Host
	portChanged := s.Port != s.spec.Port

	// apply Host and Port so that we can use them to generate server and remote URLs. use a hostname if host is 0.0.0.0
	// since that is not addressable by another system.
	useHostName := s.Host == "0.0.0.0"
	if useHostName {
		if hostname, err := os.Hostname(); err == nil {
			c.Common.Host = hostname
		}
	}

	if portChanged || hostChanged || s.ServerURL == "" {
		s.ServerURL = c.BindPlaneURL()
	}
	if portChanged || hostChanged || s.RemoteURL == "" {
		s.RemoteURL = c.WebsocketURL()
	}
}

func (s *serverInitOptions) completePassword() error {
	// password prompting is more complicated
	//
	// 1. If password exists, blank will preserve
	//
	// 2. If password does not exist, blank is not allowed
	//
	password := s.Password
	passwordOpts := []survey.AskOpt{survey.WithValidator(survey.Required)}
	passwordPrompt := "Password (must not be empty)"
	if s.Password != "" {
		passwordPrompt = "Password (blank will preserve the current password)"
		passwordOpts = nil
	}
	for {
		err := survey.AskOne(&survey.Password{
			Message: passwordPrompt,
		}, &password, passwordOpts...)
		if err == terminal.InterruptErr {
			return err
		}

		// blank preserves
		if password == "" {
			break
		}

		// confirm the password
		confirm := ""
		err = survey.AskOne(&survey.Password{
			Message: "Enter the password again to confirm",
		}, &confirm)
		if err == terminal.InterruptErr {
			return err
		}

		if password == confirm {
			break
		}

		// inform that they don't match
		fmt.Printf("Passwords did not match\n")
	}
	s.Password = password
	return nil
}

func validateUUID(ans interface{}) error {
	_, err := uuid.Parse(ans.(string))
	return err
}
