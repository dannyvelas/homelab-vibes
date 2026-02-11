# IaC System Contracts

**Feature Branch**: `002-simplify-iac-media`
**Created**: 2026-02-09
**Status**: Phase 1 Design
**Input**: Functional Requirements from `/specs/002-simplify-iac-media/spec.md`, Research findings from `/specs/002-simplify-iac-media/research.md`

This document defines the CLI contracts for interacting with the IaC system. The primary interface is command-line driven, wrapping Terraform and Ansible operations behind simplified commands. The `iac` CLI (Go binary) reads all configuration from a single `homelab.yml` file and generates tool-specific configs before invoking underlying tools. Engineers never edit Ansible inventory or Terraform tfvars files directly.

Compared to the 001 contracts, all Nomad references are removed. Container management is handled by Ansible's `community.docker.docker_container` module. Auto-updates use a systemd timer + shell script inside each VM instead of a Nomad periodic job.

## IaC Operations

### 1. Server Provisioning and Host Setup

**Purpose**: Automate initial setup of a bare-metal server: OS hardening, hypervisor installation, WireGuard, NAT networking, and VM creation with Docker.

-   **Functional Requirements**: FR-001, FR-007, FR-008
-   **CLI Command**: `iac provision host --name <host_name> --ip <ip_address>`
-   **Inputs**:
    -   `host_name`: String - Unique identifier for the host (must match a key in `homelab.yml` hosts map).
    -   `ip_address`: String (e.g., `192.168.1.10`) - Host IP on the home LAN.
    -   `ssh_user`: String (default: `admin`) - Initial SSH user.
    -   `ssh_key_path`: String (default: `~/.ssh/id_ed25519`) - Path to SSH private key.
-   **What it does** (in order):
    1.  Generates Ansible inventory and Terraform tfvars from `homelab.yml`.
    2.  Runs Terraform to create the VM on the target host (per-host state in `.generated/terraform/<hostname>/`).
    3.  Runs Ansible playbooks: OS hardening (UFW, SSH, auto-updates, least privilege users).
    4.  Installs KVM/libvirt hypervisor.
    5.  Configures NAT networking (Linux bridge, private subnet, iptables rules).
    6.  Creates the workload VM (Debian guest, Docker engine).
    7.  Applies VM OS hardening.
-   **Outputs**:
    -   Confirmation of successful provisioning.
    -   Host and VM IPs.

### 2. VPN Deployment

**Purpose**: Deploy and configure WireGuard on the designated host for secure remote access.

-   **Functional Requirements**: FR-002, FR-003
-   **CLI Command**: `iac deploy vpn --host <host_name>`
-   **Inputs**:
    -   `host_name`: String - Target host for WireGuard deployment (must have gateway port forward configured).
    -   `vpn_subnet`: String (optional, default: `10.0.0.0/24`) - Subnet for VPN clients.
    -   `endpoint`: String (optional) - Public hostname or IP for remote clients. If omitted, read from config.
    -   `lan_routes`: List of Strings (optional, default: `["192.168.1.0/24"]`) - Home LAN subnets to route to VPN clients.
-   **What it does**:
    1.  Runs Ansible playbook to enable WireGuard kernel module on the host.
    2.  Generates server key pair (if not already present).
    3.  Configures wg0 interface with listen port (UDP 51820), server private key, and VPN subnet.
    4.  Sets up iptables rules for forwarding VPN client traffic to the home LAN and masquerading.
    5.  Enables and starts the wg-quick systemd service.
-   **Outputs**:
    -   Server public key (needed for client configs, auto-used by `generate vpn-client`).
    -   VPN endpoint address.
    -   Confirmation of wg0 interface status.

### 2a. VPN Client Config Generation

**Purpose**: Generate a WireGuard client configuration file for a team member's device.

-   **Functional Requirements**: FR-003
-   **CLI Command**: `iac generate vpn-client --name <peer_name>`
-   **Inputs**:
    -   `peer_name`: String - Human-readable name for the client (e.g., `danny-laptop`, `danny-phone`).
-   **What it does**:
    1.  Generates a unique key pair for the client.
    2.  Assigns the next available IP in the VPN subnet.
    3.  Adds the peer's public key to the server's WireGuard config.
    4.  Generates a client config file with server public key, endpoint, allowed IPs (home LAN routes), and the client's private key.
    5.  Generates a QR code (for mobile devices).
-   **Outputs**:
    -   Client config file path (e.g., `.generated/vpn-clients/danny-laptop.conf`).
    -   QR code (displayed in terminal for mobile scanning).
    -   Assigned VPN IP for the client.

### 3. Media Application Deployment

**Purpose**: Deploy a media application as an Ansible-managed Docker container inside the workload VM.

-   **Functional Requirements**: FR-004, FR-005, FR-006, FR-012, FR-014, FR-015
-   **CLI Command**: `iac deploy app --name <app_name>`
-   **Inputs**:
    -   `app_name`: String - Application to deploy (`plex`, `sonarr`, `radarr`, `bazarr`).
    -   All other configuration (image, port, storage paths) is read from `homelab.yml`.
-   **What it does**:
    1.  Reads app configuration from `homelab.yml`.
    2.  Runs Ansible playbook targeting the VM:
        a.  `community.docker.docker_image_pull` pulls the latest image.
        b.  `community.docker.docker_container` creates/recreates the container with:
            -   `read_only: true` root filesystem
            -   Writable volume mounts for `/config`, `/downloads`, `/media`
            -   `restart_policy: unless-stopped`
            -   Port mappings from `homelab.yml`
        c.  Templates a per-container run script to `/opt/homelab/run-<app>.sh` inside the VM.
    3.  If `auto_update.enabled` is `true`, deploys the systemd timer and update script (if not already present).
