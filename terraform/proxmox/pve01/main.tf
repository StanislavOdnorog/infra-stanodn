resource "proxmox_virtual_environment_vm" "ubuntu_vms" {
  for_each = var.virtual_machines

  node_name   = var.proxmox_node
  vm_id       = each.value.vm_id
  name        = each.value.name
  description = each.value.description
  tags        = each.value.tags
  on_boot     = true
  started     = true

  agent {
    enabled = true
  }

  cpu {
    cores   = each.value.cpu.cores
    sockets = each.value.cpu.sockets
    type    = "host"
  }

  memory {
    dedicated = each.value.memory.dedicated
  }

  network_device {
    bridge = var.vm_network_bridge
    model  = "virtio"
  }

  clone {
    vm_id = var.vm_template_id
  }

  disk {
    datastore_id = var.vm_storage
    interface    = "scsi0"
    iothread     = true
    discard      = "on"
    size         = each.value.disk.size
  }

  initialization {
    datastore_id = var.vm_storage

    dns {
      servers = var.dns_servers
    }

    ip_config {
      ipv4 {
        address = each.value.initialization.ip_config.ipv4.address
        gateway = each.value.initialization.ip_config.ipv4.gateway
      }
    }

    user_account {
      keys     = var.ssh_public_keys
      username = var.default_username
    }
  }
}