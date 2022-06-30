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

package_name="bindplane"

PREREQS="printf systemctl sed uname sudo curl"
INDENT_WIDTH='  '
indent=""

# Colors
num_colors=$(tput colors 2>/dev/null)
if test -n "$num_colors" && test "$num_colors" -ge 8; then
  reset="$(tput sgr0)"
  fg_cyan="$(tput setaf 6)"
  fg_green="$(tput setaf 2)"
  fg_red="$(tput setaf 1)"
  fg_yellow="$(tput setaf 3)"
fi

if [ -z "$reset" ]; then
  sed_ignore=''
else
  sed_ignore="/^[$reset]+$/!"
fi

printf() {
  if command -v sed >/dev/null; then
    command printf -- "$@" | sed -E "$sed_ignore s/^/$indent/g"  # Ignore sole reset characters if defined
  else
    # Ignore $* suggestion as this breaks the output
    # shellcheck disable=SC2145
    command printf -- "$indent$@"
  fi
}

increase_indent() { indent="$INDENT_WIDTH$indent" ; }
decrease_indent() { indent="${indent#*$INDENT_WIDTH}" ; }

# Color functions reset only when given an argument
# Ignore "parameters are never passed"
# shellcheck disable=SC2120
reset() { command printf "$reset$*$(if [ -n "$1" ]; then command printf "$reset"; fi)" ; }
fg_cyan() { command printf "$fg_cyan$*$(if [ -n "$1" ]; then command printf "$reset"; fi)" ; }
fg_green() { command printf "$fg_green$*$(if [ -n "$1" ]; then command printf "$reset"; fi)" ; }
fg_red() { command printf "$fg_red$*$(if [ -n "$1" ]; then command printf "$reset"; fi)" ; }
fg_yellow() { command printf "$fg_yellow$*$(if [ -n "$1" ]; then command printf "$reset"; fi)" ; }

# Intentionally using variables in format string
# shellcheck disable=SC2059
info() { printf "$*\\n" ; }

# Intentionally using variables in format string
# shellcheck disable=SC2059
error() {
  increase_indent
  printf "$fg_red$*$reset\\n"
  decrease_indent
}

# Intentionally using variables in format string
# shellcheck disable=SC2059
success() { printf "$fg_green$*$reset\\n" ; }

observiq_banner()
{
  fg_cyan "           888                                        8888888 .d88888b.\\n"
  fg_cyan "           888                                          888  d88P\" \"Y88b\\n"
  fg_cyan "           888                                          888  888     888\\n"
  fg_cyan "   .d88b.  88888b.  .d8888b   .d88b.  888d888 888  888  888  888     888\\n"
  fg_cyan "  d88\"\"88b 888 \"88b 88K      d8P  Y8b 888P\"   888  888  888  888     888\\n"
  fg_cyan "  888  888 888  888 \"Y8888b. 88888888 888     Y88  88P  888  888 Y8b 888\\n"
  fg_cyan "  Y88..88P 888 d88P      X88 Y8b.     888      Y8bd8P   888  Y88b.Y8b88P\\n"
  fg_cyan "   \"Y88P\"  88888P\"   88888P'  \"Y8888  888       Y88P  8888888 \"Y888888\"\\n"
  fg_cyan "                                                                   Y8b  \\n"

  reset
}

separator() { printf "===================================================\\n" ; }

banner() {
  printf "\\n"
  separator
  printf "| %s\\n" "$*" ;
  separator
}

usage() {
  increase_indent
  USAGE=$(cat <<EOF
Usage:
  $(fg_yellow '-v, --version')
      An optional BindPlane package version. Defaults to the latest version
      present in the package repository.
EOF
  )
  info "$USAGE"
  decrease_indent
  return 0
}

force_exit() {
  # Exit regardless of subshell level with no "Terminated" message
  kill -PIPE $$
  # Call exit to handle special circumstances (like running script during docker container build)
  exit 1
}

