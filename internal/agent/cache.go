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
	"encoding/json"
	"io"
	"io/ioutil"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	"go.uber.org/zap"
)

// Cache implements a cache of agent versions.
//
// Certain operations of the cache may fail and any errors will be logged.
//
// For read operations that encounter an error, a nil pointer will be returned indicating that the item is not cached.
// To a caller there is no difference between an item that is not cached and an item that cannot be read from the cache
// due to an error. In both cases it is not available in the cache.
//
// For write operations that encounter an error, the write operation will do nothing. This allows the caller to make a
// best effort to save items to the cache but ignore failure.
type Cache interface {
	// Enabled TODO(docs)
	Enabled() bool

	// LatestVersion TODO(docs)
	LatestVersion() *Version
	// SaveLatestVersion TODO(docs)
	SaveLatestVersion(*Version)

	// Version TODO(docs)
	Version(string) *Version
	// SaveVersion TODO(docs)
	SaveVersion(*Version)

	// Artifact
	Artifact(artifactType ArtifactType, version *Version, platform string) CacheArtifact
}

// CacheArtifact TODO(docs)
type CacheArtifact interface {
	Name() string
	Path() string
	Exists() bool
	Read() ([]byte, error)
	Reader() (io.ReadCloser, error)
	Write([]byte) error
	Writer() io.WriteCloser
}

// CacheSettings TODO(docs)
type CacheSettings struct {
	Directory string
	Logger    *zap.Logger
}

// ----------------------------------------------------------------------

type cache struct {
	cacheDirectory string
	logger         *zap.Logger
}

var _ Cache = (*cache)(nil)

// NewCache takes a CacheSettings and returns a new Cache
func NewCache(settings CacheSettings) Cache {
	// disable the cache if the directory isn't specified
	if settings.Directory == "" {
		return &nopCache{}
	}
	return &cache{
		cacheDirectory: settings.Directory,
		logger:         settings.Logger,
	}
}

func (c *cache) Enabled() bool {
	return true
}

// LatestVersion TODO(docs)
func (c *cache) LatestVersion() *Version {
	return c.Version(VersionLatest)
}

// SaveLatestVersion TODO(docs)
func (c *cache) SaveLatestVersion(version *Version) {
	if version == nil {
		return
	}
	// save the latest version
	err := c.writeVersion(version, c.latestVersionPath())
	if err != nil {
		c.logger.Error("error writing latest version", zap.Error(err), zap.String("version", version.Version))
	}
	// also save the specific version
	c.SaveVersion(version)
}

// Version TODO(doc)
func (c *cache) Version(version string) *Version {
	var path string
	if version == VersionLatest {
		path = c.latestVersionPath()
	} else {
		path = c.versionPath(version)
	}
	v, err := c.readVersion(path)
	if err != nil {
		c.logger.Error("error reading version", zap.Error(err), zap.String("version", version))
		return nil
	}
	return v
}

// SaveVersion TODO(doc)
func (c *cache) SaveVersion(version *Version) {
	if version == nil || version.Version == "" {
		// if version is missing, it's not a valid version
		return
	}
	err := c.writeVersion(version, c.versionPath(version.Version))
	if err != nil {
		c.logger.Error("error writing version", zap.Error(err), zap.String("version", version.Version))
	}
}

func (c *cache) Artifact(artifactType ArtifactType, version *Version, platform string) CacheArtifact {
	path := c.artifactPath(artifactType, version, platform)
	if path == "" {
		return nopCacheArtifact{}
	}
	return cacheArtifact{path: path}
}

func (c *cache) artifactPath(artifactType ArtifactType, version *Version, platform string) string {
	name := artifactName(artifactType, version, platform)
	if name == "" {
		return ""
	}
	return filepath.Join(c.cacheDirectory, version.Version, platform, name)
}

func artifactName(artifactType ArtifactType, version *Version, platform string) string {
	urls, ok := version.Downloads[platform]
	if !ok {
		return ""
	}
	value, ok := urls[artifactType]
	if !ok {
		return ""
	}
	parsed, err := url.Parse(value)
	if err != nil {
		return ""
	}
	parts := strings.Split(parsed.Path, "/")
	return parts[len(parts)-1]
}

// ----------------------------------------------------------------------

func (c *cache) latestVersionPath() string {
	return filepath.Join(c.cacheDirectory, "latest.json")
}

func (c *cache) versionPath(version string) string {
	return filepath.Join(c.cacheDirectory, version, "version.json")
}

// readVersion attempts to read and unmarshal a Version from the specified path. If the file does not exist, it returns
// nil and not an error.
func (c *cache) readVersion(path string) (*Version, error) {
	var version Version

	bytes, err := ioutil.ReadFile(filepath.Clean(path))
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}

	err = json.Unmarshal(bytes, &version)
	if err != nil {
		return nil, err
	}

	return &version, nil
}

func (c *cache) writeVersion(version *Version, path string) error {
	// make sure the directory exists
	parent, _ := filepath.Split(path)
	if err := os.MkdirAll(parent, 0700); err != nil {
		return err
	}

	bytes, err := json.Marshal(version)
	if err != nil {
		return err
	}

	if err = ioutil.WriteFile(path, bytes, 0600); err != nil {
		return err
	}

	return nil
}

// ----------------------------------------------------------------------

type cacheArtifact struct {
	path   string
	logger *zap.Logger
}

var _ CacheArtifact = (*cacheArtifact)(nil)

func (c cacheArtifact) Name() string {
	return filepath.Base(c.path)
}

func (c cacheArtifact) Path() string {
	return c.path
}

func (c cacheArtifact) Exists() bool {
	info, err := os.Stat(c.path)
	return err == nil && info.Size() > 0
}

func (c cacheArtifact) Read() ([]byte, error) {
	return ioutil.ReadFile(c.path)
}
func (c cacheArtifact) Reader() (io.ReadCloser, error) {
	return os.Open(c.path)
}

func (c cacheArtifact) Write(bytes []byte) error {
	writer := c.Writer()
	_, err := writer.Write(bytes)
	if err != nil {
		return err
	}
	return writer.Close()
}
func (c cacheArtifact) Writer() io.WriteCloser {
	// make sure the directory exists
	parent, _ := filepath.Split(c.path)
	err := os.MkdirAll(parent, 0700)
	if err != nil {
		c.logger.Error("unable to create folder for cache", zap.String("path", c.path), zap.Error(err))
		return &nopWriter{}
	}
	file, err := os.Create(c.path)
	if err != nil {
		c.logger.Error("unable to open writer for cache", zap.String("path", c.path), zap.Error(err))
		return &nopWriter{}
	}
	return file
}

// ----------------------------------------------------------------------

// TODO(andy) write tmpfile writer with copy https://github.com/observIQ/bindplane/issues/245
