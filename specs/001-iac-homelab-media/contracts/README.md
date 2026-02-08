# IaC System Contracts

**Feature Branch**: `001-iac-homelab-media`
**Created**: 2026-02-07
**Status**: Phase 1 Design
**Input**: Functional Requirements from `/specs/001-iac-homelab-media/spec.md`, Research findings from `/specs/001-iac-homelab-media/research.md`

This document defines the CLI contracts for interacting with the IaC system. The primary interface is command-line driven, wrapping Terraform, Ansible, and Nomad operations behind simplified commands. Custom Go scripts may provide the wrapper layer for enhanced UX and error handling.

## IaC Operations

### 1. Server Provisioning and Host Setup

**Purpose**: Automate initial setup of a bare-metal server: OS hardening, hypervisor installation, Nomad server/client, Tailscale, NAT networking, and VM creation.

-   **Functional Requirements**: FR-001, FR-007, FR-008
-   **CLI Command**: `iac provision host --name <host_name> --ip <ip_address>`
-   **Inputs**:
    -   `host_name`: String - Unique identifier for the host (must match Ansible inventory).
    -   `ip_address`: String (e.g., `192.168.1.10`) - Host IP on the home LAN.
    -   `ssh_user`: String (default: `admin`) - Initial SSH user.
    -   `ssh_key_path`: String (default: `~/.ssh/id_ed25519`) - Path to SSH private key.
-   **What it does** (in order):
    1.  Runs Terraform to register the host in state.
    2.  Runs Ansible playbooks: OS hardening (UFW, SSH, auto-updates, least privilege users).
    3.  Installs KVM/libvirt hypervisor.
    4.  Installs and configures Nomad (server + client mode).
    5.  Configures NAT networking (Linux bridge, private subnet, iptables rules).
    6.  Creates the workload VM (Debian guest, Nomad client, Docker).
    7.  Applies VM OS hardening.
-   **Outputs**:
    -   Confirmation of successful provisioning.
    -   Host and VM IPs, Nomad cluster status.

### 2. VPN Deployment

**Purpose**: Deploy and configure Tailscale on the host for secure remote access.

-   **Functional Requirements**: FR-002, FR-003
-   **CLI Command**: `iac deploy vpn --host <host_name> --auth-key <tailscale_auth_key>`
-   **Inputs**:
    -   `host_name`: String - Target host for Tailscale deployment.
    -   `tailscale_auth_key`: String (Sensitive) - Pre-authenticated key for node registration.
    -   `advertised_routes`: List of Strings (optional, default: `["192.168.1.0/24"]`) - Subnets to advertise.
-   **What it does**:
    1.  Runs Ansible playbook to install Tailscale on the host.
    2.  Registers the node with the Tailnet using the auth key.
    3.  Configures subnet routing for the specified routes.
-   **Outputs**:
    -   Tailscale IP address of the host.
    -   Confirmation of subnet router advertisement.

### 3. Media Application Deployment

**Purpose**: Deploy a media application as a Nomad-scheduled Docker container inside the workload VM.

-   **Functional Requirements**: FR-004, FR-004a, FR-005, FR-006
-   **CLI Command**: `iac deploy app --name <app_name>`
-   **Inputs**:
    -   `app_name`: String - Application to deploy (`plex`, `sonarr`, `radarr`, `bazarr`).
    -   `media_path`: String (optional) - Host path for media storage passthrough.
-   **What it does**:
    1.  Submits the corresponding Nomad job file (`iac/nomad/jobs/<app_name>.hcl`) to the Nomad cluster.
    2.  Nomad schedules the container on an available VM's Docker daemon.
    3.  Container runs with `--read-only` root filesystem; `/config`, `/downloads`, `/media` mounted as writable volumes.
    4.  Updates reverse proxy configuration to route traffic to the new container.
-   **Outputs**:
    -   Nomad job status (running/pending).
    -   Access URL via reverse proxy (e.g., `https://<host_ip>/plex`).

### 4. Reverse Proxy Deployment

**Purpose**: Deploy the reverse proxy (Traefik/Caddy) inside the workload VM as the single ingress point.

-   **Functional Requirements**: FR-007, FR-009
-   **CLI Command**: `iac deploy proxy`
-   **What it does**:
    1.  Submits the reverse proxy Nomad job file.
    2.  Configures TLS termination, rate limiting.
    3.  Host iptables forwards port 443 to the VM's reverse proxy.
-   **Outputs**:
    -   Confirmation of proxy deployment.
    -   Listening port and TLS status.

### 5. Cluster Status

**Purpose**: Display the current state of all hosts, VMs, Nomad jobs, and Tailscale nodes.

-   **CLI Command**: `iac status`
-   **Outputs**:
    -   Table of hosts with status, IP, Tailscale IP.
    -   Table of VMs with status, NAT IP, allocated resources.
    -   Table of Nomad jobs with status, placement (which VM), health.
    -   Network security summary (iptables rules active, egress blocking status).

## Nomad Job File Schema

Each media application has an HCL job file in `iac/nomad/jobs/`. Common structure:

```hcl
job "<app_name>" {
  datacenters = ["dc1"]
  type        = "service"

  group "<app_name>" {
    count = 1

    network {
      port "<app_name>" {
        static = <port>
        to     = <container_port>
      }
    }

    task "<app_name>" {
      driver = "docker"

      config {
        image    = "<docker_image>"
        readonly_rootfs = true
        ports    = ["<app_name>"]

        volumes = [
          "<host_config_path>:/config",
          "<host_downloads_path>:/downloads",
          "<host_media_path>:/media",
        ]

        # Writable tmpfs for apps that need /tmp
        tmpfs = ["/tmp"]
      }

      resources {
        cpu    = <cpu_mhz>
        memory = <memory_mb>
      }
    }
  }
}
```

## Ansible Inventory Structure

```yaml
# iac/ansible/inventory/hosts.yml
all:
  children:
    hosts:
      hosts:
        homelab-host-01:
          ansible_host: 192.168.1.10
          nat_subnet: 192.168.122.0/24
        homelab-host-02:
          ansible_host: 192.168.1.11
          nat_subnet: 192.168.123.0/24
    vms:
      hosts:
        homelab-vm-media-01:
          ansible_host: 192.168.122.50
          parent_host: homelab-host-01
        homelab-vm-media-02:
          ansible_host: 192.168.123.50
          parent_host: homelab-host-02
```

## Interaction Model

The primary interaction model is a CLI-driven workflow. An engineer executes `iac` commands that internally orchestrate Terraform, Ansible, and Nomad. The `iac` wrapper (potentially a Go binary) provides:
-   Simplified, single-command operations
-   Input validation and error handling
-   Progress reporting
-   Idempotent execution (safe to re-run)