error_exit() {
  line_num=$(if [ -n "$1" ]; then command printf ":$1"; fi)
  error "ERROR ($SCRIPT_NAME$line_num): ${2:-Unknown Error}" >&2
  if [ -n "$0" ]; then
    increase_indent
    error "$*"
    decrease_indent
  fi
  force_exit
}

succeeded() {
  increase_indent
  success "Succeeded!"
  decrease_indent
}

failed() {
  error "Failed!"
}

root_check() {
  system_user_name=$(id -un)
  if [ "${system_user_name}" != 'root' ]
  then
    # If not root, ensure the running user has access to
    # sudo.
    if sudo -l | grep "${system_user_name}" >/dev/null; then
        return
    else
        failed
        error_exit "$LINENO" "Script needs to be run as root or with sudo"
    fi
  fi
}

os_check() {
  info "Checking that the operating system is supported..."
  os_type=$(uname -s)
  case "$os_type" in
    Linux)
      succeeded
      ;;
    *)
      failed
      error_exit "$LINENO" "The operating system $(fg_yellow "$os_type") is not supported by this script."
      ;;
  esac
}

os_arch_check() {
  info "Checking for valid operating system architecture..."
  arch=$(uname -m)
  case "$arch" in
    x86_64|amd64)
      arch=amd64
      ;;
    aarch64|arm64|aarch64_be|armv8b|armv8l)
      arch=arm64
      succeeded
      ;;
    *)
      failed
      error_exit "$LINENO" "The operating system architecture $(fg_yellow "$arch") is not supported by this script."
      ;;
  esac
}

package_manager_check() {
    if command -v apt-get >/dev/null; then
        succeeded
    elif command -v dnf >/dev/null; then
        succeeded
    elif command -v yum >/dev/null; then
        succeeded
    else
        error_exit "$LINENO" "Failed to detect Linux package manager. Supported package managers are 'apt-get' and 'yum'."
    fi
}

# This will check if the current environment has
# all required shell dependencies to run the installation.
dependencies_check() {
  info "Checking for script dependencies..."
  FAILED_PREREQS=''
  for prerequisite in $PREREQS; do
    if command -v "$prerequisite" >/dev/null; then
      continue
    else
      if [ -z "$FAILED_PREREQS" ]; then
        FAILED_PREREQS="${fg_red}$prerequisite${reset}"
      else
        FAILED_PREREQS="$FAILED_PREREQS, ${fg_red}$prerequisite${reset}"
      fi
    fi
  done

  if [ -n "$FAILED_PREREQS" ]; then
    failed
    error_exit "$LINENO" "The following dependencies are required by this script: [$FAILED_PREREQS]"
  fi
  succeeded
}

check_prereqs() {
  banner "Checking Prerequisites"
  increase_indent
  root_check
  os_check
  os_arch_check
  dependencies_check
  success "Prerequisite check complete!"
  decrease_indent
}

# latest_version gets the tag of the latest release, without the v prefix.
latest_version() {
  curl -sSL -H"Accept: application/vnd.github.v3+json" https://api.github.com/repos/observIQ/bindplane-op/releases/latest | \
    grep "\"tag_name\"" | \
    sed -E 's/ *"tag_name": "v([0-9]+\.[0-9]+\.[0-9+])",/\1/'
}

# download_url returns the url for downloading a package with
# the given version, arch, type. Package type is passed as an argument.
download_url() {
  package_type="$1"
  if [ -z "$package_type" ] ; then
    error_exit "$LINENO" "Package type not set"
  fi

  # Detect latest release if version not set
  if [ -z "$version" ] ; then
    version=$(latest_version)
  fi

  if [ -z "$version" ] ; then
    error_exit "$LINENO" "Could not determine version to install"
  fi

  # Example:
  #       https://github.com/observIQ/bindplane-op/releases/download/v0.0.47/bindplane_0.0.47_linux_amd64.deb
  url="https://github.com/observiq/bindplane-op/releases/download/v$version/${package_name}_${version}_linux_${arch}.${package_type}"
  printf "%s" "$url"
}

