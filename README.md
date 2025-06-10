# StanODN Infrastructure

A comprehensive infrastructure-as-code project for managing a complete homelab environment using Docker, Ansible, Jenkins, and Nginx reverse proxy. This repository provides automated deployment, configuration management, and monitoring for various services including VPN, DNS filtering, CI/CD, and observability stack.

## Architecture Overview

The StanODN infrastructure is designed as a modular, containerized homelab setup that includes:

- **Reverse Proxy Layer**: Nginx-based load balancer with SSL termination
- **VPN Access**: Pritunl VPN server for secure remote access
- **DNS & Network Security**: Pi-hole for ad-blocking and DNS management
- **CI/CD Pipeline**: Jenkins for automated deployments
- **Monitoring Stack**: Prometheus + Grafana for observability
- **Infrastructure Automation**: Ansible for configuration management
- **Virtualization Integration**: Proxmox VE cluster management

## Project Structure

```
stanodn-infra/
├── ansible/                    # Infrastructure automation and configuration management
│   ├── playbooks/             # Ansible playbooks for various tasks
│   │   ├── general/           # General server management playbooks
│   │   │   ├── docker-deploy.yml      # Deploy Docker applications
│   │   │   ├── docker-install.yml     # Install Docker and Docker Compose
│   │   │   ├── ping.yml               # Infrastructure health checks
│   │   │   └── server_sync_users.yml  # User synchronization
│   │   └── proxmox/           # Proxmox-specific automation
│   │       ├── vm/            # Virtual machine management
│   │       └── template/      # VM template management
│   ├── inventory/             # Host and group definitions
│   │   ├── static.yml         # Static infrastructure hosts
│   │   ├── proxmox.yml        # Dynamic Proxmox inventory
│   │   └── *.example          # Example configuration files
│   ├── group_vars/            # Ansible group variables
│   │   ├── all.yml            # Global variables
│   │   ├── users.yml          # User management configuration
│   │   └── cloud-init.yml     # Cloud-init VM configuration
│   ├── roles/                 # Custom Ansible roles
│   │   ├── configure/         # Configuration roles
│   │   │   ├── docker/        # Docker installation and setup
│   │   │   ├── docker-deploy/ # Application deployment automation
│   │   │   └── sync_users/    # User synchronization
│   │   └── provision/         # Infrastructure provisioning
│   │       └── proxmox/       # Proxmox VM and template provisioning
│   └── ansible.cfg            # Ansible configuration
├── docker-apps/               # Containerized applications
│   └── monitoring/            # Monitoring stack (Prometheus + Grafana)
│       ├── prometheus/        # Prometheus configuration
│       ├── grafana/          # Grafana configuration and dashboards
│       └── docker-compose.yml # Monitoring stack deployment
├── reverse-proxy/             # Nginx reverse proxy configuration
│   ├── nginx/                # Nginx configuration files
│   │   ├── conf.d/           # Nginx configuration directory
│   │   │   ├── servers/      # Server-specific configurations
│   │   │   ├── common/       # Shared configuration snippets
│   │   │   └── ssl/          # SSL configuration
│   │   └── nginx.conf        # Main Nginx configuration
│   ├── ssl/                  # SSL certificates storage
│   └── docker-compose.yml    # Reverse proxy deployment
├── jenkins/                   # CI/CD pipeline infrastructure
│   ├── jenkins-agent/        # Custom Jenkins agent configuration
│   └── docker-compose.yml    # Jenkins master and agent setup
├── pritunl/                  # VPN server setup
│   └── docker-compose.yml    # Pritunl VPN deployment
├── pihole/                   # DNS filtering and DHCP
│   ├── docker-compose.yml    # Pi-hole deployment
│   └── example.env           # Environment configuration template
├── jenkinsfiles/             # Jenkins pipeline definitions
│   └── Jenkinsfile.ansible-playbooks # Ansible automation pipeline
└── .github/                  # GitHub Actions workflows
    ├── workflows/            # CI/CD workflow definitions
    └── actions/              # Custom GitHub Actions
```

## Quick Start

### Prerequisites

- Docker and Docker Compose installed on target hosts
- Ansible installed on control machine
- SSH access to target infrastructure
- Domain name with DNS configured (stanodn.org in configurations)

### Initial Setup

1. **Clone the repository**:
   ```bash
   git clone <repository-url>
   cd stanodn-infra
   ```

2. **Configure Ansible inventory**:
   ```bash
   # Copy and configure static inventory
   cp ansible/inventory/static.yml.example ansible/inventory/static.yml
   # Edit with your actual host information
   
   # Configure Proxmox dynamic inventory (if using Proxmox)
   cp ansible/inventory/proxmox.yml.example ansible/inventory/proxmox.yml
   # Add your Proxmox credentials
   ```

3. **Set up user configuration**:
   ```bash
   cp ansible/group_vars/users.yml.example ansible/group_vars/users.yml
   # Configure users, SSH keys, and access
   
   cp ansible/group_vars/cloud-init.yml.example ansible/group_vars/cloud-init.yml
   # Configure cloud-init settings for VMs
   ```

