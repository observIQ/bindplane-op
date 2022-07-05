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
	"errors"
	"fmt"
	"os"
	"path"

	"github.com/observiq/bindplane-op/common"
)

// Helper TODO(doc)
type Helper interface {
	Folder() Folder

	Directory() string
	directoryExists() (bool, error)
	mkDir() error

	HomeFolderSetup() error
}

// implements the ConfigHelper interface
type helper struct {
	bindplaneHomePath string

	folder Folder
}

var _ Helper = (*helper)(nil)

// NewHelper TODO(doc)
func NewHelper(bindplaneHomePath string) Helper {
	h := &helper{
		bindplaneHomePath: bindplaneHomePath,
	}
	h.folder = LoadFolder(h.profilesFolderPath())
	return h
}

// BindPlaneProfilesFolderPath returns the path to the folder where individual configuration profiles are stored
func (h *helper) profilesFolderPath() string {
	return path.Join(h.bindplaneHomePath, common.ProfilesFolderName)
}

func (h *helper) Folder() Folder {
	return h.folder
}

// Directory is the bindplane home path, defaulting to $HOME/.bindplane
func (h *helper) Directory() string {
	return h.bindplaneHomePath
}

// true indicates the config directory .bindplane exists in the user home directory
func (h *helper) directoryExists() (bool, error) {
	d := h.Directory()

	_, err := os.Stat(d)
	if err == nil {
		return true, nil
	}

	if errors.Is(err, os.ErrNotExist) {
		return false, nil
	}

	return false, fmt.Errorf("unexpected error when verifying if directory exists: %s, %w", h.Directory(), err)
}

// creates the directory .bindplane in the user home path
func (h *helper) mkDir() error {
	d := h.Directory()

	err := os.Mkdir(d, 0750)
	return err
}

// Makes sure that there is an existing config directory and file to work with
func (h *helper) HomeFolderSetup() error {
	// Check if config directory dirExists
	dirExists, err := h.directoryExists()
	if err != nil {
		return fmt.Errorf("error when trying to verify config directory exists: %s", h.Directory())
	}

	if !dirExists {
		err = h.mkDir()
		if err != nil {
			return err
		}

		return nil
	}

	return nil
}
