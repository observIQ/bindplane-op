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

package common

import (
	"fmt"
	"os"
	"path"

	"github.com/observiq/bindplane-op/internal/trace"
)

const (
	// ProfilesFolderName is the name of the folder where individual configuration profiles are stored
	ProfilesFolderName = "profiles"
	// BindPlaneDirectoryName TODO(doc)
	BindPlaneDirectoryName = ".bindplane"
	// BoldDatabaseName is the name of the bbolt database file
	BoldDatabaseName = "storage"
	// DownloadsDirectoryName is the name of the directory where downloads are cached
	DownloadsDirectoryName = "downloads"
	// BindPlaneLogName returns the name of the BindPlane log file
	BindPlaneLogName = "bindplane.log"
	// DefaultProfileName is the name of the default profile
	DefaultProfileName = "default"
)

// LogOutput is an enum of possible values for the LogOutput configuration setting
type LogOutput string

const (
	// LogOutputFile will write logs to the file specified by LogFilePath
	LogOutputFile LogOutput = "file"

	// LogOutputStdout will write logs to stdout
	LogOutputStdout LogOutput = "stdout"
)

// DefaultBindPlaneHomePath returns the default value of the bindplane home path
func DefaultBindPlaneHomePath() string {
	return os.ExpandEnv(fmt.Sprintf("$HOME/%s", BindPlaneDirectoryName))
}

// Config TODO(doc)
type Config struct {
	// Server TODO(doc)
	Server `mapstructure:"server" yaml:"server,omitempty"`
	// Client TODO(doc)
	Client `mapstructure:"client" yaml:"client,omitempty"`
	// Command TODO(doc)
	Command `mapstructure:"command" yaml:"command,omitempty"`
}

// Env is an enum indicating the environment in which BindPlane is running.
type Env string

const (
	// EnvDevelopment should be used for development and uses debug logging and normal gin request logging to stdout.
	EnvDevelopment Env = "development"

	// EnvTest should be used for tests and uses debug logging with json gin request logging to the log file.
	EnvTest Env = "test"

	// EnvProduction the the default and should be used in production and uses info logging with json gin request logging to the log file.
	EnvProduction Env = "production"
)

// Common TODO(doc)
type Common struct {
	// Env is one of
	Env Env `mapstructure:"env" yaml:"env,omitempty"`

	// Host is the Host to which the server will bind.
	Host string `mapstructure:"host" yaml:"host,omitempty"`

	// Port is the Port on which the server will serve.
	Port string `mapstructure:"port" yaml:"port,omitempty"`

	// ServerURL is the URL that clients should use to contact the server.
	ServerURL string `mapstructure:"serverURL" yaml:"serverURL,omitempty"`

	// Username the basic auth username used for communication between client and server.
	Username string `mapstructure:"username" yaml:"username,omitempty"`
	// The basic auth password used for communication between client and server.
	Password string `mapstructure:"password" yaml:"password,omitempty"`

	// TLSConfig is an optional TLS configuration for communication between client and server.
	TLSConfig `yaml:",inline" mapstructure:",squash"`

	// LogFilePath is the path of the bindplane log file, defaulting to $HOME/.bindplane/bindplane.log
	LogFilePath string `mapstructure:"logFilePath" yaml:"logFilePath,omitempty"`

	// LogOutput indicates where logs should be written, defaulting to "file"
	LogOutput LogOutput `mapstructure:"logOutput" yaml:"logOutput,omitempty"`

	// bindplaneHomePath is the root folder path of BindPlane home, defaulting to $HOME/.bindplane.
	// It is read-only and available via BindPlaneHomePath()
	bindplaneHomePath string

	// TraceType enables tracing
	TraceType string `mapstructure:"traceType,omitempty" yaml:"traceType,omitempty"`

	// GoogleCloudTracing is used to send traces to Google Cloud when TraceType is set to "google".
	GoogleCloudTracing trace.GoogleCloudTracing `mapstructure:"googleTracing,omitempty" yaml:"googleTracing,omitempty"`

	// OpenTelemetryTracing is used to send traces to an Open Telemetry OTLP receiver when
	// TraceType is set to "otlp".
	OpenTelemetryTracing trace.OpenTelemetryTracing `mapstructure:"otlpTracing,omitempty" yaml:"otlpTracing,omitempty"`
}

