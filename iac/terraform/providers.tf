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
