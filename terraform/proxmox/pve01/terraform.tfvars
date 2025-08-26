# Proxmox Connection Configuration
# Copy this file to terraform.tfvars and update with your actual values

# Proxmox API Configuration
proxmox_api_url          = "https://192.168.1.50:8006"
proxmox_api_insecure     = true                    # Set to false if using valid SSL certificates
proxmox_node             = "pve01"                   # Your Proxmox node name

# VM Template and Storage Configuration
vm_storage          = "disk1tb-01"                  # Storage location for VM disks
vm_network_bridge   = "vmbr0"                     # Network bridge name

# SSH and User Configuration
ssh_public_keys      = [
  "ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABgQCzUXlDKjVBkd5g9HZmlElOJHaAhRSOACSe3N16JCMcj7w96opfwdrmq+wri9tUkyPQJeNknO83KtZyIeOwyTNcDAuDS27UWPHCv1qLCr2xk7iUXe+8hW7DisuVJ2/GOkdqxwjto3a8Npvc3gHJmrvKDLqiRaal6pxkVTwA/peMjLIKrEc6NATY4+Uu/EGNDbo9dToqmQyvD/iArz5+1kwG0VD4x1PXzBmTsv7PC6QDEciMN/TcXAKaN3oqtOm0wheYYInFy36yQEP4lgRwTUseMr1hSzpdpi3eN2BujB1zzDxWvN+NN3GdWVvdabbb2ORkC2F/XgTUu9UHffKDJreF7rJ66nKORcgMxuBBv9fOj1yXM1topsBFNfI+R+YDkSJozgfZCZHQUbzXiuzDyLtfA3fGcT4k1pja19OpodVrimoxsHAzoeQo2hyBDtPdOrXUEsZNeY4Mh2hkGYwrdXarsuhW2oUsrQrr2qByzlvQITf9PCdpACxcxpJEVNctUBk=",
  "ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAILx4I1Jn1KrZ7sbGRZkmtAY+1wG4kONRONL1pqWxG1BJ" 
]
default_username    = "ubuntu"
dns_servers         = ["192.168.1.55", "1.1.1.1"]

# Virtual Machines Configuration
# This section defines all your VMs and their specifications
virtual_machines = {

  # GitLab/Jenkins Runner VM
  "stan-runner01" = {
    vm_id       = 1001
    name        = "stan-runner01"
    description = "GitLab/Jenkins CI/CD Runner - Automated build and deployment tasks"
    tags        = ["terraform", "ubuntu", "runner", "gitlab", "jenkins-runner", "cicd", "stan"]
    
    cpu = {
      cores   = 2
      sockets = 1
    }
    
    memory = {
      dedicated = 6144
    }
    
    disk = {
      size = 120
    }
    
    initialization = {
      ip_config = {
        ipv4 = {
          address = "dhcp"
        }
      }
    }
  }
  
  # Network NameServer + DHCP VM
  "stan-ns01" = {
    vm_id       = 1002
    name        = "stan-ns01"
    description = "Network NS -  DNS, DHCP services"
    tags        = ["terraform", "ubuntu", "dns", "dhcp", "networking", "stan"]
    
    cpu = {
      cores   = 1
      sockets = 1
    }
    
    memory = {
      dedicated = 4096
    }
    
    disk = {
      size = 100
    }
    
    initialization = {
              ip_config = {
          ipv4 = {
            address = "192.168.1.55/24"
            gateway = "192.168.1.1"
          }
        }
    }
  }

  # Network Gateway VM
  "stan-gw01" = {
    vm_id       = 1005
    name        = "stan-gw01"
    description = "Network Gateway - NGINX reverse proxy"
    tags        = ["terraform", "ubuntu", "gateway", "nginx", "stan"]
    
    cpu = {
      cores   = 1
      sockets = 1
    }
    
    memory = {
      dedicated = 4096
    }
    
    disk = {
      size = 100
    }
    
    initialization = {
              ip_config = {
          ipv4 = {
            address = "192.168.1.60/24"
            gateway = "192.168.1.1"
          }
        }
    }
  }
  
  # Automation Control Node
  "stan-acn01" = {
    vm_id       = 1003
    name        = "stan-acn01"
    description = "Automation Control Node - Jenkins master, Ansible control plane"
    tags        = ["terraform", "ubuntu", "automation", "jenkins", "ansible", "control-node", "stan"]
    
    cpu = {
      cores   = 4
      sockets = 1
    }
    
    memory = {
      dedicated = 16384
    }
    
    disk = {
      size = 200
    }
    
    initialization = {
      ip_config = {
        ipv4 = {
          address = "dhcp"
        }
      }
    }
  }

  # N8N
  "stan-n8n01" = {
    vm_id       = 1004
    name        = "stan-n8n01"
    description = "N8N Automation Node - Self-hosted workflow automation tool"
    tags        = ["terraform", "ubuntu", "automation", "n8n", "workflow", "stan"]
    
    cpu = {
      cores   = 4
      sockets = 1
    }
    
    memory = {
      dedicated = 16384
    }
    
    disk = {
      size = 200
    }
    
    initialization = {
      ip_config = {
        ipv4 = {
          address = "dhcp"
        }
      }
    }
  }
}