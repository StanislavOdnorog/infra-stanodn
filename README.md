# StanODN Infra

Operational infrastructure repository for Ansible-managed hosts, Proxmox template work, Docker app deployment, and supporting Terraform assets.

The repo is optimized around two day-to-day flows:

- prepare and maintain manually managed servers through static inventory
- manage Proxmox-related automation and application deployment from the same workspace

## Repository Layout

- `ansible/`
  - `inventory/static.yml` for manually managed hosts
  - `inventory/proxmox.yml` for encrypted Proxmox inventory access
  - `inventory/vpn.yml` for the VPN deployment inventory
  - `group_vars/` and `host_vars/` for defaults and per-host overrides
  - `playbooks/general/configure/server/init.yml` for base server preparation
  - `playbooks/general/install/docker.yml` for Docker installation
  - `playbooks/general/install/custom-app.yml` for deploying folders from `docker-apps/`
  - `playbooks/proxmox/template/ubuntu.yml` for the Ubuntu cloud template workflow
- `docker-apps/` application payloads for Docker-based deployments
- `terraform/` Terraform code and state related to Proxmox and DNS
- `reverse-proxy/`, `n8n/`, `semaphore/`, `pxe-boot/`, and others for service-specific assets
- `docs/NEW_SERVER.md` short checklist for onboarding a new server

## Prerequisites

- `ansible` installed locally
- SSH access to the target host
- the private key referenced by [ansible.cfg](ansible.cfg), or an equivalent override in your environment
- for Proxmox inventory work, access to the vault password used to decrypt `ansible/inventory/proxmox.yml`

By default, [ansible.cfg](ansible.cfg) uses:

```ini
inventory = ansible/inventory/proxmox.yml, ansible/inventory/static.yml
private_key_file = ~/.ssh/home-pc/private
```

For most local operational work, use the static inventory explicitly or the `Makefile` targets below. That avoids vault warnings from the encrypted Proxmox inventory when you are only working with normal hosts.

## Quick Start

List the available helper commands:

```bash
make help
```

Show the static inventory graph:

```bash
make inventory-graph
```

Check SSH connectivity:

```bash
make ping TARGET=vpn-ge
```

Prepare a server:

```bash
make server-init TARGET=vpn-ge
```

Install Docker:

```bash
make docker-install TARGET=vpn-ge
```

Deploy a Docker app:

```bash
make app-deploy TARGET=stan-gw01 APP_DIR=reverse-proxy
```

Deploy the VPN stack:

```bash
make vpn-install
```

Run syntax checks:

```bash
make check
```

## Common Workflows

### New Server

Use [docs/NEW_SERVER.md](docs/NEW_SERVER.md).

In short:

1. Add the host to `ansible/inventory/static.yml`.
2. Create `ansible/inventory/host_vars/<host>/`.
3. Start from the `_snippets/` files.
4. Run `make ping TARGET=<host>`.
5. Run `make server-init TARGET=<host>`.
6. Optionally run `make docker-install TARGET=<host>`.

### Base Server Preparation

The main onboarding playbook is [ansible/playbooks/general/configure/server/init.yml](ansible/playbooks/general/configure/server/init.yml).

It applies:

- user sync and SSH authorized keys
- hostname configuration
- DNS resolver configuration
- package updates and base packages
- UFW rules
- SSH hardening and fail2ban
- system optimization

Defaults come from:

- [ansible/inventory/group_vars/all/users.yml](ansible/inventory/group_vars/all/users.yml)
- [ansible/inventory/group_vars/all/packages.yml](ansible/inventory/group_vars/all/packages.yml)
- [ansible/inventory/group_vars/all/firewall.yml](ansible/inventory/group_vars/all/firewall.yml)
- [ansible/inventory/group_vars/all/resolved.yml](ansible/inventory/group_vars/all/resolved.yml)

Per-host overrides belong in `ansible/inventory/host_vars/<host>/`.

### Docker Host Setup

Install Docker and Docker Compose:

```bash
make docker-install TARGET=<host>
```

Then deploy an app from `docker-apps/<app>`:

```bash
make app-deploy TARGET=<host> APP_DIR=<app>
```

### Proxmox Template Workflow

Create or update the Ubuntu template on a Proxmox host:

```bash
make template-ubuntu TARGET=pve01
```

Relevant defaults live in [ansible/roles/provision/proxmox/image-templates/ubuntu/defaults/main.yml](ansible/roles/provision/proxmox/image-templates/ubuntu/defaults/main.yml).

## Inventory Model

- `static.yml` is for normal hosts and is the default local entrypoint.
- `proxmox.yml` is encrypted and intended for Proxmox-aware inventory access.
- `vpn.yml` is dedicated to the split VPN deployment flow.

Useful snippet directories:

- `ansible/inventory/group_vars/_snippets/`
- `ansible/inventory/host_vars/_snippets/`

## Notes

- The repo may contain local operational changes in inventory and roles. Check `git status` before pulling or pushing.
- If a local Ansible run fails due to SSH routing or environment-specific auth, switch to the remote operator environment rather than retrying blindly from macOS.
- The base security role disables password authentication for SSH, so confirm key-based access before running `server-init`.
