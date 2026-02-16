# homelab-vibe Development Guidelines

Auto-generated from all feature plans. Last updated: 2026-02-09

## Active Technologies
- Go (for `iac` CLI) + Ansible (with `community.docker` collection), Terraform (with `dmacvicar/libvirt` provider), WireGuard, KVM/libvirt, Docker (002-simplify-iac-media)
- Host-level storage with passthrough to VM and containers (media libraries, config, downloads) (001-iac-homelab-media)

## Project Structure

```text
iac/
  cli/         # Go CLI source
  ansible/     # Ansible playbooks and roles
  terraform/   # Terraform VM lifecycle
tests/
```

## Commands

```bash
cd iac/cli && go build ./...   # Build CLI
go vet ./...                    # Lint
```

## Code Style

Go: Follow standard conventions (gofmt, go vet)

## Recent Changes
- 002-simplify-iac-media: Removed Nomad, replaced with Ansible docker_container module
- 001-iac-homelab-media: Initial IaC implementation with Go CLI, Ansible, Terraform, KVM/libvirt, Docker

<!-- MANUAL ADDITIONS START -->
<!-- MANUAL ADDITIONS END -->
