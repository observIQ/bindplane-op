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

package agent

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

var testCacheDirectory = filepath.Join("testdata", "cache")
var testCache = &cache{
	cacheDirectory: testCacheDirectory,
}
var emptyCacheDirectory = filepath.Join("testdata", "empty")
var emptyCache = &cache{
	cacheDirectory: emptyCacheDirectory,
}
var badCacheDirectory = filepath.Join("testdata", "does-not-exist")
var badCache = &cache{
	cacheDirectory: badCacheDirectory,
}

func TestCacheLatestVersionPath(t *testing.T) {
	require.Equal(t, filepath.Join(testCacheDirectory, "latest.json"), testCache.latestVersionPath())
}

func TestCacheLatestVersion(t *testing.T) {
	latest := testCache.LatestVersion()
	require.Equal(t, "2.0.6", latest.Version)
}

func TestCacheSpecificVersion(t *testing.T) {
	latest := testCache.Version("2.0.5")
	require.Equal(t, "2.0.5", latest.Version)
}

func TestCacheMissingArtifact(t *testing.T) {
	version := testCache.Version("2.0.5")
	a := testCache.Artifact(Download, version, "linux-arm64")
	require.False(t, a.Exists())
}

func TestCachePresentArtifact(t *testing.T) {
	version := testCache.Version("2.0.5")
	a := testCache.Artifact(Installer, version, "darwin-arm64")
	require.True(t, a.Exists())
	bytes, err := a.Read()
	require.NoError(t, err)
	require.Equal(t, string(bytes), "#!/bin/sh\n# fake test installer\n")
}

func TestCacheSaveArtifact(t *testing.T) {
	version := testCache.Version("2.0.5")
	platform := "linux-arm64"
	defer func() {
		path := testCache.artifactPath(Installer, version, platform)
		parent, _ := filepath.Split(path)
		err := os.RemoveAll(parent)
		require.NoError(t, err)
	}()

	a := testCache.Artifact(Installer, version, platform)
	require.False(t, a.Exists())

	// save a new artifact
	err := a.Write([]byte("test"))
	require.NoError(t, err)

	// confirm it exists and loads
	require.True(t, a.Exists())
	bytes, err := a.Read()
	require.NoError(t, err)
	require.Equal(t, "test", string(bytes))
	require.Equal(t, "observiq-agent-installer.txt", a.Name())
}

func TestCacheArtifactPath(t *testing.T) {
	version := testCache.Version("2.0.5")
	tests := []struct {
		artifactKey ArtifactType
		platform    string
		path        string
	}{
		{
			artifactKey: Installer,
			platform:    "linux-arm64",
			path:        "testdata/cache/2.0.5/linux-arm64/observiq-agent-installer.txt",
		},
		{
			artifactKey: "bad-key",
			platform:    "linux-arm64",
			path:        "",
		},
		{
			artifactKey: Installer,
			platform:    "bad-platform",
			path:        "",
		},
	}
	for i, test := range tests {
		t.Run(fmt.Sprint(i), func(t *testing.T) {
			p := testCache.artifactPath(test.artifactKey, version, test.platform)
			require.Equal(t, test.path, p)
		})
	}
}

func TestCacheEmptyLatestVersion(t *testing.T) {
	latest := emptyCache.LatestVersion()
	require.Nil(t, latest)
}
func TestCacheEmptyVersion(t *testing.T) {
	version := emptyCache.Version("2.0.5")
	require.Nil(t, version)
}

func TestCacheWriteBadPath(t *testing.T) {
	artifact := cacheArtifact{path: "//bad:path://does//notexist", logger: zap.NewNop()}
	writer := artifact.Writer()
	defer writer.Close()
	bytes := []byte("will not be written")
	l, err := writer.Write(bytes)
	require.NoError(t, err)
	require.Equal(t, len(bytes), l)
}

func TestBadCacheLatestVersion(t *testing.T) {
	latest := badCache.LatestVersion()
	require.Nil(t, latest)
}
func TestBadCacheEmptyVersion(t *testing.T) {
	version := badCache.Version("2.0.5")
	require.Nil(t, version)
}
