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

package flags

import (
	"github.com/observiq/bindplane-op/internal/agent"
	"github.com/spf13/cobra"
)

// Global adds global flags
func Global(cmd *cobra.Command) {
	pf := newflags(cmd.PersistentFlags())
	pf.String("env", "production", "BindPlane environment. One of test|development|production")
	pf.StringP("output", "o", "table", "output format. One of: json|table|yaml|raw")
	pf.String("host", "localhost", "domain on which the BindPlane server will run")
	pf.String("port", "3001", "port on which the rest server is listening")
	pf.String("server-url", "", "http url that clients use to connect to the server")
	pf.String("username", "admin", "username to use with Basic auth")
	pf.String("password", "admin", "password to use with Basic auth")
	pf.String("tls-cert", "", "TLS certificate file")
	pf.String("tls-key", "", "TLS private key file")
	pf.StringSlice("tls-ca", make([]string, 0), "TLS certificate authority file(s) for mutual TLS authentication")
	pf.Bool("tls-skip-verify", false, "Whether to verify the server's certificate chain and host name when making client requests")
	pf.String("log-file-path", "", "full path of the BindPlane log file, defaults to $HOME/.bindplane/bindplane.log")
	pf.String("log-output", "", "output of the log. One of: file|stdout")
}

// Serve adds flags for the serve command
func Serve(cmd *cobra.Command) {
	f := newflags(cmd.Flags())
	f.Bool("offline", false, "BindPlane server should be run in offline mode")
	f.String("store-type", "", "type of store to use for storing agent status and configuration resources")
	f.String("remote-url", "", "websocket url that agents use to connect to the server")
	f.String("secret-key", "", "secret key used by agents when connecting to the server")
	f.String("sessions-secret", "", "secret key used to sign cookies for session authentication, must be a UUID")
	f.String("storage-file-path", "", "full path to the desired storage file, defaults to the $HOME/.bindplane/storage")
	f.String("downloads-folder-path", "", "full path to the downloads folder where agents are cached, defaults to $HOME/.bindplane/downloads")
	f.String("agents-service-url", agent.DefaultAgentVersionsURL, "url of the service that provides agent release information")
	f.Bool("disable-downloads-cache", false, "true if agent distributions should be cached")
}
