package agent

import (
	"fmt"
	"time"

	"github.com/go-resty/resty/v2"
)

type github struct {
	client *resty.Client
}

var _ Client = (*github)(nil)

// newGithub creates a new github client for retrieving agent versions
func newGithub() Client {
	c := resty.New()
	c.SetTimeout(time.Second * 20)
	c.SetBaseURL("https://api.github.com")
	return &github{
		client: c,
	}
}

// LatestVersion returns the latest agent release.
func (c *github) LatestVersion() (*Version, error) {
	return c.Version(VersionLatest)
}

type githubReleaseAsset struct {
	Name        string
	DownloadURL string `json:"browser_download_url"`
}

type githubRelease struct {
	Name       string
	TagName    string `json:"tag_name"`
	Draft      bool
	Prerelease bool
	Assets     []githubReleaseAsset
}

const owner = "observIQ"
const repo = "observiq-otel-collector"

func latestURL() string {
	return fmt.Sprintf("/repos/%s/%s/releases/latest", owner, repo)
}
func versionURL(version string) string {
	return fmt.Sprintf("/repos/%s/%s/releases/tags/%s", owner, repo, version)
}

func (c *github) Version(version string) (*Version, error) {
	var release githubRelease
	res, err := c.client.R().SetResult(&release).Get(versionURL(version))

	if err != nil {
		return nil, err
	}
	if res.StatusCode() == 404 {
		return nil, ErrVersionNotFound
	}
	if res.StatusCode() != 200 {
		return nil, fmt.Errorf("Unable to get version %s: %s", version, res.Status())
	}

	return convertRelease(&release), nil
}

func (c *github) Artifact(artifactType ArtifactType, version *Version, platform string) Artifact {
	return nil
}

var platformArtifacts = map[string]struct {
	// format for use with Sprintf(format, version)
	downloadPackageFormat string
	// name of the installer for this platform
	installerName string
}{
	"darwin-amd64": {
		downloadPackageFormat: "observiq-otel-collector-%s-darwin-amd64.tar.gz",
		installerName:         "install_macos.sh",
	},
	"darwin-arm64": {
		downloadPackageFormat: "observiq-otel-collector-%s-darwin-arm64.tar.gz",
		installerName:         "install_macos.sh",
	},
	"linux-amd64": {
		downloadPackageFormat: "observiq-otel-collector-%s-linux-amd64.tar.gz",
		installerName:         "install_unix.sh",
	},
	"linux-arm64": {
		downloadPackageFormat: "observiq-otel-collector-%s-linux-arm64.tar.gz",
		installerName:         "install_unix.sh",
	},
	"linux-arm": {
		downloadPackageFormat: "observiq-otel-collector-%s-linux-arm.tar.gz",
		installerName:         "install_unix.sh",
	},
	"windows-amd64": {
		downloadPackageFormat: "observiq-otel-collector-%s-windows-amd64.zip",
		installerName:         "observiq-otel-collector.msi",
	},
}

func convertRelease(r *githubRelease) *Version {
	downloads := map[string]map[ArtifactType]string{}
	for platform, artifacts := range platformArtifacts {
		downloads[platform] = map[ArtifactType]string{
			Download:  releaseAssetURL(fmt.Sprintf(artifacts.downloadPackageFormat, r.TagName), r.Assets),
			Installer: releaseAssetURL(artifacts.installerName, r.Assets),
		}
	}
	return &Version{
		Version:   r.Name,
		Public:    !r.Prerelease && !r.Draft,
		Downloads: downloads,
	}
}

func releaseAssetURL(name string, assets []githubReleaseAsset) string {
	for _, asset := range assets {
		if asset.Name == name {
			return asset.DownloadURL
		}
	}
	return ""
}
