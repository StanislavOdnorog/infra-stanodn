# New Server Onboarding

This checklist is for a host that already has your SSH key installed and is reachable from the environment where Ansible will run.

## Required Input

- `hostname`
- `IP address` or DNS name
- `ssh user` if it is not `ubuntu`
- server role, for example `generic`, `docker`, `gateway`, `vpn`, or `n8n`
- extra firewall ports to open
- extra packages or Docker requirement

## 1. Add the Host to Inventory

Add the new host to [ansible/inventory/static.yml](../ansible/inventory/static.yml).

Example:

```yaml
all:
  children:
    ungrouped:
      hosts:
        new-app-01:
          ansible_host: 203.0.113.10
```

If the host belongs in an existing group such as `wireguard_servers` or `shadowsocks_servers`, place it there instead of `ungrouped`.

## 2. Create Host Variables

Create a directory:

```bash
mkdir -p ansible/inventory/host_vars/new-app-01
```

Start with these snippets:

- [ansible/inventory/host_vars/_snippets/basics.yml](../ansible/inventory/host_vars/_snippets/basics.yml)
- [ansible/inventory/host_vars/_snippets/firewall.yml](../ansible/inventory/host_vars/_snippets/firewall.yml)
- [ansible/inventory/host_vars/_snippets/packages.yml](../ansible/inventory/host_vars/_snippets/packages.yml)
- [ansible/inventory/host_vars/_snippets/users.yml](../ansible/inventory/host_vars/_snippets/users.yml)

Typical minimum:

```yaml
# ansible/inventory/host_vars/new-app-01/basics.yml
---
ansible_user: ubuntu
hostname: "new-app-01"
```

For a generic server, `users.yml` is usually optional because defaults already come from `group_vars/all/users.yml`.

## 3. Validate Access

```bash
make ping TARGET=new-app-01
```

If this fails locally because of SSH routing or auth, switch to the remote operator environment before retrying.

## 4. Apply Base Preparation

```bash
make server-init TARGET=new-app-01
```

This runs:

- user sync and SSH authorized keys
- hostname setup
- DNS resolver setup
- package install and upgrade
- UFW rules
- hardening for SSH and fail2ban
- basic system optimization

## 5. Optional Follow-up

Install Docker:

```bash
make docker-install TARGET=new-app-01
```

Deploy a Docker app from `docker-apps/`:

```bash
make app-deploy TARGET=new-app-01 APP_DIR=reverse-proxy
```

## Notes

- Static inventory is the safest default for local work because it does not require the encrypted Proxmox inventory.
- The base hardening disables password auth in `sshd_config`, so key-based access must work before `server-init`.
- Host-specific ports belong in `ansible/inventory/host_vars/<host>/firewall.yml`.