// TLSConfig contains configuration for connecting over TLS and mTLS.
type TLSConfig struct {
	// Certificate is the path to the x509 PEM encoded certificate file that will be used to
	// establish TLS connections.
	//
	// When operating in server mode, this certificate is presented to clients.
	// When operating in client mode with mTLS, this certificate is used for authentication
	// against the server.
	Certificate string `mapstructure:"tlsCert" yaml:"tlsCert,omitempty"`

	// PrivateKey is the matching x509 PEM encoded private key for the Certificate.
	PrivateKey string `mapstructure:"tlsKey" yaml:"tlsKey,omitempty"`

	// CertificateAuthority is one or more file paths to x509 PEM encoded certificate authority chains.
	// These certificate authorities are used for trusting incoming client mTLS connections.
	CertificateAuthority []string `mapstructure:"tlsCa" yaml:"tlsCa,omitempty"`
}

const (
	// StoreTypeMap uses an in-memory store
	StoreTypeMap = "map"
	// StoreTypeBbolt uses go.etcd.io/bbolt for storage
	StoreTypeBbolt = "bbolt"
	// StoreTypeGoogleCloud uses Google Cloud Datastore for storage
	StoreTypeGoogleCloud = "googlecloud"
)

// Server TODO(doc)
type Server struct {
	// StoreType indicates the type of store to use. "map", "bbolt", and "googlecloud" are currently supported.
	StoreType string `mapstructure:"storeType,omitempty" yaml:"storeType,omitempty"`

	// GoogleCloudDatastore contains configuration for contacting Google Could Datastore and is used if StoreType == "googlecloud"
	GoogleCloudDatastore *GoogleCloudDatastore `mapstructure:"datastore,omitempty" yaml:"datastore,omitempty"`

	// GoogleCloudPubSub contains configuration for contacting Google Could Pub/Sub and is used if StoreType == "googlecloud"
	GoogleCloudPubSub *GoogleCloudPubSub `yaml:"pubsub,omitempty" mapstructure:"pubsub,omitempty"`

	// StorageFilePath TODO(doc)
	StorageFilePath string `mapstructure:"storageFilePath,omitempty" yaml:"storageFilePath,omitempty"`

	// SecretKey is a shared secret between the server and the agent to ensure agents are authorized to communicate with the server.
	SecretKey string `mapstructure:"secretKey,omitempty" yaml:"secretKey,omitempty"`

	// RemoteURL is the URL that agents should use to contact the server
	RemoteURL string `mapstructure:"remoteURL,omitempty" yaml:"remoteURL,omitempty"`

	// Offline mode indicates if the server should be considered offline. An offline server will not attempt to contact
	// any other services. It will still allow agents to connect and serve api requests.
	Offline bool `mapstructure:"offline,omitempty" yaml:"offline,omitempty"`

	// AgentsServiceURL is the url of the Agent Versions server which manages the release of the agent. This service will
	// be contacted to determine the most recent agent version and to identify the location of artifacts for each version
	// of the agent.
	AgentsServiceURL string `mapstructure:"agentsServiceURL,omitempty" yaml:"agentsServiceURL,omitempty"`

	// DownloadsFolderPath is the path to the folder where agent versions are stored for upgrading agents
	DownloadsFolderPath string `mapstructure:"downloadsFolderPath,omitempty" yaml:"downloadsFolderPath,omitempty"`
	// DisableDownloadsCache TODO(doc)
	DisableDownloadsCache bool `mapstructure:"disableDownloadsCache,omitempty" yaml:"disableDownloadsCache,omitempty"`

	// SessionSecret is used to encode the user sessions cookies.  It should be a uuid.
	SessionsSecret string `mapstructure:"sessionsSecret,omitempty" yaml:"sessionsSecret,omitempty"`

	Common `yaml:",inline" mapstructure:",squash"`
}

// GoogleCloudDatastore contains the configuration for google cloud datastore
type GoogleCloudDatastore struct {
	ProjectID       string `mapstructure:"projectID,omitempty" yaml:"projectID,omitempty"`
	Endpoint        string `mapstructure:"endpoint,omitempty" yaml:"endpoint,omitempty"`
	CredentialsFile string `mapstructure:"credentialsFile,omitempty" yaml:"credentialsFile,omitempty"`
}