4. **Create Ansible vault password file**:
   ```bash
   echo "your-vault-password" > ansible/.vault_pass
   chmod 600 ansible/.vault_pass
   ```

## Component Details

### Reverse Proxy (Nginx)

The reverse proxy provides:
- **SSL Termination**: Automatic HTTPS with security headers
- **Load Balancing**: Route traffic to backend services
- **Security**: Headers and access control
- **Monitoring**: Access and error logging

**Configured Services**:
- `jenkins.stanodn.org` → Jenkins CI/CD (port 8080)
- `grafana.stanodn.org` → Grafana dashboards (port 3000)
- `prometheus.stanodn.org` → Prometheus metrics (port 9090)
- `pritunl.stanodn.org` → VPN management (port 8443)
- `pihole.stanodn.org` → DNS admin interface
- `proxmox.stanodn.org` → Proxmox VE web interface
- `disk.stanodn.org` → File management interface

**Deployment**:
```bash
cd reverse-proxy
docker-compose up -d
```

### VPN Access (Pritunl)

Pritunl provides secure VPN access with:
- **Multi-protocol Support**: OpenVPN and WireGuard
- **User Management**: Web-based user and organization management
- **High Availability**: MongoDB backend support
- **Network Routing**: Site-to-site and client-to-site VPN

**Configuration**:
- Web interface: `https://pritunl.stanodn.org`
- VPN ports: 1194 (OpenVPN), 15343 (management)
- Admin interface: Port 8443 (HTTPS), 8080 (HTTP)

**Deployment**:
```bash
cd pritunl
docker-compose up -d
```

### DNS & Ad-Blocking (Pi-hole)

Pi-hole provides network-wide ad blocking and DNS management:
- **DNS Filtering**: Block ads, trackers, and malicious domains
- **DHCP Server**: Network IP address management
- **Query Logging**: DNS request monitoring and statistics
- **Custom DNS**: Local domain resolution

**Features**:
- Web admin interface for management
- Custom DNS upstream servers (8.8.8.8, 8.8.4.4)
- DHCP range: 192.168.0.1-250
- Network interface monitoring

**Configuration**:
```bash
cd pihole
cp example.env .env
# Edit .env with your network settings
docker-compose up -d
```

### CI/CD Pipeline (Jenkins)

Jenkins provides automated deployment and infrastructure management:
- **Master-Agent Architecture**: Scalable build execution
- **Ansible Integration**: Infrastructure automation workflows
- **Pipeline Management**: Parameterized job execution
- **Security**: Credential management and access control

**Key Features**:
- **Ansible Playbook Pipeline**: Dynamic inventory discovery and execution
- **Docker Integration**: Containerized build environments
- **Parameterized Builds**: Flexible job configuration
- **Monitoring**: Build history and status tracking

**Jenkinsfile Pipeline Features**:
- Dynamic playbook and inventory discovery
- Dry-run (check mode) support
- Tag-based task execution
- Custom variable injection
- Vault-encrypted credential management

**Deployment**:
```bash
cd jenkins
docker-compose up -d
```

### Monitoring Stack (Prometheus + Grafana)

Comprehensive observability solution:

**Prometheus**:
- **Metrics Collection**: Time-series data storage
- **Service Discovery**: Automatic target discovery
- **Alerting**: Rule-based alerting system
- **Data Retention**: 7-day default retention

**Grafana**:
- **Visualization**: Interactive dashboards
- **Data Sources**: Prometheus integration
- **User Management**: Role-based access
- **Alerting**: Visual alert management

**Configuration**:
- Prometheus: `http://prometheus:9090`
- Grafana: `http://grafana:3000` (admin/admin default)
- Custom exporter: `monad-exporter:8000`

**Deployment**:
```bash
cd docker-apps/monitoring
docker-compose up -d
```

### Infrastructure Automation (Ansible)

Comprehensive automation framework:

**Playbooks**:
- **`ping.yml`**: Infrastructure connectivity testing
- **`docker-install.yml`**: Automated Docker installation
- **`docker-deploy.yml`**: Application deployment automation
- **`server_sync_users.yml`**: User account synchronization

**Roles**:
- **`configure/docker`**: Docker and Docker Compose installation
- **`configure/docker-deploy`**: Application deployment with backup
- **`configure/sync_users`**: User management and SSH key deployment
- **`provision/proxmox`**: VM and template provisioning

**Inventory Management**:
- **Static Inventory**: Physical hosts and infrastructure
- **Dynamic Inventory**: Proxmox VE integration
- **Group Variables**: Environment-specific configuration
- **Vault Integration**: Encrypted credential storage

### Proxmox Integration

Virtual machine and container management:
- **Dynamic Inventory**: Automatic VM discovery
- **Template Management**: Standardized VM deployment
- **Cloud-init Integration**: Automated VM configuration
- **Cluster Management**: Multi-node Proxmox support

## Security Features

### SSL/TLS Configuration
- **Modern Protocols**: TLS 1.2 and 1.3 only
- **Strong Ciphers**: High-security cipher suites
- **HSTS**: HTTP Strict Transport Security
- **OCSP Stapling**: Certificate validation optimization