deb_install() {
    banner "Installing package ${package_name}"
    increase_indent

    url=$(download_url "deb")
    curl -s -o bindplane.deb "$url" || error_exit "$LINENO" "Failed to download BindPlane package from ${url}"

    if dpkg -s bindplane &>/dev/null; then
        sudo apt-get install --only-upgrade -y -f ./bindplane.deb  || error_exit "$LINENO" "Failed to upgrade BindPlane"
    else
        sudo apt-get install -y -f ./bindplane.deb  || error_exit "$LINENO" "Failed to install BindPlane"
    fi

    succeeded
    decrease_indent
}

dnf_install() {
    banner "Installing package ${package_name}"
    increase_indent

    url=$(download_url "rpm")
    curl -s -o bindplane.rpm "$url" || error_exit "$LINENO" "Failed to download BindPlane package from ${url}"

    if rpm -q bindplane &>/dev/null; then
        sudo dnf upgrade -y bindplane.rpm || error_exit "$LINENO" "Failed to upgrade BindPlane"
    else
        sudo dnf install -y bindplane.rpm || error_exit "$LINENO" "Failed to install BindPlane"
    fi

    succeeded
    decrease_indent
}

yum_install() {
    banner "Installing package ${package_name}"
    increase_indent

    url=$(download_url "rpm")
    curl -s -o bindplane.rpm "$url" || error_exit "$LINENO" "Failed to download BindPlane package from ${url}"

    if rpm -q bindplane &>/dev/null; then
        sudo yum upgrade -y bindplane.rpm || error_exit "$LINENO" "Failed to upgrade BindPlane"
    else
        sudo yum install -y bindplane.rpm || error_exit "$LINENO" "Failed to install BindPlane"
    fi

    succeeded
    decrease_indent
}

# run_service always enables and restarts BindPlane and is safe to
# call after initial install or upgrade.
run_service() {
    banner "Configuring BindPlane service"
    increase_indent

    sudo systemctl enable bindplane || error_exit "$LINENO" "Failed to enable BindPlane service"
    sudo systemctl restart bindplane || error_exit "$LINENO" "Failed to start BindPlane service"

    succeeded

    decrease_indent
}

install() {
    # package manager has already been verified by calling check_prereqs
    # which calls package_manager_check.
    if command -v apt-get >/dev/null; then
        deb_install
    elif command -v dnf >/dev/null; then
        dnf_install
    elif command -v yum >/dev/null; then
        yum_install
    fi

    run_service
}

display_results() {
    banner 'Information'
    increase_indent
    info "Config:               $(fg_cyan "/etc/bindplane/config.yaml")$(reset)"
    info "Start Command:        $(fg_cyan "sudo systemctl start bindplane")$(reset)"
    info "Stop Command:         $(fg_cyan "sudo systemctl stop bindplane")$(reset)"
    info "Server Logs Command:  $(fg_cyan "sudo tail -F /var/log/bindplane/bindplane.log")$(reset)"
    info "Access Logs Command:  $(fg_cyan "sudo journalctl -f --unit bindplane")$(reset)"
    decrease_indent

    banner 'Server Initialization'
    increase_indent
    info "To initialize the server, run: $(fg_cyan "sudo /usr/local/bin/bindplane init server --config /etc/bindplane/config.yaml")$(reset)"
    decrease_indent

    banner 'Support'
    increase_indent
    info "For more information on configuring BindPlane, see the docs: $(fg_cyan "https://github.com/observIQ/bindplane")$(reset)"
    info "If you have any other questions please contact us at $(fg_cyan support@observiq.com)$(reset)"
    decrease_indent

    banner "$(fg_green Installation Complete!)"
    return 0
}

main() {
  if [ $# -ge 1 ]; then
    while [ -n "$1" ]; do
      case "$1" in
        -v|--version)
          version=$2 ; shift 2
          ;;
        -h|--help)
          usage
          force_exit
          ;;
      --)
        shift; break ;;
      *)
        error "Invalid argument: $1"
        usage
        force_exit
        ;;
      esac
    done
  fi


  observiq_banner
  check_prereqs
  install
  display_results
}

main "$@"
