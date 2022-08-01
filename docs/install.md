# Installation

  * [Server](#server)
    + [Linux](#linux)
    + [Docker](#docker)
  * [Client](#client)
    + [Linux](#linux-1)
    + [Windows](#windows)
    + [macOS](#macos)
  * [Collector](#collector)
    + [Web Interface](#web-interface)
    + [Bindplane CLI](#bindplane-cli)
    + [Manual Install](#manual-install)
      - [Linux](#linux-2)
      - [Windows](#windows-1)

BindPlane OP is comprised of three components:

- Server
- Client CLI
- Collector(s)

## Server

### Linux

The BindPlane server has two installation methods. Once installed, the following will be present:

- Systemd service `bindplane` located at `/usr/lib/systemd/system/bindplane.service`.
- System user `bindplane` with group `bindplane`. This user does not have a login shell, and is meant to run the `bindplane` service.
- Configuration at `/etc/bindplane/config.yaml`
- Persistent storage at `/var/lib/bindplane`
- Logging at `/var/log/bindplane/bindplane.log`. The log file is rotated automatically by the `bindplane` process.

Modifications to the service file can be done with a [systemd override](https://wiki.archlinux.org/title/Systemd#Drop-in_files).

The service can be started, stopped, and restarted using `sudo systemctl [start|stop|restart] bindplane`

See the [configuration documentation](./configuration.md#configuration) for detailed server configuration instructions.

**Script**

The installation script provides a quick and easy way to install BindPlane OP. The script
detects the latest release, Linux platform, installs the correct package. If you do not wish to use the
script, see the `package` option.

```bash
curl -fsSlL https://github.com/observiq/bindplane-op/releases/latest/download/install-linux.sh | bash -s --
```

**Package**

Packages can be downloaded and installed from a Github Release.

1. Download a bindplane package (rpm or deb) from the [releases page](https://github.com/observIQ/bindplane-op/releases)
2. Install the package
3. Enable and start the server

On RHEL based platforms:

```bash
sudo dnf install https://github.com/observIQ/bindplane-op/releases/download/v0.5.0/bindplane_0.5.0_linux_amd64.rpm
sudo systemctl enable --now bindplane
```

On Debian based platforms:

```bash
curl -L -o bindplane.deb https://github.com/observIQ/bindplane-op/releases/download/v0.5.0/bindplane_0.5.0_linux_amd64.deb
sudo apt install -f ./bindplane.deb
sudo systemctl enable --now bindplane
```

### Docker

BindPlane server can run as a container. Persistent data is stored in a volume
named `bindplane`.

```bash
docker volume create bindplane

docker run -d \
    --name bindplane \
    --restart always \
    --mount source=bindplane,target=/data \
    -e BINDPLANE_CONFIG_USERNAME=admin \
    -e BINDPLANE_CONFIG_PASSWORD=admin \
    -e BINDPLANE_CONFIG_SERVER_URL=http://localhost:3001 \
    -e BINDPLANE_CONFIG_REMOTE_URL=ws://localhost:3001 \
    -e BINDPLANE_CONFIG_SESSIONS_SECRET=2c23c9d3-850f-4062-a5c8-3f9b814ae144 \
    -e BINDPLANE_CONFIG_SECRET_KEY=8a5353f7-bbf4-4eea-846d-a6d54296b781 \
    -e BINDPLANE_CONFIG_LOG_OUTPUT=stdout \
    -p 3001:3001 \
    observiq/bindplane:latest
```

Be sure to replace username, password, session secret, and secret key environment
variables with your own unique values.

Server URL and Remote URL should be set to the docker host's hostname or IP address.

## Client

The `bindplanectl` command can be installed on Linux, Windows, and macOS. See the 
[client profile configuration documentation](./configuration.md#client-profiles) for configuration instructions.

### Linux

Packages can be downloaded and installed from a Github Release.

1. Download a bindplanectl package (rpm or deb) from the [releases page](https://github.com/observIQ/bindplane-op/releases)
2. Install the package
3. Create a profile

On RHEL based platforms:

```bash
sudo dnf install https://github.com/observIQ/bindplane-op/releases/download/v0.5.0/bindplanectl_0.5.0_linux_amd64.rpm
```

On Debian based platforms:

```bash
curl -L -o bindplanectl.deb https://github.com/observIQ/bindplane-op/releases/download/v0.5.0/bindplanectl_0.5.0_linux_amd64.deb
sudo apt install -f ./bindplanectl.deb
```

Once installed, create a [client profile](./configuration.md#client-profiles).

### Windows

Binary releases can be downloaded from a Github Release.

1. Download a `bindplanectl` binary from the [releases page](https://github.com/observIQ/bindplane-op/releases)
2. Extract
3. Create a profile

Example steps:

1. Download a Windows Release: https://github.com/observIQ/bindplane-op/releases/download/v0.5.0/bindplane-v0.5.0-windows-amd64.zip
2. Extract the zip file
3. Run `bindplanectl.exe` from command prompt or PowerShell.
4. Optionally [add the executable to your path](https://docs.microsoft.com/en-us/previous-versions/office/developer/sharepoint-2010/ee537574(v=office.14))

Once installed, create a [client profile](./configuration.md#client-profiles).

### macOS

The macOS client can be installed using homebrew. Follow the instructions [here](https://github.com/observIQ/homebrew-bindplane-op).

Once installed, create a [client profile](./configuration.md#client-profiles).

## Collector

Collectors can be instaled via the BindPlane web interface or manually. The manual approach could be implemented
with configuration management using tools such as Chef and Ansible, if you are using such tools.

### Web Interface

1. Log into the BindPlane web interface
2. Click "Install Agents"
3. Choose your target operating system

Copy the command generated by the web interface, and run it on the system you wish to install the collector on.

After running the collector install command, the collector will appear in the web interface under the Agents tab.

### Bindplane CLI

If you wish to use `bindplanectl` to generate your install command, you can run the following:

```bash
bindplanectl install agent
```
or specify a platform
```
bindplanectl install agent --platform linux
```

Available platforms can be found with `bindplanectl install agent --help`.

Copy the output and run the command on the system your wish to install the collector on.

After running the collector install command, the collector will appear in the `bindplanectl get agents` output.

### Shell Completion

#### Linux: bash

1. Verify that `bash-completion` is installed on the host
2. `bindplanectl completion bash | sudo tee -a /etc/bash_completion.d/bindplanectl` appends the output to a file in the bash completion directory
3. Restart the shell

#### macOS/Linux: ZSH

To setup zsh completion for bindplanectl on MacOS:
1. Include the following lines in `~/.zshrc`&nbsp;
```
autoload -Uz compinit
compinit
```
2. Locate `fpath` by running `echo $fpath`, there may be several listed, some may not exist, use an existing one in the next step.
3. Run the following command to generate the zsh tab completion script.\
`bindplanectl completion zsh ><YOUR FPATH HERE>/_bindplanectl`
4. Restart zsh and the bindplanectl tab completions will be available.

### Manual Install

Installing manually can be desired if you wish to avoid running shell scripts, or you require
a more flexable approach to installation.

The steps documented here can be used as a starting point if implementing collector installation
with configuration management. 

#### Linux

Follow the [collector installation guide](https://github.com/observIQ/observiq-otel-collector/blob/main/docs/installation-linux.md).

Once installed, create a manager configuration at `/opt/observiq-otel-collector/manager.yaml`. 

The manager configuration consists of the following required parameters:
- endpoint: The websocket URL used to connect to bindplane (with `/v1/opamp` as the path)
- secret_key: The secret key configured on the BindPlane server
- agent_id: A randomly generated UUIDv4, unique to this agent

Create manager.yaml

```bash
cat << EOF | sudo tee /opt/observiq-otel-collector/manager.yaml
endpoint: ws://localhost:3001/v1/opamp
secret_key: b1a71608-e80e-46dd-bc51-59c5a5634d25
agent_id: ad3caa0c-ac90-4f8d-8691-2f43d9addc71
EOF

sudo chown observiq-otel-collector:observiq-otel-collector /opt/observiq-otel-collector/manager.yaml
sudo chmod 0600 /opt/observiq-otel-collector/manager.yaml
```

Restart the collector

```bash
sudo systemctl restart observiq-otel-collector
```

Once the collector is restarted, it will connect to BindPlane and appear in the list of agents. BindPlane will preserve the collector's
existing configuration.

#### Windows

Follow the [collector installation guide](https://github.com/observIQ/observiq-otel-collector/blob/main/docs/installation-windows.md).

Once installed, create a manager Configuration at `C:\Program Files\observIQ OpenTelemetry Collector\manager.yaml`.

The manager configuration consists of the following required parameters:
- endpoint: The websocket URL used to connect to bindplane (with `/v1/opamp` as the path)
- secret_key: The secret key configured on the BindPlane server
- agent_id: A randomly generated UUIDv4, unique to this agent

A manager.yaml configuration should look like this:

```yaml
endpoint: ws://localhost:3001/v1/opamp
secret_key: b1a71608-e80e-46dd-bc51-59c5a5634d25
agent_id: ad3caa0c-ac90-4f8d-8691-2f43d9addc71
```

Restart the service (powershell)

```ps
Restart-Service -Name "observiq-otel-collector"
```

Once the collector is restarted, it will connect to BindPlane and appear in the list of agents. BindPlane will preserve the collector's
existing configuration.