### Security Headers
- **X-Frame-Options**: Clickjacking protection
- **X-XSS-Protection**: Cross-site scripting prevention
- **Content-Security-Policy**: Resource loading restrictions
- **X-Content-Type-Options**: MIME type validation

### Access Control
- **SSH Key Authentication**: Public key-based access
- **Ansible Vault**: Encrypted credential storage
- **Network Segmentation**: Service isolation
- **Firewall Integration**: Traffic filtering support

## Automation Workflows

### GitHub Actions

Automated deployment workflows for each service:
- **deploy-jenkins.yml**: Jenkins service deployment
- **deploy-pihole.yml**: Pi-hole service deployment
- **deploy-pritunl.yml**: Pritunl VPN deployment
- **deploy-reverse-proxy.yml**: Nginx reverse proxy deployment

**Workflow Features**:
- **Path-based Triggers**: Deploy only when service files change
- **SSH Deployment**: Secure remote deployment
- **Service Validation**: Health check verification
- **Manual Triggers**: On-demand deployment support

### Jenkins Pipelines

**Ansible Playbook Pipeline**:
- **Dynamic Discovery**: Automatic playbook and inventory detection
- **Parameter Validation**: Input validation and syntax checking
- **Execution Options**: Dry-run, tagging, and custom variables
- **Result Reporting**: Detailed execution summaries

**Pipeline Parameters**:
- **PLAYBOOK**: Dynamic list of available playbooks
- **TARGET**: Auto-discovered hosts and groups
- **DRY_RUN**: Check mode execution
- **TAGS**: Selective task execution
- **EXTRA_VARS**: Custom variable injection

## Monitoring and Observability

### Metrics Collection
- **System Metrics**: CPU, memory, disk, network
- **Application Metrics**: Service-specific monitoring
- **Custom Exporters**: Specialized metric collection
- **Alert Rules**: Automated issue detection

### Dashboard Features
- **Real-time Monitoring**: Live system status
- **Historical Analysis**: Trend identification
- **Alert Visualization**: Issue tracking and resolution
- **Custom Queries**: Flexible data exploration

### Log Management
- **Centralized Logging**: Service log aggregation
- **Access Logs**: HTTP request monitoring
- **Error Tracking**: Issue identification and debugging
- **Retention Policies**: Storage optimization

## Maintenance and Operations

### Backup Procedures
- **Configuration Backup**: Automated config preservation
- **Data Backup**: Service data protection
- **Rollback Support**: Quick service restoration
- **Version Control**: Change tracking and history

### Update Management
- **Service Updates**: Container image updates
- **Security Patches**: OS and application patching
- **Configuration Changes**: Version-controlled updates
- **Rollback Procedures**: Safe change reversal

### Health Monitoring
- **Service Health**: Container status monitoring
- **Resource Usage**: Performance tracking
- **Connectivity Tests**: Network validation
- **Automated Alerts**: Issue notification

## Troubleshooting

### Common Issues

**Ansible Connection Problems**:
```bash
# Test connectivity
ansible all -m ping

# Check inventory
ansible-inventory --list

# Validate playbook syntax
ansible-playbook --syntax-check playbooks/general/ping.yml
```

**Docker Service Issues**:
```bash
# Check service status
docker-compose ps

# View service logs
docker-compose logs -f [service-name]

# Restart services
docker-compose restart [service-name]
```

**SSL Certificate Issues**:
```bash
# Check certificate validity
openssl x509 -in ssl/origin.pem -text -noout

# Verify certificate chain
openssl verify -CAfile ca-bundle.pem ssl/origin.pem
```

### Log Locations
- **Nginx Logs**: `/var/log/nginx/`
- **Container Logs**: `docker-compose logs`
- **Ansible Logs**: Jenkins build console
- **System Logs**: `/var/log/syslog`

## Deployment Strategies

### Development Environment
1. Use local containers for testing
2. Validate playbooks with check mode
3. Test SSL with self-signed certificates
4. Monitor resource usage

### Production Deployment
1. Deploy monitoring stack first
2. Configure SSL certificates
3. Deploy core services (DNS, VPN)
4. Configure reverse proxy
5. Deploy CI/CD pipeline
6. Set up automated backups

### Scaling Considerations
- **Horizontal Scaling**: Add more nodes to cluster
- **Load Distribution**: Balance services across hosts
- **Resource Allocation**: Monitor and adjust limits
- **Network Optimization**: Optimize traffic routing

## Additional Resources

### Documentation
- [Ansible Documentation](https://docs.ansible.com/)
- [Docker Compose Reference](https://docs.docker.com/compose/)
- [Nginx Configuration Guide](https://nginx.org/en/docs/)
- [Prometheus Configuration](https://prometheus.io/docs/)

### Community
- Submit issues and feature requests
- Contribute improvements and fixes
- Share deployment experiences
- Participate in discussions

## License

This project is licensed under the MIT License - see the LICENSE file for details.

## Contributing

1. Fork the repository
2. Create a feature branch
3. Make changes and test thoroughly
4. Submit a pull request with detailed description

---

**Maintained by**: Stanislav Odnorog
**Last Updated**: 06.2024  