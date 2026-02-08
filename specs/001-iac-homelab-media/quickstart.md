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
-   Git

### Accounts & Keys
-   Tailscale account with a pre-authenticated auth key
-   SSH key pair deployed to both servers

### Server Preparation
-   Debian installed (minimal/netinst) on both servers
-   SSH access working from your workstation to both servers
-   Intel VT-x enabled in BIOS (required for KVM)

## Setup (One-Time)

### 1. Clone the repository

```bash
git clone <repo-url> && cd homelab-vibe
```

### 2. Configure inventory

Edit `iac/ansible/inventory/hosts.yml` with your server details:

```yaml
all:
  children:
    hosts:
      hosts:
        homelab-host-01:
          ansible_host: <server-1-ip>
          nat_subnet: 192.168.122.0/24
        homelab-host-02:
          ansible_host: <server-2-ip>
          nat_subnet: 192.168.123.0/24
```

### 3. Configure secrets

Copy the example secrets file and fill in your values:

```bash
cp iac/ansible/inventory/group_vars/all/secrets.yml.example iac/ansible/inventory/group_vars/all/secrets.yml
# Edit with your Tailscale auth key, SSH key paths, etc.
```

## Deployment

### Step 1: Provision hosts

Sets up both servers: OS hardening, hypervisor, Nomad, NAT networking, and workload VMs.

```bash
iac provision host --name homelab-host-01 --ip <server-1-ip>
iac provision host --name homelab-host-02 --ip <server-2-ip>
```

**What happens**: Ansible hardens the OS (firewall, SSH, auto-updates), installs KVM/libvirt, installs Nomad (server + client), creates a workload VM with Docker and a Nomad client, and configures NAT networking with iptables egress blocking.

### Step 2: Deploy VPN

Installs Tailscale on both hosts for secure remote access.

```bash
iac deploy vpn --host homelab-host-01 --auth-key <your-tailscale-key>
iac deploy vpn --host homelab-host-02 --auth-key <your-tailscale-key>
```

**What happens**: Tailscale is installed on the host (not the VM), registers with your Tailnet, and advertises the home LAN subnet. Remote team members can now VPN in via Tailscale.

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
  ├── Host 01 (192.168.1.10)
  │     ├── Nomad server + client
  │     ├── Tailscale
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
        ├── Tailscale
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
