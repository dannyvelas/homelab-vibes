terraform {
  required_version = ">= 1.5.0"

  required_providers {
    libvirt = {
      source  = "dmacvicar/libvirt"
      version = "~> 0.8"
    }
  }
}

# Libvirt provider — connects to the KVM/libvirt API on each host.
# The URI is dynamically configured per host via terraform.tfvars.
provider "libvirt" {
  uri = var.libvirt_uri
}

# Libvirt VM lifecycle management.
# Terraform creates/destroys VMs via the dmacvicar/libvirt provider.
# OS configuration inside the VM is handled by Ansible (vm-guest role).

# Base OS image volume (shared, read-only backing store)
resource "libvirt_volume" "debian_base" {
  name   = "debian-12-base.qcow2"
  pool   = "default"
  source = "https://cloud.debian.org/images/cloud/bookworm/latest/debian-12-generic-amd64.qcow2"
  format = "qcow2"
}

# Per-host VM disk (copy-on-write from base image)
resource "libvirt_volume" "vm_disk" {
  for_each = var.hosts

  name           = "${each.key}-vm.qcow2"
  pool           = "default"
  base_volume_id = libvirt_volume.debian_base.id
  size           = var.vm_disk_gb * 1024 * 1024 * 1024 # Convert GB to bytes
  format         = "qcow2"
}

# Cloud-init config for each VM
resource "libvirt_cloudinit_disk" "vm_cloudinit" {
  for_each = var.hosts

  name = "${each.key}-vm-cloudinit.iso"
  pool = "default"

  user_data = templatefile("${path.module}/templates/cloud-init-user-data.yml.tftpl", {
    hostname = "${each.key}-vm"
    ssh_user = var.ssh_user
  })

  meta_data = templatefile("${path.module}/templates/cloud-init-meta-data.yml.tftpl", {
    instance_id = "${each.key}-vm"
    hostname    = "${each.key}-vm"
  })
}

# NAT network for each host's VMs
resource "libvirt_network" "vm_nat" {
  for_each = var.hosts

  name      = "${each.key}-nat"
  mode      = "nat"
  domain    = "${each.key}.local"
  autostart = true

  addresses = [each.value.nat_subnet]

  dhcp {
    enabled = true
  }

  dns {
    enabled = true
  }
}

# The workload VM (one per host)
resource "libvirt_domain" "workload_vm" {
  for_each = var.hosts

  name   = "${each.key}-vm"
  memory = var.vm_memory_mb
  vcpu   = var.vm_vcpus

  cloudinit = libvirt_cloudinit_disk.vm_cloudinit[each.key].id

  disk {
    volume_id = libvirt_volume.vm_disk[each.key].id
  }

  network_interface {
    network_id     = libvirt_network.vm_nat[each.key].id
    wait_for_lease = true
  }

  console {
    type        = "pty"
    target_type = "serial"
    target_port = "0"
  }

  graphics {
    type        = "vnc"
    listen_type = "none"
  }

  autostart = true
}

# Output VM IPs for use by Ansible and other tools
output "vm_ips" {
  description = "Map of host name to VM IP address"
  value = {
    for host, domain in libvirt_domain.workload_vm :
    host => domain.network_interface[0].addresses[0]
  }
}
