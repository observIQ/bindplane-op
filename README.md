<p align="center"><img src="docs/images/bindplaneop.png?raw=true"></p>

<center>

[![Continuous Integration](https://github.com/observIQ/bindplane-op/actions/workflows/ci.yml/badge.svg)](https://github.com/observIQ/bindplane-op/actions/workflows/ci.yml)
[![License](https://img.shields.io/badge/License-Apache_2.0-blue.svg)](https://opensource.org/licenses/Apache-2.0)

</center>

# BindPlane OP

BindPlane OP is an open source observability pipeline that gives you the ability to collect, refine, and ship metrics, logs, and traces to any destination. BindPlane OP provides the controls you need to reduce observability costs and simplify the deployment and management of telemetry agents at scale.

## Features

  * Manage the lifecycle of telemetry agents, starting with the [observIQ OpenTelemetry Collector](https://github.com/observIQ/observiq-otel-collector)
  * Build, deploy, and manage telemetry configurations for different Sources and deploy them to your agents
  * Ship metric, log, and trace data to one or many Destinations
  * Utilize flow controls to adjust the flow of your data in realtime

## Architecture

BindPlane OP is a lightweight web server (no dependencies) that you can deploy anywhere in your environment. It's composed of the following components:

  * GraphQL Server: provides configuration and agent details via GraphQL
  * REST Server: BindPlane CLI and UI make requests to the server via REST
  * WebSocket Server: Agents connect to receive configuration updates via [OpAMP](https://github.com/open-telemetry/opamp-spec)
  * Store: pluggable storage manages configuration and Agent state
  * Manager: dispatches configuration changes to Agents

## Support

BindPlane OP does not have any dependencies and can run on Windows, Linux, or macOS although we recommend installing BindPlane OP server one of the following supported Linux distributions:

  * Red Hat, Centos, Oracle Linux 7 and 8
  * Debian 10 and 11
  * Ubuntu LTS 18.04, 20.04
  * Suse Linux 12 and 15
  * Alma and Rocky Linux


The BindPlane CLI (included in the BindPlane OP binary) will run on Linux, Windows, and macOS. For a detailed list of commands and installation instructions for remote clients, see our [CLI](https://docs.bindplane.observiq.com/docs/cli) documentation page.

# Quick Start

The following are our recommended installation options. For more details on installation, please visit our [Getting Started](https://docs.bindplane.observiq.com/docs/getting-started) page or [Installation](https://docs.bindplane.observiq.com/docs/installation) page.

## Server

To install BindPlane Server, we recommend using our single-line installer. Alternatively, packages are available for download on our [releases](https://github.com/observIQ/bindplane-op/releases) page.

### Linux
```bash
curl -fsSlL https://github.com/observiq/bindplane-op/releases/latest/download/install-linux.sh | bash -s --
```

After the installer finishes, initialize the server using the `init` command. The specific command for your system is found in the installer output under the `Server Initialization` section. By default it is the following command for Linux:

```bash
sudo /usr/local/bin/bindplane init server --config /etc/bindplane/config.yaml
```

## Client

To install BindPlane CLI on macOS or Linux, we recommend using the following installation commands. Alternatively, packages are available for download on our [releases](https://github.com/observIQ/bindplane-op/releases) page.

### Linux
```bash
curl -fsSlL https://github.com/observiq/bindplane-op/releases/latest/download/install-linux.sh | bash -s --
```

### macOS
```bash
curl -fsSlL https://github.com/observiq/bindplane-op/releases/latest/download/install-macos.sh | bash -s --
```

## Configuration

The configuration of BindPlane OP is best done through the UI which can be accessed via a web browser on port 3001. The URL will be `http://<IP Address>:3001` with IP Address being the IP of the BindPlane server. To log in, use the credentials you specified when running the init command.

For more information on configuration, view our [Configuration](https://docs.bindplane.observiq.com/docs/configuration) documentation page.

# Community

BindPlane OP is an open source project. If you'd like to contribute, take a look at our [contribution guidelines](/docs/CONTRIBUTING.md). We look forward to building with you.

## Code of Conduct

BindPlane OP follows the [CNCF Code of Conduct](https://github.com/cncf/foundation/blob/main/code-of-conduct.md). Please report violations of the Code of Conduct to any or all maintainers.

# Other questions?

Join the conversation at the [BindPlane Slack](https://observiq.com/support-bindplaneop/)!

You can also check out our [documentation](https://docs.bindplane.observiq.com/), send us an [email](mailto:support.observiq.com), or open an issue with your question. We'd love to hear from you!
