# StanODN Infrastructure

Complete infrastructure setup for StanODN lab environment using Proxmox, Terraform, and Ansible.

## Architecture Overview

The infrastructure consists of:
- **Proxmox VE** - Virtualization platform
- **Terraform** - Infrastructure as Code for VM provisioning
- **Ansible** - Configuration management and automation
- **4 VMs** with specific roles:
  - `stan-runner01` - GitLab/Jenkins CI/CD Runner (2 CPU, 6GB RAM)
  - `stan-ns01` - DNS/DHCP Server (1 CPU, 4GB RAM)
  - `stan-gw01` - NGINX Gateway (1 CPU, 4GB RAM)
  - `stan-acn01` - Jenkins/Ansible Control Node (4 CPU, 16GB RAM)

## Prerequisites

- Physical machine with at least 8 CPU cores and 32GB RAM
- Ubuntu 22.04 LTS for the host system
- Network access for downloading packages and templates

## Step-by-Step Setup Guide

### 1. Prepare PXE Boot Configuration

Create a PXE boot directory with configuration files and environment variables:

```bash
# Create PXE boot directory structure
mkdir -p pxe-boot/{config,.env}
# Add your PXE configuration files here
```

### 2. Provision Proxmox Machine

1. **Install Proxmox VE** on your physical machine
2. **Configure storage** - Set up your disk layout (e.g., `disk1tb-01`)
3. **Configure networking** - Set up `vmbr0` bridge
4. **Note the Proxmox IP** - You'll need this for Terraform configuration

### 3. Create VM Templates with Ansible

```bash
cd ansible
# Create Ubuntu template (ID 110) with cloud-init
ansible-playbook playbooks/proxmox/create-templates.yml
```

### 4. Create Terraform API Token (Manual)

1. **Login to Proxmox Web UI**
2. **Go to Datacenter → Permissions → API Tokens**
3. **Create new token**:
   - User: `terraform@pam`
   - Token ID: `terraform`
   - Privilege Separation: `Yes`
   - **Save the token secret** - you'll need it for Terraform

### 5. Configure Terraform

```bash
cd terraform/proxmox/pve01
# Copy example configuration
cp terraform.tfvars.example terraform.tfvars
# Edit terraform.tfvars with your actual values:
# - proxmox_api_url
# - proxmox_api_token_secret
# - vm_template_id
# - vm_storage
# - ssh_public_keys
```

### 6. Provision VMs with Terraform

```bash
cd terraform/proxmox/pve01
terraform init
terraform plan
terraform apply
```

This will create all 4 VMs with the specified resources:
- Total: 8 CPU cores, 30.7GB RAM allocated

### 7. Setup GitLab Runner (Manual)

On `stan-runner01`:

```bash
# Download and configure GitLab runner
curl -LJO "https://gitlab-runner-downloads.s3.amazonaws.com/latest/deb/gitlab-runner_amd64.deb"
sudo dpkg -i gitlab-runner_amd64.deb
sudo gitlab-runner register

# Create systemd service
sudo tee /etc/systemd/system/gitlab-runner.service << EOF
[Unit]
Description=GitLab Runner
After=network.target

[Service]
User=ubuntu
WorkingDirectory=/home/ubuntu/actions-runner
ExecStart=/home/ubuntu/actions-runner/run.sh
KillMode=control-group
Restart=always

[Install]
WantedBy=multi-user.target
EOF

sudo systemctl enable gitlab-runner
sudo systemctl start gitlab-runner
```

### 8. Configure DNS Server

On `stan-ns01`:

1. **Install Docker Compose**:
```bash
sudo apt update
sudo apt install docker-compose
sudo adduser ubuntu docker
```

2. **Generate environment variables**:
```bash
# Use this vim command to generate random strings
:r !bash -lc 'randhex(){ n=$1; openssl rand -hex $(( (n+1)/2 )) | cut -c1-"$n"; }; printf "%s-%s-%s\n" "$(randhex 5)" "$(randhex 7)" "$(randhex 5)"'
```

