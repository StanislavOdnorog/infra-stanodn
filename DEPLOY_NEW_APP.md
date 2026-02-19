# Deploying a New Application (StanODN Lab)

This repo ships apps via **Docker Compose** + **Ansible** and exposes them through the **NGINX reverse proxy** (stan-gw01) with DNS managed by **Terraform Cloudflare**.

## Quick Map
- **Apps** live in: `docker-apps/<app-name>/`
- **Deploy** via Ansible: `ansible/playbooks/general/install/custom-app.yml`
- **Reverse proxy**: `reverse-proxy/nginx/conf.d/servers/*.conf` (runs on `stan-gw01`)
- **DNS**: `terraform/cloudflare/stanodn.org/dns-records.json`

Hosts (aliases from `projects/.stan-lab-list-of-hosts`):
- `stan-acn-lab` → control/ansible node
- `stan-lab-gw01` → nginx gateway
- `stan-lab-n8n` → app host (example)

---

## 1) Create the App Folder
Create a new app directory in `docker-apps/`:
```
 docker-apps/
   my-app/
     docker-compose.yml
     .env.example   # recommended
```

Notes:
- Keep everything self‑contained in the app folder.
- If the app needs persistent data, define named volumes in the compose file.

---

## 2) Ensure Docker is Installed on Target
If the target host does not have Docker/Compose, run:
```
ansible-playbook ansible/playbooks/general/install/docker.yml \
  --extra-vars "target=<host-or-group>"
```

> Inventory is **vault‑encrypted**; Jenkins pipeline (see below) is easiest for running playbooks.

---

## 3) Deploy the App to a Host
Use the custom app deploy playbook:
```
ansible-playbook ansible/playbooks/general/install/custom-app.yml \
  --extra-vars "app_dir=my-app target=<host-or-group>"
```

What it does (see `ansible/roles/configure/docker-deploy`):
- Packs `docker-apps/<app>` locally
- Uploads to `/opt/docker-apps/<app>` on the target
- Copies `.env.example → .env` if missing
- Runs `docker-compose down` + `docker-compose up -d`

**Manual alternative** (SSH to host):
```
cd /opt/docker-apps/my-app
sudo docker-compose up -d
```

---

## 4) Expose via Reverse Proxy (stan-gw01)
Add a new server file in:
```
reverse-proxy/nginx/conf.d/servers/<app>.stanodn.org.conf
```
Use the existing configs (e.g., `n8n.stanodn.org.conf`) as a template.
Make sure `proxy_pass` points to the app’s internal host/port.

Reload nginx on the gateway:
```
ssh stan-lab-gw01
cd /opt/docker-apps/reverse-proxy  # or wherever it is deployed
sudo docker-compose exec nginx nginx -s reload
```

SSL certs are expected at `reverse-proxy/ssl/origin.pem` + `origin.key` (Cloudflare origin certs).

---

## 5) Add DNS Record (Cloudflare via Terraform)
Edit:
```
terraform/cloudflare/stanodn.org/dns-records.json
```
Add an `A` or `CNAME` for your subdomain.
Then apply:
```
cd terraform/cloudflare/stanodn.org
terraform init
terraform apply
```

---

## 6) Jenkins Pipeline (Primary Path)
There is already a Jenkins pipeline in `jenkinsfiles/Jenkinsfile.ansible-playbooks` and this is the **preferred/standard** way to deploy:
- discovers playbooks + inventory
- runs with vault password from Jenkins credentials
- avoids local vault handling

Use it for all routine deploys; run playbooks locally only when Jenkins is unavailable.

---

## Deployment Checklist
- [ ] App folder in `docker-apps/` with compose + env
- [ ] Docker installed on target host
- [ ] App deployed via Ansible playbook
- [ ] Reverse proxy config added + nginx reloaded
- [ ] DNS record added + applied
- [ ] Test HTTP/HTTPS endpoint

---

## Notes / Gaps to Fill
- Confirm exact deploy path on `stan-gw01` for reverse-proxy
- Confirm where Ansible inventory/vault password lives when running locally
