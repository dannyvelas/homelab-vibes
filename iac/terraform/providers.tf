terraform {
  required_version = ">= 1.5.0"

  required_providers {
    libvirt = {
      source  = "dmacvicar/libvirt"
      version = "~> 0.8"
    }
    nomad = {
      source  = "hashicorp/nomad"
      version = "~> 2.4"
    }
  }
}

# Libvirt provider — connects to the KVM/libvirt API on each host.
# The URI is dynamically configured per host via terraform.tfvars.
provider "libvirt" {
  uri = var.libvirt_uri
}

# Nomad provider — connects to the Nomad cluster API.
# Runs on the engineer's workstation; communicates over the network.
provider "nomad" {
  address = var.nomad_address
}
