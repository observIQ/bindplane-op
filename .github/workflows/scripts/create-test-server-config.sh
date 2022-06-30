#!/usr/bin/env bash
# Copyright  observIQ, Inc.
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#      http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

set -e

cat <<EOF | sudo tee /etc/bindplane/config.yaml
host: 127.0.0.1
port: "3001"
serverURL: http://127.0.0.1:3001
username: admin
password: test
logFilePath: /var/log/bindplane/bindplane.log
server:
    storageFilePath: /var/lib/bindplane/storage/bindplane.db
    secretKey: $(uuidgen)
    remoteURL: ws://127.0.0.1:3001
    downloadsFolderPath: /var/lib/bindplane/downloads
    sessionsSecret: $(uuidgen)
    serverURL: http://127.0.0.1:3001
EOF
