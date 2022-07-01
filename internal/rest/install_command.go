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

package rest

import (
	"fmt"
	"runtime"
	"strings"
)

type installCommandParameters struct {
	platform  supportedPlatform
	version   string
	labels    string
	secretKey string
	remoteURL string
	serverURL string
}

type supportedPlatform string

const (
	linuxArm64   supportedPlatform = "linux-arm64"
	linuxAmd64                     = "linux-amd64"
	linuxArm                       = "linux-arm"
	darwinArm64                    = "darwin-arm64"
	darwinAmd64                    = "darwin-amd64"
	windowsAmd64                   = "windows-amd64"
)

var platformAliases = map[string]supportedPlatform{
	// aliases
	"windows":     windowsAmd64,
	"linux":       linuxAmd64,
	"darwin":      darwinArm64,
	"macos":       darwinArm64,
	"macos-arm64": darwinArm64,
	"macos-amd64": darwinAmd64,

	// include supportedPlatform here for validation
	"linux-arm64":   linuxArm64,
	"linux-amd64":   linuxAmd64,
	"linux-arm":     linuxArm,
	"darwin-arm64":  darwinArm64,
	"darwin-amd64":  darwinAmd64,
	"windows-amd64": windowsAmd64,
}

func normalizePlatform(platform string) (supportedPlatform, bool) {
	if platform == "" {
		platform = fmt.Sprintf("%s-%s", runtime.GOOS, runtime.GOARCH)
	}
	p, ok := platformAliases[platform]
	return p, ok
}

// setarg returns an expression to set a command argument. It is important that it include spacing so that it can
// be directly inserted into the command.
func (p *installCommandParameters) setarg(name, value string) string {
	if value == "" {
		return ""
	}
	switch p.platform {
	case windowsAmd64:
		return fmt.Sprintf(` %s=%s`, name, value)
	default:
		return fmt.Sprintf(` %s %s`, name, value)
	}
}

func (p *installCommandParameters) versionNoV() string {
	if p.version == "latest" {
		return p.version
	}
	if strings.HasPrefix(p.version, "v") {
		return p.version[1:]
	}
	return p.version
}

func (p *installCommandParameters) versionWithV() string {
	if p.version == "latest" {
		return p.version
	}
	version := p.versionNoV()
	return fmt.Sprintf("v%s", version)
}

func (p *installCommandParameters) args() string {
	switch p.platform {
	case windowsAmd64:
		return fmt.Sprintf("%s%s%s%s",
			p.setarg("ENABLEMANAGEMENT", "1"),
			p.setarg("OPAMPENDPOINT", p.remoteURL),
			p.setarg("OPAMPSECRETKEY", p.secretKey),
			p.setarg("OPAMPLABELS", p.labels),
		)

	default:
		if p.version == "latest" {
			return fmt.Sprintf("%s%s%s",
				p.setarg("-e", p.remoteURL),
				p.setarg("-s", p.secretKey),
				p.setarg("-k", p.labels),
			)
		}
		return fmt.Sprintf("%s%s%s%s",
			p.setarg("-e", p.remoteURL),
			p.setarg("-s", p.secretKey),
			p.setarg("-k", p.labels),
			p.setarg("-v", p.versionNoV()),
		)
	}
}

func (p *installCommandParameters) installerFilename() string {
	switch p.platform {
	case windowsAmd64:
		return "observiq-otel-collector.msi"

	case darwinAmd64:
		fallthrough
	case darwinArm64:
		return "install_macos.sh"

	default:
		return "install_unix.sh"
	}
}

func (p *installCommandParameters) installerURL() string {
	if p.version == "latest" {
		return fmt.Sprintf("https://github.com/observiq/observiq-otel-collector/releases/latest/download/%s",
			p.installerFilename(),
		)
	}
	return fmt.Sprintf("https://github.com/observiq/observiq-otel-collector/releases/download/%s/%s",
		p.versionWithV(),
		p.installerFilename(),
	)
}

func (p *installCommandParameters) installCommand() string {
	switch p.platform {
	case windowsAmd64:
		return fmt.Sprintf(`msiexec /i "%s" /quiet%s`,
			p.installerURL(),
			p.args(),
		)

	default:
		return fmt.Sprintf(`sudo sh -c "$(curl -fsSlL %s)" %s%s`,
			p.installerURL(),
			p.installerFilename(),
			p.args(),
		)
	}
}

// Windows:
//
// msiexec /i "https://github.com/observiq/observiq-otel-collector/releases/latest/download/observiq-otel-collector.msi" /quiet ENABLEMANAGEMENT=1 OPAMPENDPOINT=<your-endpoint> OPAMPSECRETKEY=<secret-key> OPAMPLABELS=<comma-separated-labels>
//
// Linux:
//
// sudo sh -c "$(curl -fsSlL https://github.com/observiq/observiq-otel-collector/releases/latest/download/install_unix.sh)" install_unix.sh -e=<your-endpoint> -k=<comma-separated-labels> -s=<secret-key> -v <collector-version>
//
// macOS:
//
// sudo sh -c "$(curl -fsSlL https://github.com/observiq/observiq-otel-collector/releases/latest/download/install_macos.sh)" install_macos.sh -e=<your-endpoint> -k=<comma-separated-labels> -s=<secret-key>
