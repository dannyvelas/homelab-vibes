# Quickstart: IaC Homelab Media

**Feature Branch**: `001-iac-homelab-media`
**Created**: 2026-02-07

## Prerequisites

### Hardware
-   Two Debian servers (laptops or bare-metal), each with:
    -   8GB+ RAM
    -   Dual-core CPU with virtualization support (Intel VT-x)
    -   128GB+ storage
    -   Ethernet connection to home LAN

### Software (on your workstation)
-   SSH client with key-based authentication configured to both servers
-   Terraform (latest stable)
-   Ansible (latest stable)
-   Nomad CLI (latest stable)
-   WireGuard client (available on all major platforms)
-   Git

### Accounts & Keys
-   SSH key pair deployed to both servers

### Network Preparation
-   **One-time manual step**: Forward UDP port 51820 on your home gateway/router to the IP of the server that will act as the WireGuard endpoint. This requires access to your router's admin interface.
-   (Optional) If your home IP is dynamic, set up a free dynamic DNS hostname (e.g., via DuckDNS, No-IP) pointing to your home public IP.

### Server Preparation
-   Debian installed (minimal/netinst) on both servers
-   SSH access working from your workstation to both servers
-   Intel VT-x enabled in BIOS (required for KVM)

## Setup (One-Time)

### 1. Clone the repository

```bash
git clone <repo-url> && cd homelab-vibe
```

### 2. Configure everything (one file)

Edit `homelab.yml` — this is the **only file you need to configure**:

```yaml
# homelab.yml — single source of truth
cluster:
  name: homelab
  datacenter: dc1

hosts:
  homelab-host-01:
    ip: <server-1-ip>
    nat_subnet: 192.168.122.0/24
    wireguard_endpoint: true
  homelab-host-02:
    ip: <server-2-ip>
    nat_subnet: 192.168.123.0/24

vpn:
  subnet: 10.0.0.0/24
  endpoint: <your-public-ip-or-ddns-hostname>
  port: 51820
  lan_routes:
    - 192.168.1.0/24

storage:
  media_path: /mnt/media
  downloads_path: /mnt/downloads
  config_path: /mnt/config

apps:
  plex:
    enabled: true
    port: 32400
  sonarr:
    enabled: true
    port: 8989
  radarr:
    enabled: true
    port: 7878
  bazarr:
    enabled: true
    port: 6767

auto_update:
  enabled: true
  schedule: "0 3 * * *"  # daily at 3 AM

secrets:
  ssh_key_path: ~/.ssh/id_ed25519
  ssh_user: admin
```

That's it. The `iac` CLI reads this file and generates all tool-specific configs (Ansible inventory, Terraform variables, Nomad job files) automatically. You never edit those generated files directly.

## Deployment

### Step 1: Provision hosts

Sets up both servers: OS hardening, hypervisor, Nomad, NAT networking, and workload VMs.

```bash
iac provision host --name homelab-host-01 --ip <server-1-ip>
iac provision host --name homelab-host-02 --ip <server-2-ip>
```

**What happens**: Ansible hardens the OS (firewall, SSH, auto-updates), installs KVM/libvirt, installs Nomad (server + client), creates a workload VM with Docker and a Nomad client, and configures NAT networking with iptables egress blocking.

### Step 2: Deploy VPN

Installs WireGuard on the designated host for secure remote access.

```bash
iac deploy vpn --host homelab-host-01
```

**What happens**: WireGuard kernel module is enabled, server keys are generated, the wg0 interface is configured with the home LAN subnet as an allowed route, and iptables rules are set for forwarding VPN traffic to the LAN. The VPN runs on one host only (the one with the port forward on the gateway). Remote team members connect using WireGuard client configs.

### Step 2a: Generate VPN client configs

```bash
iac generate vpn-client --name "danny-laptop"
iac generate vpn-client --name "danny-phone"
```

**What happens**: Generates a WireGuard client config file (and QR code for mobile) with the server's public key, endpoint address, and the client's unique key pair. The client imports this config into the WireGuard app to connect. Each team member gets their own config with a unique key pair.

### Step 3: Deploy reverse proxy

Deploys Traefik/Caddy inside the workload VMs as the single ingress point.

```bash
iac deploy proxy
```

**What happens**: Nomad schedules the reverse proxy container inside each VM. Host iptables forwards port 443 to the VM's proxy.

### Step 4: Deploy media applications

```bash
iac deploy app --name plex
iac deploy app --name sonarr
iac deploy app --name radarr
iac deploy app --name bazarr
```

**What happens**: Each command submits a Nomad job. Nomad schedules the container on an available VM's Docker daemon with a read-only root filesystem. The reverse proxy automatically routes traffic to the new container.

**Auto-updates**: If `auto_update.enabled` is `true` in `homelab.yml` (the default), a periodic Nomad job is also deployed that checks for new Docker image versions daily. When a new version is found, Nomad performs a health-checked rolling update with automatic rollback if the new version fails. No manual intervention needed.

### Step 5: Verify

```bash
iac status
```

Check that all hosts, VMs, and Nomad jobs are healthy. Access:
-   Plex: `https://<host-ip>/plex` (or `http://<host-ip>:32400` if direct)
-   Sonarr: `https://<host-ip>/sonarr`
-   Radarr: `https://<host-ip>/radarr`
-   Bazarr: `https://<host-ip>/bazarr`

## Architecture Overview

```
Home LAN (192.168.1.0/24)
  │
  ├── Home Gateway/Router
  │     └── UDP 51820 forwarded → Host 01
  │
  ├── Host 01 (192.168.1.10)
  │     ├── Nomad server + client
  │     ├── WireGuard (wg0 interface, VPN endpoint)
  │     ├── iptables (NAT, port forwarding, egress blocking)
  │     └── VM (192.168.122.50) ← private subnet, not on LAN
  │           ├── Nomad client
  │           ├── Docker
  │           ├── Traefik (reverse proxy)
  │           ├── Plex container (read-only root)
  │           └── Sonarr container (read-only root)
  │
  └── Host 02 (192.168.1.11)
        ├── Nomad server + client
        ├── iptables (NAT, port forwarding, egress blocking)
        └── VM (192.168.123.50) ← private subnet, not on LAN
              ├── Nomad client
              ├── Docker
              ├── Traefik (reverse proxy)
              ├── Radarr container (read-only root)
              └── Bazarr container (read-only root)
```

## Security Layers

1. **OS hardening**: UFW, SSH key-only, auto-updates, least privilege users
2. **NAT networking**: VMs on private subnet, invisible to home LAN
3. **Egress blocking**: VMs cannot reach other LAN devices
4. **VM kernel isolation**: Container escapes land in VM kernel, not host
5. **Reverse proxy**: Single ingress point, TLS termination, rate limiting
6. **Read-only containers**: Immutable root filesystem, writable only for data volumes

## Scaling (Future)

When you migrate to beefier hardware:
1. Add new hosts to Ansible inventory
2. Run `iac provision host` for each new host
3. Nomad automatically distributes workloads across the expanded cluster
4. No changes to Nomad job files needed
5. Run `iac deploy vpn --host <new-host>` if you want to move the WireGuard endpoint (update gateway port forward accordingly)
