[![Continuous Integration](https://github.com/observIQ/bindplane-op/actions/workflows/ci.yml/badge.svg)](https://github.com/observIQ/bindplane-op/actions/workflows/ci.yml)
[![License](https://img.shields.io/badge/License-Apache_2.0-blue.svg)](https://opensource.org/licenses/Apache-2.0)

# BindPlane OP

Next generation agent management platform

---

## Installation and Configuration

Installation: [docs/install.md](docs/install.md)

For configurable variables, BindPlane will look for flags, environment variables, and a configuration file, with precedence: flags > environment variables > configuration file.

### Global configuration

| Name        | Description                                                                              | Flag            | Env Variable              | Default            |
| ----------- | ---------------------------------------------------------------------------------------- | --------------- | ------------------------- | ------------------ |
| Port        | The port on which the bindplane server runs                                                   | --port          | BINDPLANE_CONFIG_PORT          | `3001`             |
| Host        | The host on which the bindplane server runs                                                   | --host          | BINDPLANE_CONFIG_HOST          | `localhost`        |
| Server URL  | The address of the remote bindplane server, if not set will be inferred as `http://host:port` | --server-url    | BINDPLANE_CONFIG_SERVER_URL    |                    |
| Username    | Basic auth username                                                                      | --username      | BINDPLANE_CONFIG_USERNAME      | `admin`            |
| Password    | Basic auth password                                                                      | --password      | BINDPLANE_CONFIG_PASSWORD      | `admin`            |
| TLSConfig   | See "TLS configuration" section                                                          |                 |                           |                    |
| LogFilePath | The full path to the log file                                                            | --log-file-path | BINDPLANE_CONFIG_LOG_FILE_PATH | `~/.bindplane/bindplane.log` |

### Server configuration

| Name                    | Description                                                                                                         | Flag                      | Env Variable                        | Default                           |
| ----------------------- | ------------------------------------------------------------------------------------------------------------------- | ------------------------- | ----------------------------------- | --------------------------------- |
| Storage File Path       | Full path to the desired storage file for persistent data                                                           | --storage-file-path       | BINDPLANE_CONFIG_STORAGE_FILE_PATH       | `~/.bindplane/storage`                 |
| Secret Key              | Shared key (UUID) between server and agent for authentication                                                       | --secret-key              | BINDPLANE_CONFIG_SECRET_KEY              |                                   |
| Remote URL              | Websocket URL used by agents connecting to BindPlane, if not set will be inferred as `ws://host:port`                    | --remote-url              | BINDPLANE_CONFIG_REMOTE_URL              |                                   |
| Offline                 | BindPlane will not attempt to connect to external systems. REST and Websocket are still available for clients and agents | --offline                 | BINDPLANE_CONFIG_OFFLINE                 | `false`                           |
| Agents Service URL      | The URL used to download agent releases                                                                             | --agents-service-url      | BINDPLANE_CONFIG_AGENTS_SERVICE_URL      | `https://agents.app.observiq.com` |
| Downloads Folder        | Directory used to cache agent downloads                                                                             | --downloads-folder-path   | BINDPLANE_CONFIG_DOWNLOADS_FOLDER_PATH   | `~/.bindplane/downloads`               |
| Disable Downloads Cache | BindPlane will not attempt to cache agent downloads                                                                      | --disable-downloads-cache | BINDPLANE_CONFIG_DISABLE_DOWNLOADS_CACHE | `false`                           |

### TLS configuration

BindPlane supports TLS for communication between client, server, and agent. When a certificate authority is set, mTLS is enabled. All certificate
and private keys are expected to be x509 PEM encoded. TLS is disabled by default.

| Name                      | Description                                                              | Flag       | Env Variable         | Default |
| ------------------------- | ------------------------------------------------------------------------ | ---------- | -------------------- | ------- |
| TLS Certificate           | The TLS Certificate (x509 PEM) to use for client and server interaction  | --tls-cert | BINDPLANE_CONFIG_TLS_CERT |         |
| TLS Private Key           | The TLS Private Key (x509 PEM) to use for client and server interfaction | --tls-key  | BINDPLANE_CONFIG_TLS_KEY  |         |
| TLS Certificate Authority | The TLS certificate authority (x509 PEM), when set, **mTLS is enabled**  | --tls-ca   | BINDPLANE_CONFIG_TLS_CA   |         |


### Command configuration

| Name   | Description                                                            | Flag         | Env Variable       | Default |
| ------ | ---------------------------------------------------------------------- | ------------ | ------------------ | ------- |
| Output | specify either json, yaml, or table formatting for command line output | --output, -o | BINDPLANE_CONFIG_OUTPUT | table   |

### Custom Configuration File

You can pass in a full path to a configuration file with the --config flag. BindPlane expects the global variables to be unnested like so:

`sample-config.yaml`

```sh
host: localhost
port: "5000"
server:
  remoteURL: ws://localhost:3001
```

```sh
bindplane [commands] --config sample-config.yaml
```

### BindPlane App configuration with Profile command.

You can set values on a saved configuration with the `bindplane profile set` command. For example if you wanted to save a profile that uses a server hosted at `https://remote-address.com` you could save that under the `local` profile with this command:

```sh
bindplane profile set local --server-url https://remote-address.com
```

`bindplane profile get <name>` command returns the profile yaml

```sh
bindplane profile get local

apiVersion: bindplane.observiq.com/v1beta
kind: Profile
metadata:
  name: local
spec:
  serverUrl: https://remote-address.com
```

`bindplane profile get --current` returns the settings of the current profile

Note that this returns the `Resource` form of the configuration, the pertinent variables are set in `spec`.

`bindplane profile list` returns the available saved profiles.

`bindplane profile delete <name>` will remove a saved profile.

`bindplane profile use <name>` will set the default context to use on startup.

`bindplane profile current` will return the name of the currently used profile

---

### BindPlane Home

BindPlane home defaults to the running user's home directory. For example, the `observiq` user's bindplane directory would be
found in `/home/observiq/.bindplane`.

```
/home/observiq/.bindplane
├── bindplane-2022-03-04T21-07-26.022.log.gz
├── bindplane.log
├── profiles
│   ├── current
│   ├── docker-https-mtls.yaml
│   ├── docker-https.yaml
│   ├── docker-http.yaml
│   ├── local.yaml
│   └── poc.yaml
└── storage
```

You can set `BINDPLANE_CONFIG_HOME` to override this behavior. For example, the server package will install
with `BINDPLANE_CONFIG_HOME=/var/lib/bindplane` in addition to setting the log, storage, and download paths.

Systemd Service and Config snippet:

```
Environment="BINDPLANE_CONFIG_HOME=/var/lib/bindplane"
```

```
logFilePath: /var/log/bindplane/bindplane.log
bindplaneHome: /var/lib/bindplane
server:
  storageFilePath: /var/lib/bindplane/storage/bindplane.db
  downloadsFolderPath: /var/lib/bindplane/downloads
```

## Developing

### Setup

- BindPlane is Developed using Go. If you have not worked with Go before, it is recommended to work through the [How to Write Go Code](https://golang.org/doc/code.html) tutorial, which will help you get your Go environment configured.
- It is important to configure your `$GOPATH` within your shell's environment using `.bashrc` or `.zshrc` (Depending on which shell you use)
- Tooling can be installed with the `make install-tools` command

### Testing

Several commands are available for testing

- `make vet`: Runs [go vet](https://pkg.go.dev/cmd/vet) for each supported platform
- `make test`: Runs `make vet` and `go test ./... -race -cover`
- `make lint`: Runs [revive](https://revive.run/) linter

### Building

BindPlane can be built with the `make build` command. The output binary will be in the `dist/` directory.

A full test release can be built with the `make release-test`. A test release will do the following:

- Build BindPlane for all supported platforms and architectures
- Archive binaries using zip
- Build Linux packages
- Build container images
- Generate a sha256 checksum file
- Place all output in the `dist/` directory

### Additional Commands

- `make secure` will run [gosec](https://github.com/securego/gosec) to catch things such as unhandled errors, weak file permissions, bad TLS versions, etc
- `make check-license` and `make add-license` for handling source file license headers
- `make generate` will generate GraphQL files and add in license headers (use this instead of using `go generate ./...` directly)
- `make swagger` will generate swagger documentation in the `/docs` directory.
- `make install` will install tools and node dependencies.
- `make ci` will run `npm ci` in the `ui` directory.
- `make dev` will run the BindPlane UI dev server on port 3000 and serve BindPlane on port 3001.
- `make ui-build` will clean install node modules, build the UI bundle, and build bindplane.
- `make ui-test` will run npm run test in the ui directory.