// GoogleCloudPubSub is configuration for a server's Pub/Sub subscriber and publisher
type GoogleCloudPubSub struct {
	ProjectID       string `mapstructure:"projectID,omitempty" yaml:"projectID,omitempty"`
	Endpoint        string `mapstructure:"endpoint,omitempty" yaml:"endpoint,omitempty"`
	CredentialsFile string `mapstructure:"credentialsFile,omitempty" yaml:"credentialsFile,omitempty"`
	Topic           string `mapstructure:"topic,omitempty" yaml:"topic,omitempty"`

	// Subscription is the name of the subscription that this node should use. In production this will be generated but it
	// is useful to specify in development.
	Subscription string `mapstructure:"subscription,omitempty" yaml:"subscription,omitempty"`
}

// Client TODO(doc)
type Client struct {
	Common
}

// Command TODO(doc)
type Command struct {
	// TODO(doc)
	Output string `mapstructure:"output" yaml:"output"`
}

// InitConfig returns a Config struct with zero values, which will be assigned during Command's PersistentPreRun
func InitConfig(bindplaneHomePath string) *Config {
	common := Common{bindplaneHomePath: bindplaneHomePath}
	server := Server{Common: common}
	client := Client{Common: common}
	command := Command{}

	return &Config{
		Server:  server,
		Client:  client,
		Command: command,
	}
}

// ----------------------------------------------------------------------
// Server

// BindAddress is the address (host:port) to which the server will bind
func (c *Server) BindAddress() string {
	return fmt.Sprintf("%s:%s", c.Host, c.Port)
}

// WebsocketURL is the URL that should be used for agents connecting to the server
func (c *Server) WebsocketURL() string {
	if c.RemoteURL != "" {
		return c.RemoteURL
	}
	if c.Host == "" && c.Port == "" {
		return ""
	}
	return fmt.Sprintf("%s://%s:%s", c.WebsocketScheme(), c.Host, c.Port)
}

// BoltDatabasePath returns the path to the bolt database file
func (c *Server) BoltDatabasePath() string {
	if c.StorageFilePath != "" {
		return c.StorageFilePath
	}
	return path.Join(c.BindPlaneHomePath(), BoldDatabaseName)
}

// BindPlaneDownloadsPath returns the path to the directory where downloads are cached
func (c *Server) BindPlaneDownloadsPath() string {
	if c.DownloadsFolderPath != "" {
		return c.DownloadsFolderPath
	}
	return path.Join(c.BindPlaneHomePath(), DownloadsDirectoryName)
}

// ----------------------------------------------------------------------
// Common

// BindPlaneEnv ensures that Env has a valid value and defaults to EnvProduction
func (c *Common) BindPlaneEnv() Env {
	switch c.Env {
	case EnvDevelopment:
		return EnvDevelopment
	case EnvTest:
		return EnvTest
	default:
		return EnvProduction
	}
}

// BindPlaneHomePath returns the path to the BindPlane home where files are stored by default
func (c *Common) BindPlaneHomePath() string {
	return c.bindplaneHomePath
}

// BindPlaneLogFilePath returns the path to the log file for bindplane
func (c *Common) BindPlaneLogFilePath() string {
	if c.LogFilePath != "" {
		return c.LogFilePath
	}
	return path.Join(c.BindPlaneHomePath(), BindPlaneLogName)
}

// EnableTLS returns true if TLS is enabled
func (c *Common) EnableTLS() bool {
	return c.Certificate != "" && c.PrivateKey != ""
}

// WebsocketScheme returns ws or wss
func (c *Common) WebsocketScheme() string {
	if c.EnableTLS() {
		return "wss"
	}
	return "ws"
}

// ServerScheme returns http or https
func (c *Common) ServerScheme() string {
	if c.EnableTLS() {
		return "https"
	}
	return "http"
}

// BindPlaneURL returns the configured server url. If one is not configured,
// a url derived from the configured host and port is used.
func (c *Common) BindPlaneURL() string {
	if c.ServerURL != "" {
		return c.ServerURL
	}
	if c.Host == "" && c.Port == "" {
		return ""
	}
	return fmt.Sprintf("%s://%s:%s", c.ServerScheme(), c.Host, c.Port)
}
