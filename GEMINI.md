# homelab-vibe Development Guidelines

Auto-generated from all feature plans. Last updated: 2026-02-07

## Active Technologies
- Go (for custom scripting/logic within IaC if needed, per user preference for Go, Nim, OCaml) + Ansible, Terraform, Tailscale, Nomad, KVM/libvirt, Docker (001-iac-homelab-media)
- Host-level storage with passthrough to VM and containers (media libraries, config, downloads) (001-iac-homelab-media)
- Go (for `iac` CLI) + Ansible (with `community.docker` collection), Terraform (with `dmacvicar/libvirt` provider), WireGuard, KVM/libvirt, Docker (002-simplify-iac-media)

- Go (for custom scripting/logic within IaC if needed, per user preference for Go, Nim, OCaml) + Ansible, Terraform, Tailscale (001-iac-homelab-media)

## Project Structure

```text
src/
tests/
```

## Commands

# Add commands for Go (for custom scripting/logic within IaC if needed, per user preference for Go, Nim, OCaml)

## Code Style

Go (for custom scripting/logic within IaC if needed, per user preference for Go, Nim, OCaml): Follow standard conventions

## Recent Changes
- 002-simplify-iac-media: Added Go (for `iac` CLI) + Ansible (with `community.docker` collection), Terraform (with `dmacvicar/libvirt` provider), WireGuard, KVM/libvirt, Docker
- 001-iac-homelab-media: Added Go (for custom scripting/logic within IaC if needed, per user preference for Go, Nim, OCaml) + Ansible, Terraform, Tailscale, Nomad, KVM/libvirt, Docker

- 001-iac-homelab-media: Added Go (for custom scripting/logic within IaC if needed, per user preference for Go, Nim, OCaml) + Ansible, Terraform, Tailscale

<!-- MANUAL ADDITIONS START -->
<!-- MANUAL ADDITIONS END -->
