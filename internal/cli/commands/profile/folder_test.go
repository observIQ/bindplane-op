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
	"os"
	"path"
	"testing"

	"github.com/observiq/bindplane-op/common"
	"github.com/observiq/bindplane-op/model"
	"github.com/stretchr/testify/require"
)

var profilesFolderPath = path.Join("testfiles", "profiles")

func TestCurrentProfile(t *testing.T) {
	f := LoadFolder(profilesFolderPath)
	name, err := f.CurrentProfileName()
	require.NoError(t, err)
	require.Equal(t, "local", name)
	p, err := f.CurrentProfilePath()
	require.NoError(t, err)
	require.Equal(t, path.Join(profilesFolderPath, "local.yaml"), p)
}

func TestProfileNames(t *testing.T) {
	f := LoadFolder(path.Join("testfiles", "profiles"))
	names := f.ProfileNames()
	require.ElementsMatch(t, []string{"local", "mindplane"}, names)
}

func TestReadProfile(t *testing.T) {
	f := LoadFolder(path.Join("testfiles", "profiles"))
	tests := []struct {
		name string
		port string
		err  bool
	}{
		{
			name: "local",
			port: "3001",
			err:  false,
		},
		{
			name: "mindplane",
			port: "80",
			err:  false,
		},
		{
			name: "does-not-exist",
			port: "",
			err:  true,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			profile, err := f.ReadProfile(test.name)
			if test.err {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				require.Equal(t, test.port, profile.Spec.Port)
			}
		})
	}
}

func forTestingProfile() *model.Profile {
	name := "for-testing"
	spec := model.ProfileSpec{
		Common: common.Common{
			Port: "9999",
		},
	}
	return model.NewProfile(name, spec)
}

func TestWriteProfile(t *testing.T) {
	f := LoadFolder(path.Join("testfiles", "profiles"))
	profile := forTestingProfile()
	name := profile.Name()
	defer func() {
		f.RemoveProfile(name)
	}()
	err := f.WriteProfile(profile)
	require.NoError(t, err)
	require.True(t, f.ProfileExists(name))

	actual, err := f.ReadProfile(name)
	require.NoError(t, err)
	require.Equal(t, "9999", actual.Spec.Common.Port)
}

func TestFolderCreation(t *testing.T) {
	f := LoadFolder(path.Join("testfiles", "for-testing-profiles"))
	profile := forTestingProfile()
	name := profile.Name()
	defer func() {
		os.RemoveAll(f.ProfilesFolderPath())
	}()
	err := f.WriteProfile(profile)
	require.NoError(t, err)
	require.True(t, f.ProfileExists(name))

	actual, err := f.ReadProfile(name)
	require.NoError(t, err)
	require.Equal(t, "9999", actual.Spec.Common.Port)

	current, err := f.CurrentProfileName()
	require.NoError(t, err)
	require.Equal(t, name, current)
}

func TestLoadFolderError(t *testing.T) {
	f := LoadFolder(path.Join("testfiles", "profiles", "local.yaml"))
	err := f.(*folder).ensureProfilesFolderExists()
	require.Error(t, err)
}

func TestCurrentProfileNameError(t *testing.T) {
	f := LoadFolder(path.Join("testfiles", "profiles-error-current-folder"))
	_, err := f.CurrentProfileName()
	require.Error(t, err)
	_, err = f.CurrentProfilePath()
	require.Error(t, err)
}

func TestProfileBadParseError(t *testing.T) {
	f := LoadFolder(path.Join("testfiles", "profiles-error-bad-yaml"))

	name, err := f.CurrentProfileName()
	require.Error(t, err, "name %s", name)

	_, err = f.ReadProfile("malformed")
	require.Error(t, err)
}

func TestSetCurrentProfileName(t *testing.T) {
	f := LoadFolder(path.Join("testfiles", "profiles"))
	defer func() {
		_ = f.SetCurrentProfileName("local")
	}()

	// warning: a test failure in this test could modify the testfiles

	err := f.SetCurrentProfileName("mindplane")
	require.NoError(t, err)
	name, err := f.CurrentProfileName()
	require.NoError(t, err)
	require.Equal(t, "mindplane", name)
}
