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
	"io/ioutil"
	"os"
	"path"
	"strings"

	"github.com/observiq/bindplane-op/model"
	"gopkg.in/yaml.v3"
)

// DefaultName is the default profile name
const DefaultName string = "default"

// Folder manages the folder of profiles. There is a .yaml file for each profile with the name of the profile and there
// is a current file which is also yaml and contains a single `name: name` entry with the name of the current
// profile.
type Folder interface {
	// Exists returns true if the profiles folder exists
	Exists() bool

	// ProfilesFolderPath returns the path to the folder of profiles which defaults to ~/.bindplane/profiles
	ProfilesFolderPath() string

	// CurrentProfilePath returns the full path of the current profile .yaml file
	CurrentProfilePath() (string, error)

	// CurrentProfileName returns the name of the current profile
	CurrentProfileName() (string, error)

	// SetCurrentProfileName changes the current profile name to a profile with the specified name. It returns an error if
	// a profile with that name doesn't exist or if the current profile could not be written.
	SetCurrentProfileName(name string) error

	// ProfileNames returns the names of the profiles. If we are unable to read the folder of profiles, an empty slice is
	// returned because there are no profiles that can be read.
	ProfileNames() []string

	// ProfilePath returns true if the profile file exists
	ProfilePath(name string) string

	// ProfileExists returns true if the profile file exists
	ProfileExists(name string) bool

	// ReadProfile reads and returns the profile with the specified name
	ReadProfile(name string) (*model.Profile, error)

	// WriteProfile writes a new profile with the specified name, overwriting an existing profiles with the same name. If
	// there is no current profile name, this profile will be set as the current profile.
	WriteProfile(profile *model.Profile) error

	// UpsertProfile creates or modifies a profile using ReadProfile and WriteProfile
	UpsertProfile(name string, updater func(*model.Profile) error) error

	// RemoveProfile deletes a profile with the specified name. If the profile doesn't exist, it does nothing.
	RemoveProfile(name string) error
}

type folder struct {
	profilesFolderPath string
}

var _ Folder = (*folder)(nil)

// LoadFolder returns an implementation of the profile.Folder interface and after ensuring that the folder exists and
// attempting to create it if necessary.
func LoadFolder(profilesFolderPath string) Folder {
	return &folder{profilesFolderPath: profilesFolderPath}
}

func (f *folder) ProfilesFolderPath() string {
	return f.profilesFolderPath
}

func (f *folder) CurrentProfilePath() (string, error) {
	name, err := f.CurrentProfileName()
	if err != nil {
		return "", err
	}
	return f.ProfilePath(name), nil
}

// currentProfileName comes from the ~/.bindplane/profiles/current file
func (f *folder) CurrentProfileName() (string, error) {
	filename := f.currentFilePath()
	bytes, err := ioutil.ReadFile(path.Clean(filename))
	if err != nil {
		return "", err
	}

	var current current
	if err = yaml.Unmarshal(bytes, &current); err != nil {
		return "", err
	}
	if current.Name == "" {
		return "", fmt.Errorf("unable to read the current profile")
	}
	return current.Name, nil
}

// SetCurrentProfileName changes the current profile name to a profile with the specified name. It returns an error if
// a profile with that name doesn't exist or if the current profile could not be written.
func (f *folder) SetCurrentProfileName(name string) error {
	// ensure it exists
	if !f.ProfileExists(name) {
		return fmt.Errorf("no profile found with name '%s'", name)
	}

	if err := f.ensureProfilesFolderExists(); err != nil {
		return err
	}

	current := current{Name: name}
	bytes, err := yaml.Marshal(current)
	if err != nil {
		return err
	}

	filename := path.Join(f.profilesFolderPath, "current")
	err = ioutil.WriteFile(filename, bytes, 0600)
	if err != nil {
		return err
	}

	return nil
}

// ProfileNames returns the names of the profiles. If we are unable to read the folder of profiles, an empty slice is
// returned because there are no profiles that can be read.
func (f *folder) ProfileNames() []string {
	files, err := ioutil.ReadDir(f.profilesFolderPath)
	if err != nil {
		return []string{}
	}
	names := []string{}
	for _, file := range files {
		if strings.HasSuffix(file.Name(), ".yaml") {
			name := strings.TrimSuffix(file.Name(), ".yaml")
			names = append(names, name)
		}
	}
	return names
}

// ReadProfile reads and returns the profile with the specified name
func (f *folder) ReadProfile(name string) (*model.Profile, error) {
	file := f.ProfilePath(name)
	bytes, err := ioutil.ReadFile(path.Clean(file))
	if err != nil {
		return nil, err
	}
	var spec model.ProfileSpec
	if err = yaml.Unmarshal(bytes, &spec); err != nil {
		return nil, err
	}
	return model.NewProfile(name, spec), nil
}

// WriteProfile writes a new profile with the specified name, overwriting an existing profiles with the same name
func (f *folder) WriteProfile(profile *model.Profile) error {
	if err := f.ensureProfilesFolderExists(); err != nil {
		return err
	}
	bytes, err := yaml.Marshal(profile.Spec)
	if err != nil {
		return err
	}
	file := f.ProfilePath(profile.Name())
	if err = ioutil.WriteFile(file, bytes, 0600); err != nil {
		return err
	}
	name, err := f.CurrentProfileName()
	if name == "" || err != nil {
		return f.SetCurrentProfileName(profile.Name())
	}
	return nil
}

// UpsertProfile creates or modifies a profile using ReadProfile and WriteProfile
func (f *folder) UpsertProfile(name string, updater func(*model.Profile) error) error {
	profile, err := f.ReadProfile(name)
	if err != nil {
		profile = model.NewProfile(name, model.ProfileSpec{})
	}
	if err := updater(profile); err != nil {
		return err
	}
	return f.WriteProfile(profile)
}

// RemoveProfile deletes a profile with the specified name
func (f *folder) RemoveProfile(name string) error {
	filename := f.ProfilePath(name)
	return os.Remove(filename)
}

// profileExists returns true if the specified profile file name exists
func (f *folder) ProfileExists(name string) bool {
	filename := f.ProfilePath(name)
	info, err := os.Stat(filename)
	return err == nil && !info.IsDir()
}

// ProfilePath returns the path to a profile with the specified name
func (f *folder) ProfilePath(name string) string {
	return path.Join(f.profilesFolderPath, fmt.Sprintf("%s.yaml", name))
}

// Exists returns true if the profiles folder exists
func (f *folder) Exists() bool {
	info, err := os.Stat(f.profilesFolderPath)
	return err == nil && info.IsDir()
}

func (f *folder) ensureProfilesFolderExists() error {
	info, err := os.Stat(f.profilesFolderPath)
	if err == nil {
		if info.IsDir() {
			return nil
		}
		return fmt.Errorf("profile folder name %s exists but is not a directory", f.profilesFolderPath)
	}
	if errors.Is(err, os.ErrNotExist) {
		return os.MkdirAll(f.profilesFolderPath, 0750)
	}
	return nil
}

func (f *folder) currentFilePath() string {
	return path.Join(f.profilesFolderPath, "current")
}

// ----------------------------------------------------------------------
type current struct {
	Name string
}
