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
	"fmt"
	"net/url"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"

	"github.com/observiq/bindplane/internal/cli/flags"
	"github.com/observiq/bindplane/model"
)

// SetCommand returns the BindPlane profile set cobra command
func SetCommand(h Helper) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "set <name>",
		Short: "set a parameter on a saved profile",
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) == 0 {
				return fmt.Errorf("missing required argument <name>")
			}

			name := args[0]
			f := h.Folder()

			profile, err := f.ReadProfile(name)
			if err != nil {
				profile = model.NewProfile(name, model.ProfileSpec{})
			}

			handleFlag := func(f *pflag.Flag) {
				if f.Changed {
					switch f.Name {
					case "port":
						profile.Spec.Port = f.Value.String()
					case "host":
						profile.Spec.Host = f.Value.String()
					case "server-url":
						serverAddress := f.Value.String()
						u, err := url.Parse(serverAddress)
						if err == nil {
							if u.Scheme == "" {
								u.Scheme = "http"
							}
							serverAddress = u.String()
						}
						profile.Spec.Common.ServerURL = serverAddress
					case "remote-url":
						remoteURL := f.Value.String()
						u, err := url.Parse(remoteURL)
						if err == nil {
							if u.Scheme == "" {
								u.Scheme = "ws"
							}
							remoteURL = u.String()
						}
						profile.Spec.Server.RemoteURL = remoteURL
					case "agents-service-url":
						agentsURL := f.Value.String()
						u, err := url.Parse(agentsURL)
						if err == nil {
							if u.Scheme == "" {
								u.Scheme = "http"
							}
							agentsURL = u.String()
						}
						profile.Spec.Server.AgentsServiceURL = agentsURL
					case "secret-key":
						profile.Spec.Server.SecretKey = f.Value.String()
					case "username":
						profile.Spec.Username = f.Value.String()
					case "password":
						profile.Spec.Password = f.Value.String()
					case "storage-file-path":
						profile.Spec.Server.StorageFilePath = f.Value.String()
					case "tls-cert":
						profile.Spec.Common.Certificate = f.Value.String()
					case "tls-key":
						profile.Spec.Common.PrivateKey = f.Value.String()
					case "tls-ca":
						stringValue := f.Value.String()                                // In the case of StringSlice this looks like `"[one,two]"`
						value := strings.Split(stringValue[1:len(stringValue)-1], ",") // removes the brackets
						profile.Spec.Common.CertificateAuthority = value
					case "log-file-path":
						profile.Spec.Common.LogFilePath = f.Value.String()
					case "downloads-folder-path":
						profile.Spec.Server.DownloadsFolderPath = f.Value.String()
					case "disable-downloads-cache":
						profile.Spec.Server.DisableDownloadsCache = f.Value.String() == "true"
					case "output":
						profile.Spec.Command.Output = f.Value.String()
					case "offline":
						profile.Spec.Server.Offline = f.Value.String() == "true"
					case "sessions-secret":
						profile.Spec.Server.SessionsSecret = f.Value.String()
					}
				}
			}

			cmd.InheritedFlags().VisitAll(handleFlag)
			cmd.Flags().VisitAll(handleFlag)

			err = f.WriteProfile(profile)
			if err != nil {
				return err
			}

			return nil
		},
	}

	flags.Serve(cmd)

	return cmd
}
