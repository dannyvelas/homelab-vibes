# Variables matching the homelab.yml schema.
# Values are generated into terraform.tfvars by the iac CLI.

# --- Provider connection ---

variable "libvirt_uri" {
  description = "Libvirt connection URI (e.g., qemu+ssh://admin@192.168.1.10/system)"
  type        = string
}

variable "nomad_address" {
  description = "Nomad cluster API address (e.g., http://192.168.1.10:4646)"
  type        = string
}

# --- Cluster ---

variable "cluster_name" {
  description = "Cluster identifier"
  type        = string
  default     = "homelab"
}

variable "datacenter" {
  description = "Nomad datacenter name"
  type        = string
  default     = "dc1"
}

# --- Hosts ---

variable "hosts" {
  description = "Map of host configurations from homelab.yml"
  type = map(object({
    ip                  = string
    nat_subnet          = string
    wireguard_endpoint  = optional(bool, false)
  }))
}

# --- VPN ---

variable "vpn_subnet" {
  description = "VPN client subnet CIDR"
  type        = string
  default     = "10.0.0.0/24"
}

variable "vpn_endpoint" {
  description = "Public hostname or IP for WireGuard endpoint"
  type        = string
}

variable "vpn_port" {
  description = "WireGuard listen port"
  type        = number
  default     = 51820
}

# --- Storage ---

variable "storage_media_path" {
  description = "Host path for media libraries"
  type        = string
}

variable "storage_downloads_path" {
  description = "Host path for download staging"
  type        = string
}

variable "storage_config_path" {
  description = "Host path for application config/databases"
  type        = string
}

# --- Apps ---

variable "apps" {
  description = "Map of application configurations from homelab.yml"
  type = map(object({
    enabled = bool
    image   = string
    port    = number
  }))
}

# --- VM ---

variable "vm_memory_mb" {
  description = "RAM allocated to each workload VM in MB"
  type        = number
  default     = 5120
}

variable "vm_vcpus" {
  description = "vCPUs allocated to each workload VM"
  type        = number
  default     = 2
}

variable "vm_disk_gb" {
  description = "Disk allocated to each workload VM in GB"
  type        = number
  default     = 40
}

# --- Secrets ---

variable "ssh_user" {
  description = "SSH user for server access"
  type        = string
  default     = "admin"
}

variable "ssh_key_path" {
  description = "Path to SSH private key"
  type        = string
  default     = "~/.ssh/id_ed25519"
}
