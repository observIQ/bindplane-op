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

FROM debian:11.3-slim

ENV USER=bindplane
ENV UID=10001
RUN groupadd "$USER"
RUN useradd \
    --shell /sbin/nologin \
    --system "$USER" -g "$USER" -u "$UID"

COPY bindplane /bindplane

# Default home is /data. A volume should be mounted here in order
# to persist data.
RUN mkdir /data
RUN chown bindplane:bindplane /data
RUN chmod 0750 /data
ENV BINDPLANE_CONFIG_HOME="/data"
ENV BINDPLANE_CONFIG_LOG_OUTPUT="stdout"

# Bind to all interfaces and use port 3001
ENV BINDPLANE_CONFIG_HOST=0.0.0.0
ENV BINDPLANE_CONFIG_PORT="3001"
EXPOSE 3001

USER bindplane

ENTRYPOINT [ "/bindplane", "serve" ]
