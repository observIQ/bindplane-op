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

// latest can be used in requests instead of an actual version
const (
	VersionLatest = "latest"
)

// Version represents an agent release
type Version struct {
	// Version is the release version tag
	Version string `json:"version"`
	// Public is true if this version has been publicly released
	Public bool `json:"public"`
	// Downloads is a map from platform => artifactKey => url
	Downloads map[string]map[ArtifactType]string `json:"downloads"`
}

// ArtifactURL returns the download URL of the specified artifact type on the specified platform
func (v *Version) ArtifactURL(artifactType ArtifactType, platform string) string {
	return v.Downloads[platform][artifactType]
}