-   **Outputs**:
    -   Container status (running).
    -   Access URL via reverse proxy (e.g., `https://<host_ip>/plex`).
    -   Auto-update status (enabled/disabled).

### 4. Reverse Proxy Deployment

**Purpose**: Deploy the reverse proxy (Traefik) inside the workload VM as the single ingress point.

-   **Functional Requirements**: FR-007, FR-009
-   **CLI Command**: `iac deploy proxy`
-   **What it does**:
    1.  Runs Ansible playbook to deploy Traefik as a Docker container inside each VM.
    2.  Configures TLS termination, rate limiting.
    3.  Host iptables forwards port 443 to the VM's reverse proxy.
-   **Outputs**:
    -   Confirmation of proxy deployment.
    -   Listening port and TLS status.

### 5. Cluster Status

**Purpose**: Display the current state of all hosts, VMs, containers, and WireGuard peers.

-   **CLI Command**: `iac status`
-   **Outputs**:
    -   Table of hosts with status, IP, WireGuard VPN IP (if VPN endpoint).
    -   Table of VMs with status, NAT IP, allocated resources.
    -   Table of Docker containers with status, image digest, restart policy.
    -   Network security summary (iptables rules active, egress blocking status).

### 6. Auto-Update Management

**Purpose**: View auto-update status. Updates run autonomously via systemd timer.

-   **Functional Requirements**: FR-012
-   **CLI Commands**:
    -   `iac update status` -- Show auto-update status for all apps (last run time from journald, current image digest).
-   **What auto-updates do** (runs automatically per `auto_update.schedule` in `homelab.yml`):
    1.  A systemd timer triggers the update shell script inside each VM.
    2.  The script iterates over configured containers.
    3.  Pulls the latest Docker image for each container.
    4.  Compares the new image digest to the running container's image digest.
    5.  If a new image is detected, stops the old container and runs the per-container run script to recreate it with the new image.
    6.  Logs all actions to journald.
-   **Outputs** (from `iac update status`):
    -   Table of apps with: current image digest, last update time, auto-update enabled/disabled.

### 7. Generate Configs

**Purpose**: Regenerate all tool-specific configuration files from `homelab.yml`.

-   **Functional Requirements**: FR-013
-   **CLI Command**: `iac generate`
-   **What it does**:
    1.  Reads `homelab.yml`.
    2.  Generates Ansible inventory (`hosts.yml`) and group_vars into `.generated/ansible/`.
    3.  Generates per-host Terraform tfvars into `.generated/terraform/<hostname>/`.
-   **Outputs**:
    -   List of generated files.

### 8. Teardown

**Purpose**: Destroy all VMs and clean up infrastructure.

-   **CLI Command**: `iac teardown`
-   **What it does**:
    1.  Iterates over all hosts in `homelab.yml`.
    2.  Runs `terraform destroy` per-host using the host's tfvars and state file.
-   **Outputs**:
    -   Confirmation of teardown per host.

## Unified Configuration Schema

**Purpose**: Define the single source of truth for all infrastructure and application configuration (FR-013).

**File**: `homelab.yml` at repository root.

The `iac` CLI reads this file before every operation and generates tool-specific configs into `.generated/`:
-   `.generated/ansible/inventory/` -- Ansible hosts.yml and group_vars
-   `.generated/terraform/<hostname>/terraform.tfvars` -- Per-host Terraform variable values

**Schema**:

```yaml
cluster:
  name: String          # Cluster identifier (e.g., "homelab")

hosts:
  <host_name>:          # Unique host identifier
    ip: String          # Host IP on home LAN
    nat_subnet: String  # Private NAT subnet for VMs (e.g., "192.168.122.0/24")
    wireguard_endpoint: Boolean  # Whether this host runs the WireGuard server (only one host)

vpn:
  subnet: String        # VPN client subnet (e.g., "10.0.0.0/24")
  endpoint: String      # Public hostname or IP for remote clients
  port: Integer         # WireGuard listen port (default: 51820)
  lan_routes: [String]  # Home LAN subnets to route to VPN clients

storage:
  media_path: String    # Host path for media libraries
  downloads_path: String # Host path for download staging
  config_path: String   # Host path for application config/databases

apps:
  <app_name>:           # Application identifier (plex, sonarr, radarr, bazarr)
    enabled: Boolean    # Whether to deploy this app
    image: String       # Docker image (default: linuxserver/<app_name>:latest)
    port: Integer       # Application web UI port

auto_update:
  enabled: Boolean      # Enable automatic image updates (default: true)
  schedule: String      # Cron expression for update checks (default: "0 3 * * *")

secrets:
  ssh_key_path: String  # Path to SSH private key
  ssh_user: String      # SSH user for server access
```

**Changes from 001 schema**:
-   Removed `cluster.datacenter` (was a Nomad concept).
-   Removed `auto_update.auto_revert` and `auto_update.health_check_timeout` (rollback deferred to future feature).

**Rules**:
-   All values defined here are the canonical source. Tool-specific configs are generated, not hand-edited.
-   Sensitive values (WireGuard private keys) are NOT stored in `homelab.yml`. They are generated by the `iac` CLI at deploy time.
-   `.generated/` directory is gitignored -- it is regenerated on every `iac` command invocation.

## Ansible Inventory Structure

```yaml
# .generated/ansible/inventory/hosts.yml
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

The primary interaction model is a CLI-driven workflow. An engineer executes `iac` commands that internally orchestrate Terraform and Ansible. The `iac` wrapper (Go binary) provides:
-   Simplified, single-command operations
-   Input validation and error handling
-   Progress reporting
-   Idempotent execution (safe to re-run)
