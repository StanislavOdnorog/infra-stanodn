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
Create a new app directory at the repo root (same level as `n8n/`, `jenkins/`, etc.):
```
 my-app/
   docker-compose.yml
   .env.example   # recommended
```

Notes:
- Keep everything self‑contained in the app folder.
- If the app needs persistent data, define named volumes in the compose file.
- **.env is set manually on the host after deploy.**

---

## 2) Ensure Docker is Installed on Target
If the target host does not have Docker/Compose, run:
```
ansible-playbook ansible/playbooks/general/install/docker.yml \
  --extra-vars "target=<host-or-group>"
```

> Inventory is **vault‑encrypted**; Jenkins pipeline (see below) is easiest for running playbooks.

---

## 3) Deploy the App (GitHub Actions)
Use the **generic** workflow: `Deploy Generic Service` (new workflow).

Inputs:
- `service_name` = folder name on the server (e.g. `my-app`)
- `source_path` = repo path to deploy (e.g. `my-app`)
- `server_host` = SSH host (e.g. `n8n`)

The action:
- SCPs the folder to the target host
- Runs `docker-compose down` + `docker-compose up -d`

**Manual alternative** (SSH to host):
```
cd ~/my-app
sudo docker-compose up -d
```

> After deploy, SSH to the host and fill in `.env` (it’s not auto‑created).

---

## 4) Expose via Reverse Proxy (stan-gw01)
Add a new server file in:
```
reverse-proxy/nginx/conf.d/servers/<app>.stanodn.org.conf
```
Use the existing configs (e.g., `n8n.stanodn.org.conf`) as a template.
Make sure `proxy_pass` points to the **lab host**, e.g.:
```
proxy_pass http://n8n.stanodn.lab:<PORT>;
```

Reload nginx on the gateway:
```
ssh stan-lab-gw01
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

## 6) GitHub Actions (Primary Path)
Use the **generic** workflow: `.github/workflows/deploy-generic.yml`.
It wraps the composite action `.github/actions/deploy-docker-compose` and works for any folder.

Use Actions for routine deploys; run playbooks locally only when Actions are unavailable.

## 7) Jenkins Pipeline (Secondary)
There is also a Jenkins pipeline in `jenkinsfiles/Jenkinsfile.ansible-playbooks` for running Ansible playbooks with vault access.

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
