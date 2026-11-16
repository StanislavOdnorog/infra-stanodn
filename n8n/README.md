# n8n on vpn-ge

URL: https://n8n.stanodn.org — A record `181.215.15.210`, DNS-only.
Docker Compose runs n8n and PostgreSQL on vpn-ge. Apache terminates HTTPS
using Let's Encrypt; certbot.timer renews the certificate and reloads Apache.
Existing VPN services and the default Apache site stay in place.

## Deployment

Push changes under `n8n/` to main or dispatch **Deploy N8N** in GitHub Actions.
The hosted runner updates only the n8n DNS record using the existing Cloudflare
secrets, uploads four deployment files, backs up existing data, and applies Compose.
The dedicated `N8N_VPN_GE_DEPLOY_KEY` is restricted in authorized_keys to
`/usr/local/sbin/n8n-ci-deploy` (source: `ci-receiver.sh`); no forwarding or PTY.
`N8N_VPN_GE_KNOWN_HOSTS` pins the server SSH host key.

Manual deployment after copying the four files to `/opt/n8n`:

```sh
ssh vpn-ge 'bash /opt/n8n/deploy.sh'
```

Initial provisioning requires SSH access, Ubuntu, an HTTP webroot served by Apache,
and DNS pointing at this host. Install the CI receiver separately as root, mode 700.
The dedicated workflow changes the live DNS record from `dns-records.json`;
the next Terraform refresh reconciles its state. Do not apply stale saved plans.

## State and access

- `/opt/n8n/postgres`: database, including workflows, accounts and credentials.
- `/opt/n8n/data`: n8n persistent files.
- `/opt/n8n/files`: workflow files.
- `/opt/n8n/.env`: generated database password and stable encryption key (0600).
- `/opt/n8n/owner-credentials.json`: initial owner login (0600), created before
  exposing HTTPS. Read via SSH; change the initial password after signing in.

Never delete these paths during upgrades. Images are explicitly pinned in Compose.
Only loopback port 5678 is published; PostgreSQL has no host port.
Webhooks and editor use the public HTTPS URL, timezone Asia/Tbilisi.

## Backups and recovery

Root cron at 04:15 **server local time**, plus each redeploy, runs `backup.sh`.
It briefly stops n8n, dumps PostgreSQL and archives data/files/secrets coherently.
Completed snapshots have a `complete` marker in `/opt/n8n/backups/<UTC timestamp>`.
Retention: 14 days. These are local backups and do not cover losing the server disk.

To restore: stop n8n, keep a copy of current state, restore `n8n-files.tgz` under
`/opt/n8n` as root, and restore `postgres.dump` into an empty n8n database with
`pg_restore -U n8n -d n8n --no-owner`. Start Compose and check `/healthz/readiness`.
Use matching application/database image versions; restore the encryption key
together with the database. Verify a backup before deleting any existing data.
