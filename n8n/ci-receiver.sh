#!/usr/bin/env bash
set -euo pipefail
umask 077
stage=$(mktemp -d /opt/n8n-release.XXXXXX)
trap 'rm -rf "$stage"' EXIT
tar -xz --no-same-owner -C "$stage" docker-compose.yml apache.conf deploy.sh backup.sh
install -m 600 "$stage/"* /opt/n8n/
bash /opt/n8n/deploy.sh
