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
	"time"

	"github.com/observiq/bindplane-op/internal/util"
	"go.uber.org/zap"
)

// Versions TODO(doc)
type Versions interface {
	LatestVersionString() string
	LatestVersion() (*Version, error)
	Version(version string) (*Version, error)

	Artifact(artifactType ArtifactType, version *Version, platform string) Artifact
}

// VersionsSettings TODO(doc)
type VersionsSettings struct {
	Logger *zap.Logger
}

// The latest version cache is handled separately from the version cache because it just keeps the latest version in
// memory whether it was read from the filesystem cache or the agents client.
const (
	latestVersionCacheDuration = 1 * time.Minute
)

type versions struct {
	client        Client
	cache         Cache
	latestVersion util.Remember[Version]
	logger        *zap.Logger
}

var _ Versions = (*versions)(nil)

// NewVersions creates an implementation of Versions using the specified client, cache, and settings. To disable
// caching, pass nil for the Cache.
func NewVersions(client Client, cache Cache, settings VersionsSettings) Versions {
	if client == nil {
		client = &nopClient{}
	}
	if cache == nil {
		cache = &nopCache{}
	}
	return &versions{
		client:        client,
		cache:         cache,
		latestVersion: util.NewRemember[Version](latestVersionCacheDuration),
		logger:        settings.Logger,
	}
}

func (v *versions) LatestVersionString() string {
	version, err := v.LatestVersion()
	if err != nil {
		return ""
	}
	return version.Version
}

// LatestVersion returns the latest *Version.
func (v *versions) LatestVersion() (*Version, error) {
	version := VersionLatest

	// check if we have a remembered result
	if remembered := v.latestVersion.Get(); remembered != nil {
		return remembered, nil
	}

	// always check the server for the latest
	found, err := v.client.Version(version)
	if found == nil || err != nil {
		// on error, check the cache which may be outdated but allows installations to provide a cached latest version when
		// disconnected
		if cached := v.cache.Version(version); cached != nil {
			return cached, nil
		}
		return found, err
	}

	// cache it before returning
	if found != nil {
		v.cache.SaveLatestVersion(found)
		v.latestVersion.Update(found)
	}

	return found, nil
}

// Version returns the specified agent version. If the version is invalid or does not exist, it returns an error. If
// version is "latest", it returns the latest version.
func (v *versions) Version(version string) (*Version, error) {
	if version == VersionLatest {
		return v.LatestVersion()
	}

	// first try the cache
	if cached := v.cache.Version(version); cached != nil {
		return cached, nil
	}

	// not in the cache, download it
	found, err := v.client.Version(version)
	if err != nil {
		return nil, err
	}

	// cache it before returning
	if found != nil {
		v.cache.SaveVersion(found)
	}

	return found, nil
}

// Artifact returns an Artifact corresponding to the specified artifact type, version, and platform
func (v *versions) Artifact(artifactType ArtifactType, version *Version, platform string) Artifact {
	return v.client.Artifact(artifactType, version, platform)
}
