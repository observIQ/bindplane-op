package agent

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestGithubVersion(t *testing.T) {
	github := newGithub()
	version, err := github.Version("v1.4.0")
	require.NoError(t, err)
	require.Equal(t, "v1.4.0", version.Version)
	require.True(t, version.Public, "should be public")
	require.Equal(t, "https://github.com/observIQ/observiq-otel-collector/releases/download/v1.4.0/observiq-otel-collector-v1.4.0-darwin-amd64.tar.gz", version.ArtifactURL(Download, "darwin-amd64"))
	require.Equal(t, "https://github.com/observIQ/observiq-otel-collector/releases/download/v1.4.0/install_macos.sh", version.ArtifactURL(Installer, "darwin-amd64"))
}
