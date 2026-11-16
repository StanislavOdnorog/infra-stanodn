#!/usr/bin/env bash
set -euo pipefail
umask 077
cd /opt/n8n
mkdir -p backups
dest="backups/$(date -u +%Y%m%dT%H%M%SZ)"
mkdir "$dest"
trap 'docker compose start n8n >/dev/null' EXIT
docker compose stop n8n >/dev/null
docker compose exec -T postgres pg_dump -U n8n -d n8n -Fc > "$dest/postgres.dump"
tar -czf "$dest/n8n-files.tgz" .env data files owner-credentials.json
docker compose start n8n >/dev/null
trap - EXIT
touch "$dest/complete"
find backups -mindepth 1 -maxdepth 1 -type d -mtime +14 -exec rm -rf -- {} +
