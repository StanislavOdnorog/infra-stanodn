#!/usr/bin/env bash
set -euo pipefail
umask 077
cd /opt/n8n
exec 9>/run/n8n-deploy.lock
flock -w 600 9
export DEBIAN_FRONTEND=noninteractive
if ! command -v docker >/dev/null || ! command -v certbot >/dev/null; then
  apt-get update -qq
  apt-get install -y docker.io docker-compose-v2 apache2 certbot
fi
systemctl enable --now docker apache2
install -d -m 700 /opt/n8n/backups
install -d -m 700 -o 1000 -g 1000 data files
if [ ! -f .env ]; then
  printf 'POSTGRES_PASSWORD=%s\nN8N_ENCRYPTION_KEY=%s\n' "$(openssl rand -hex 32)" "$(openssl rand -hex 32)" > .env
fi
chmod 600 .env
docker compose config -q
if [ -d postgres/base ]; then bash backup.sh; fi
docker compose pull --quiet
docker compose up -d --wait --wait-timeout 300
# Claim locally before exposing the setup page.
python3 - <<'PY'
import json, pathlib, secrets, urllib.request
base = 'http://127.0.0.1:5678/rest/'
settings = json.load(urllib.request.urlopen(base + 'settings'))['data']
if settings.get('userManagement', {}).get('showSetupOnFirstLoad'):
    path = pathlib.Path('owner-credentials.json')
    if not path.exists():
        path.write_text(json.dumps(dict(email='isodnorog@gmail.com', firstName='Stanislav', lastName='Odnorog', password=secrets.token_urlsafe(30)+'aA1!'))+'\n')
        path.chmod(0o600)
    request = urllib.request.Request(base+'owner/setup', data=path.read_bytes(), headers={'Content-Type':'application/json'})
    with urllib.request.urlopen(request) as response:
        assert response.status == 200
    print('Owner initialized; credentials in /opt/n8n/owner-credentials.json')
PY
ufw allow 80/tcp
ufw allow 443/tcp
mkdir -p /var/www/html/.well-known/acme-challenge
certbot certonly --webroot -w /var/www/html -d n8n.stanodn.org --non-interactive --agree-tos --email isodnorog@gmail.com --keep-until-expiring
a2enmod ssl proxy proxy_http headers rewrite
install -m 644 apache.conf /etc/apache2/sites-available/n8n.conf
a2ensite n8n
apache2ctl configtest
systemctl reload apache2
install -d /etc/letsencrypt/renewal-hooks/deploy
printf '#!/bin/sh\napache2ctl configtest && systemctl reload apache2\n' > /etc/letsencrypt/renewal-hooks/deploy/n8n-apache
chmod 755 /etc/letsencrypt/renewal-hooks/deploy/n8n-apache
systemctl enable --now certbot.timer
printf '15 4 * * * root flock -n /run/n8n-deploy.lock bash /opt/n8n/backup.sh > /var/log/n8n-backup.log 2>&1\n' > /etc/cron.d/n8n-backup
curl --fail --silent --show-error --resolve n8n.stanodn.org:443:127.0.0.1 https://n8n.stanodn.org/healthz/readiness
