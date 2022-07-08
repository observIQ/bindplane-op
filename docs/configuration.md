# Configuration

  * [Options](#options)
  * [Initialization](#initialization)
    + [Server](#server)
    + [Client](#client)
  * [Example Configurations](#example-configurations)
    + [Basic](#basic)
    + [TLS](#tls)
      - [Server Side TLS](#server-side-tls)
      - [Mutual TLS](#mutual-tls)
  * [Client Profiles](#client-profiles)

## Options

BindPlane server configuraton can be found at `/etc/bindplane/config.yaml`.

BindPlane will look for flags, environment variables, and a configuration file, with precedence: flags > environment variables > configuration file.

| Configuration Option    | Description                                                           | Flag                | Env Variable                        | Default                      | Type                |
| ----------------------- | --------------------------------------------------------------------- | ------------------- | ----------------------------------- | ---------------------------- | ------------------- |
| host                    | Address the bindplane server binds to                                 | --host              | BINDPLANE_CONFIG_HOST               | `127.0.0.1`                  | IP address          |
| port                    | Tcp port the bindplane server listens on                              | --port              | BINDPLANE_CONFIG_PORT               | `3001`                       | Port number         |
| serverURL               | Address of the remote bindplane server                                | --server-url        | BINDPLANE_CONFIG_SERVER_URL         | `http://127.0.0.1:3001`      | URL                 |
| username                | Basic auth username                                                   | --username          | BINDPLANE_CONFIG_USERNAME           | `admin`                      | String              |
| password                | Basic auth password                                                   | --password          | BINDPLANE_CONFIG_PASSWORD           | `admin`                      | String              |
| logOutput               | Log output (`file` or `stdout`)                                       | --log-output        | BINDPLANE_CONFIG_LOG_OUTPUT         | `file`                       | String              |
| logFilePath             | Path to the log file                                                  | --log-file-path     | BINDPLANE_CONFIG_LOG_FILE_PATH      | `~/.bindplane/bindplane.log` | File path           |
| tlsCert                 | TLS full chain certificate file (See [TLS](./configuration.md#tls))   | --tls-cert          | BINDPLANE_CONFIG_TLS_CERT           | optional                     | File path           |
| tlsKey                  | TLS private key file (See [TLS](./configuration.md#tls))              | --tls-key           | BINDPLANE_CONFIG_TLS_KEY            | optional                     | File path           |
| tlsCA                   | TLS certificate authority file(S) (See [TLS](./configuration.md#tls)) | --tls-ca            | BINDPLANE_CONFIG_TLS_CA             | optional array               | array of file paths |
| server.storeType        | Storage backend type                                                  | --store-type        | BINDPLANE_CONFIG_STORE_TYPE         | `bbolt`                      | String              |
| server.storageFilePath  | Storage file for persistent data (bbolt only)                         | --storage-file-path | BINDPLANE_CONFIG_STORAGE_FILE_PATH  | `~/.bindplane/storage`       | File path           |
| server.secretKey        | Shared key between server and agent for authentication                | --secret-key        | BINDPLANE_CONFIG_SECRET_KEY         | required                     | UUID V4             |
| server.sessionsSecret   | Used to encode signed cookies for UI login                            | --sessions-secret   | BINDPLANE_CONFIG_SESSIONS_SECRET    | required                     | UUID V4             |
| server.remoteURL        | Websocket URL used by agents connecting to BindPlane                  | --remote-url        | BINDPLANE_CONFIG_REMOTE_URL         | `ws://127.0.0.1:3001`        | URL                 |

Server and client configuratins can be bootstrapped using the `init` command. See the [initialization section](./configuration.md#initialization).

For detailed examples, see the [configurations seection](./configuration.md#example-configuration).

## Initialization

The `init` command is useful for bootstrapping a server or client.

### Server

After installing BindPlane server, simply run the following command and follow the prompts.

```bash
sudo bindplane init server \
  --config /etc/bindplane/config.yaml
```

One finished, the server must be restarted.

### Client

Client initalization will create a new profile if one is not
already set. If an existing profile is in use, init will update
that profile. You can learn more about profiles in the [client profiles](./configuration.md#client-profiles) sectio.

```bash
bindplanectl init client
```

Once finished, the client configuration will exist in `~/.bindplane/profiles`. You
can also run the `profile` command:

```bash
bindplanectl profile --help
```

## Example Configurations

The following examples assume the use of [observIQ collectors](https://github.com/observIQ/observiq-otel-collector).

### Basic

This configuration assumes that the BindPlane server is running on
IP address `192.168.1.10`.

**Server Configuration**

```yaml
host: 192.168.1.10
port: 3001
username: myuser
password: mypassword
logfilePath: /var/log/bindplane/bindplane.log
serverURL: http://192.168.1.10:3001
server:
  storageFilePath: /var/lib/bindplane/storage/bindplane.db
  secretKey: e124852a-49db-4318-99a8-76bd4aa80ba5
  sessionsSecret: 99112c19-9d87-4460-958c-a9affa874e21
  remoteURL: ws://192.168.1.10:3001
```

**Client Profile**

Create a profile named `basic`:

```bash
bindplanectl profile set basic \
  --username myuser \
  --password mypassword \
  --server-url http://192.168.1.10:3001

bindplanectl profile use basic
```

The a profile will be created at `~/.bindplane/profiles/basic.yaml`:

```yaml
username: myuser
password: mypassword
serverURL: http://192.168.1.10:3001
```

**Collector Manager Configuration**

```yaml
endpoint: http://192.168.1.10:3001/v1/opamp
secret_key: e124852a-49db-4318-99a8-76bd4aa80ba5
agent_id: ad3caa0c-ac90-4f8d-8691-2f43d9addc71
```

### TLS

BindPlane OP has support for server side TLS and mutual TLS.

What is a server? A server is the process running from the `bindplane serve` command.

What is a client?
- bindplane cli
- OpAMP collectors
- Web browsers

Keep in mind that all certificate files must be readable by the user running the bindplane, client,
and collector processes.

#### Server Side TLS

**Server Configuration**

Server side TLS is configured by setting `tlsCert` and `tlsKey` on the server. 

```yaml
host: 0.0.0.0
port: 3001
username: myuser
password: mypassword
logfilePath: /var/log/bindplane/bindplane.log
serverURL: https://bindplane-op.mydomain.net:3001
server:
  storeType: bbolt
  storageFilePath: /var/lib/bindplane/storage/bindplane.db
  secretKey: e124852a-49db-4318-99a8-76bd4aa80ba5
  sessionsSecret: 99112c19-9d87-4460-958c-a9affa874e21
  remoteURL: wss://bindplane-op.mydomain.net:3001
tlsCert: /etc/bindplane/tls/bindplane.crt
tlsKey: /etc/bindplane/tls/bindplane.key
```

Note that serverURL and remoteURL have a tls protocol set (`https` / `wss`).

**Client Profile**

All clients must trust the certificate authority that signed the server's 
certificate. This can be accomplished by setting `tlsCa` on the client or 
by importing the certificate authority into your operating system's trust store.

Create a profile named `tls`:

```bash
bindplanectl profile set tls \
  --username myuser \
  --password mypassword \
  --server-url http://192.168.1.10:3001 \
  --tls-ca /etc/bindplane/tls/my-corp-ca.crt

bindplanectl profile use tls
```

The a profile will be created at `~/.bindplane/profiles/tls.yaml`:

```yaml
username: myuser
password: mypassword
serverURL: https://bindplane-op.mydomain.net:3001
tlsCa:
  - /etc/bindplane/tls/my-corp-ca.crt
```

If the server's certificate authority is already imported into the client's operating system trust
store, it is not required to be set in the configuration.

Browsers will show a TLS warning unless the certificate authority is trusted by
your operating system.

**Collector Manager Configuration**

```yaml
endpoint: https://bindplane-op.mydomain.net:3001/v1/opamp
secret_key: e124852a-49db-4318-99a8-76bd4aa80ba5
agent_id: ad3caa0c-ac90-4f8d-8691-2f43d9addc71
tls_config:
  ca_file: /opt/observiq-otel-collector/tls/bindplane-ca.crt
```

If the server's certificate authority is already imported into the client's operating system trust
store, it is not required to be set in the configuration.

#### Mutual TLS

In this example, three certificate authorities are referenced:
- `my-corp-ca.crt`: Signed the server's certificate, must be trusted by all clients / collectors
- `client-ca.crt`: Signed all client certificates, must be set in the server configuration
- `collector-ca.crt`: Signed all collector certificates, must be set in the server configuration

**Server Configuration**

Mutual TLS is configured by setting `tlsCert`, `tlsKey`, and `tlsCa` on the server. 

```yaml
host: 0.0.0.0
port: 3001
username: myuser
password: mypassword
logfilePath: /var/log/bindplane/bindplane.log
serverURL: https://bindplane-op.mydomain.net:3001
server:
  storeType: bbolt
  storageFilePath: /var/lib/bindplane/storage/bindplane.db
  secretKey: e124852a-49db-4318-99a8-76bd4aa80ba5
  sessionsSecret: 99112c19-9d87-4460-958c-a9affa874e21
  remoteURL: wss://bindplane-op.mydomain.net:3001
tlsCert: /etc/bindplane/tls/bindplane.crt
tlsKey: /etc/bindplane/tls/bindplane.key
# Any client / collector certificate signed by one of these
# authorities will be trusted.
tlsCa:
  - /etc/bindplane/tls/client-ca.crt
  - /etc/bindplane/tls/collector-ca.crt
```

Note that serverURL and remoteURL have a tls protocol set (`https` / `wss`).

Note that mutliple certificate authorities can be specified. This example will trust
incoming connections from certificates signed by `client-ca` and `collector-ca`.

**Client Profile**

All clients must trust the certificate authority that signed the server's 
certificate. This can be accomplished by setting `tlsCa` on the client or 
by importing the certificate authority into your operating system's trust store.

Create a profile named `mtls`:

```bash
bindplanectl profile set mtls \
  --username myuser \
  --password mypassword \
  --server-url http://192.168.1.10:3001 \
  --tls-cert /etc/bindplane/tls/client.crt \
  --tls-key /etc/bindplane/tls/client.key \
  --tls-ca /etc/bindplane/tls/my-corp-ca.crt

bindplanectl profile use mtls
```

The a profile will be created at `~/.bindplane/profiles/mtls.yaml`:

```yaml
username: myuser
password: mypassword
serverURL: https://bindplane-op.mydomain.net:3001
tlsCert: /etc/bindplane/tls/client.crt
tlsKey: /etc/bindplane/tls/client.key
tlsCa:
  - /etc/bindplane/tls/my-corp-ca.crt
```

If the server's certificate authority is already imported into the client's operating system trust
store, it is not required to be set in the configuration.

Browsers will show a TLS warning unless the certificate authority is trusted by
your operating system.

**Collector Manager Configuration**

```yaml
endpoint: https://bindplane-op.mydomain.net:3001/v1/opamp
secret_key: e124852a-49db-4318-99a8-76bd4aa80ba5
agent_id: ad3caa0c-ac90-4f8d-8691-2f43d9addc71
tls_config:
  cert_file: /opt/observiq-otel-collector/tls/collector.crt
  key_file: /opt/observiq-otel-collector/tls/collector.crt
  ca_file: /opt/observiq-otel-collector/tls/bindplane-ca.crt
```

If the server's certificate authority is already imported into the client's operating system trust
store, it is not required to be set in the configuration.

## Client Profiles

The `profile` command offers a convenient way to create and use multiple client configurations.

In this example, it is assumed that the BindPlane server is running at `10.99.1.10` on port `3001`.

```bash
bindplanectl profile set remote --server-url https://10.99.1.10:3001
bindplanectl profile use remote
```

See `bindplanectl profile help` for more profile sub commands.
