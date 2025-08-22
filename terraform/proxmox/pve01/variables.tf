variable "proxmox_api_url" {
  description = "Proxmox API URL"
  type        = string
}

variable "proxmox_api_token_id" {
  description = "Proxmox API token ID"
  type        = string
}

variable "proxmox_api_token_secret" {
  description = "Proxmox API token secret"
  type        = string
  sensitive   = true
}

variable "proxmox_api_insecure" {
  description = "Skip TLS verification"
  type        = bool
  default     = true
}

variable "proxmox_node" {
  description = "Proxmox node name"
  type        = string
}

variable "vm_template_id" {
  description = "VM template ID to clone from"
  type        = number
}

variable "vm_storage" {
  description = "Storage location"
  type        = string
}

variable "vm_network_bridge" {
  description = "Network bridge"
  type        = string
}

variable "ssh_public_keys" {
  description = "SSH public key"
  type        = list
}

variable "default_username" {
  description = "Default username"
  type        = string
  default     = "ubuntu"
}

variable "dns_servers" {
  description = "DNS servers"
  type        = list(string)
  default     = ["8.8.8.8", "8.8.4.4"]
}

variable "virtual_machines" {
  description = "VM configurations"
  type = map(object({
    vm_id       = number
    name        = string
    description = string
    tags        = list(string)
    cpu = object({
      cores   = number
      sockets = number
    })
    memory = object({
      dedicated = number
    })
    disk = object({
      size = number
    })
    initialization = object({
      ip_config = object({
        ipv4 = object({
          address = string
          gateway = optional(string)
        })
      })
    })
  }))
}