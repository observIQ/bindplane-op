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
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/AlecAivazis/survey/v2"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"

	"github.com/observiq/bindplane-op/common"
	"github.com/observiq/bindplane-op/internal/cli"
	"github.com/observiq/bindplane-op/internal/cli/commands/profile"
	"github.com/observiq/bindplane-op/model"
)

// Mode defines which initialize commands should be supported
type Mode string

var (
	// ServerMode enables the server initialize command
	ServerMode Mode = "server"
	// ClientMode enables the client initialize command
	ClientMode Mode = "client"
	// DualMode enables both client and server initialize commands
	DualMode Mode = "dual"
)

// Command returns the BindPlane initialize cobra command
func Command(bindplane *cli.BindPlane, h profile.Helper, mode Mode) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "init",
		Aliases: []string{"initialize"},
		Short:   "Initialize an installation",
	}

	if mode == ClientMode || mode == DualMode {
		cmd.AddCommand(
			ClientCommand(bindplane, h),
		)
	}

	if mode == ServerMode || mode == DualMode {
		cmd.AddCommand(
			ServerCommand(bindplane, h),
		)
	}

	return cmd
}

// ProfileUpdater handles the actual updates to a profile
type ProfileUpdater func(spec *model.ProfileSpec) error

// ----------------------------------------------------------------------

// modifyProfile handles load/save of profiles and is shared by client and server initialization
func modifyProfile(bindplane *cli.BindPlane, h profile.Helper, updater ProfileUpdater) error {
	profiles := h.Folder()

	// this shares a lot of behavior with "bindplane profile set" but supports interactive prompts. we handle three
	// cases:
	//
	// 1. no config file, no profile name: create a new "default" profile and set as current
	//
	// 2. config file, no profile name: modify the config file as if it were a profile, but save back to file
	//
	// 3. config file, profile name: modify the profile and save back to file
	//
	// Cases 2 & 3 can be handled the same way

	newCurrentProfileName := "" // set if we need to change the current profile
	specPath := bindplane.ConfigFile

	if bindplane.ConfigFile == "" {
		// case 1, uninitialized, create default
		specPath = profiles.ProfilePath(common.DefaultProfileName)
		newCurrentProfileName = common.DefaultProfileName
	}

	spec, err := loadProfileSpec(specPath)
	if err != nil {
		return err
	}

	err = updater(spec)
	if err != nil {
		return err
	}

	// make sure we set the current profile
	if newCurrentProfileName != "" {
		profile := model.NewProfile(newCurrentProfileName, *spec)
		if err := profiles.WriteProfile(profile); err != nil {
			return err
		}
		if err := profiles.SetCurrentProfileName(newCurrentProfileName); err != nil {
			return err
		}
	}

	// save the changes
	if err = saveProfileSpec(specPath, spec); err != nil {
		return err
	}

	return nil
}

// ----------------------------------------------------------------------
// load/save profile spec (configuration files)

func loadProfileSpec(path string) (*model.ProfileSpec, error) {
	spec := &model.ProfileSpec{}
	bytes, err := ioutil.ReadFile(filepath.Clean(path))
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			// if the file doesn't exist, we'll just return a new spec
			return spec, nil
		}
		return nil, err
	}
	if err = yaml.Unmarshal(bytes, spec); err != nil {
		return nil, err
	}
	return spec, nil
}

func saveProfileSpec(path string, spec *model.ProfileSpec) error {
	bytes, err := yaml.Marshal(spec)
	if err != nil {
		return err
	}
	if err = ioutil.WriteFile(path, bytes, 0600); err != nil {
		return err
	}
	return nil
}

// ----------------------------------------------------------------------
// questions

type question struct {
	beforeText string
	survey.Question
}

func (q *question) questions() []*survey.Question {
	return []*survey.Question{&q.Question}
}

type questions []*question

func (qs questions) ask(response interface{}) error {
	for _, q := range qs {
		if q.beforeText != "" {
			fmt.Println()
			fmt.Println(q.beforeText)
		}
		if err := survey.Ask(q.questions(), response); err != nil {
			return err
		}
	}
	return nil
}