3. **Create .env file** with the generated values

### 9. Disable Previous DHCP

1. **Stop and disable systemd-resolved**:
```bash
sudo systemctl stop systemd-resolved
sudo systemctl disable systemd-resolved
```

2. **Edit resolved.conf**:
```bash
sudo nano /etc/systemd/resolved.conf
# Add: DNSStubListener=no
```

### 10. Configure Static IP

On `stan-ns01`, set static IP to `192.168.1.55/24`:

```bash
sudo nano /etc/netplan/01-netcfg.yaml
# Configure static IP with gateway 192.168.1.1
sudo netplan apply
```

### 11. Update DNS Configuration

1. **Update Terraform DNS servers** in `terraform.tfvars`:
```hcl
dns_servers = ["192.168.1.55", "1.1.1.1"]
```

2. **Update GitLab/GitHub settings** to use `192.168.1.55` as DNS

### 12. Restart All Hosts

```bash
# Restart all VMs to apply new DNS settings
cd terraform/proxmox/pve01
terraform apply -replace="proxmox_virtual_environment_vm.ubuntu_vms"
```

### 13. Propagate SSH Keys

Ensure the runner's SSH key is available on all hosts for Ansible automation:

```bash
# Copy runner's public key to all hosts
ssh-copy-id ubuntu@stan-ns01
ssh-copy-id ubuntu@stan-gw01
ssh-copy-id ubuntu@stan-acn01
```

## Ansible Automation (Future TODOs)

The following steps are currently manual but can be automated with Ansible roles:

- [ ] **Terraform token creation** - Ansible role for API token management
- [ ] **GitLab runner setup** - Automated runner installation and configuration
- [ ] **DNS/DHCP server configuration** - Automated server setup
- [ ] **Docker Compose installation** - Automated Docker setup
- [ ] **Static IP configuration** - Automated network configuration

## Network Configuration

| VM | IP Address | Purpose |
|---|---|---|
| `stan-runner01` | DHCP | GitLab/Jenkins Runner |
| `stan-ns01` | 192.168.1.55/24 | DNS/DHCP Server |
| `stan-gw01` | 192.168.1.60/24 | NGINX Gateway |
| `stan-acn01` | DHCP | Jenkins/Ansible Control |

## Resource Allocation

| VM | CPU | RAM | Disk | Purpose |
|---|---|---|---|---|
| `stan-runner01` | 2 cores | 6GB | 120GB | CI/CD Runner |
| `stan-ns01` | 1 core | 4GB | 100GB | DNS/DHCP |
| `stan-gw01` | 1 core | 4GB | 100GB | NGINX Gateway |
| `stan-acn01` | 4 cores | 16GB | 200GB | Jenkins/Ansible |
| **Total** | **8 cores** | **30GB** | **520GB** | **Full Lab** |

## Troubleshooting

### Common Issues

1. **Terraform connection errors**: Check Proxmox API URL and token
2. **VM boot issues**: Verify cloud-init configuration and SSH keys
3. **Network connectivity**: Ensure DNS server is properly configured
4. **Ansible connection**: Verify SSH key propagation and user permissions

### Useful Commands

```bash
# Check VM status in Proxmox
qm list

# View Terraform state
terraform show

# Test Ansible connectivity
ansible all -m ping

# Check DNS resolution
nslookup google.com 192.168.1.55
```

## Maintenance

- **Regular backups**: Backup Terraform state and VM configurations
- **Updates**: Keep Proxmox, Terraform, and Ansible updated
- **Monitoring**: Monitor resource usage and VM performance
- **Security**: Regularly rotate SSH keys and API tokens

## Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Test thoroughly
5. Submit a pull request

## License


This project is licensed under the MIT License - see the LICENSE file for details. 
