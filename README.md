# homelab-vibe

Infrastructure as Code for a homelab media environment. A single Go CLI (`iac`) reads one config file (`homelab.yml`) and orchestrates Terraform and Ansible to provision servers, deploy a WireGuard VPN, and run Plex, Sonarr, Radarr, and Bazarr as isolated Docker containers inside KVM virtual machines.

Designed to run on two 8GB RAM Debian 12 laptops today and migrate to proper bare-metal servers tomorrow — change the IPs in one file and re-run.

## Architecture

```
Home LAN (192.168.1.0/24)
  |
  +-- Home Gateway/Router
  |     \-- UDP 51820 forwarded -> Host 01
  |
  +-- Host 01 (192.168.1.10)
  |     +-- WireGuard (wg0, VPN endpoint)
  |     +-- iptables (NAT, port forwarding, egress blocking)
  |     \-- VM (192.168.122.50) <- private subnet, not on LAN
  |           +-- Docker
  |           +-- Traefik (reverse proxy)
  |           +-- Plex container (read-only root)
  |           \-- Sonarr container (read-only root)
  |
  \-- Host 02 (192.168.1.11)
        +-- iptables (NAT, port forwarding, egress blocking)
        \-- VM (192.168.123.50) <- private subnet, not on LAN
              +-- Docker
              +-- Traefik (reverse proxy)
              +-- Radarr container (read-only root)
              \-- Bazarr container (read-only root)
```

**Two-layer isolation**: Trusted infrastructure (WireGuard, KVM) runs on the host. Applications run inside VMs behind NAT. A container escape lands in the VM kernel, not the host.

### Security layers

1. **OS hardening** — UFW firewall, SSH key-only auth, unattended security updates
2. **NAT networking** — VMs on private subnets, invisible to the home LAN
3. **Egress blocking** — VMs cannot initiate connections to other LAN devices
4. **VM kernel isolation** — Container breakouts are contained by the VM boundary
5. **Reverse proxy** — Single ingress point with TLS, security headers, rate limiting
6. **Read-only containers** — Immutable root filesystem, writable only for data volumes

## Prerequisites

### Hardware

- Two Debian 12 servers (laptops or bare-metal), each with:
  - 8 GB+ RAM
  - Dual-core CPU with Intel VT-x
  - 128 GB+ storage
  - Ethernet to home LAN

### Software (on your workstation)

- [Go](https://go.dev/) 1.21+ (to build the CLI)
- [Terraform](https://www.terraform.io/) (latest stable)
- [Ansible](https://docs.ansible.com/) (latest stable)
- [WireGuard client](https://www.wireguard.com/install/) (for VPN access)
- SSH client with key-based auth configured to both servers

### Network preparation

Forward UDP port **51820** on your home gateway/router to the server that will act as the WireGuard endpoint. If your home IP is dynamic, set up a dynamic DNS hostname (e.g., DuckDNS, No-IP).

### Server preparation

- Debian 12 installed (minimal/netinst)
- SSH access working from your workstation
- Intel VT-x enabled in BIOS

## Setup

### 1. Clone and build

```bash
git clone <repo-url> && cd homelab-vibe
cd iac/cli && go build -o ../../iac . && cd ../..
```

### 2. Configure

Copy the example config and fill in your values:

```bash
cp homelab.yml.example homelab.yml
```

Edit `homelab.yml` — this is the **only file you need to touch**:

```yaml
cluster:
  name: homelab

hosts:
  homelab-host-01:
    ip: 192.168.1.10              # your server 1 LAN IP
    nat_subnet: 192.168.122.0/24
    wireguard_endpoint: true
  homelab-host-02:
    ip: 192.168.1.11              # your server 2 LAN IP
    nat_subnet: 192.168.123.0/24

vpn:
  subnet: 10.0.0.0/24
  endpoint: home.example.com      # your public IP or DDNS hostname
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
    image: linuxserver/plex:latest
    port: 32400
  sonarr:
    enabled: true
    image: linuxserver/sonarr:latest
    port: 8989
  radarr:
    enabled: true
    image: linuxserver/radarr:latest
    port: 7878
  bazarr:
    enabled: true
    image: linuxserver/bazarr:latest
    port: 6767

auto_update:
  enabled: true
  schedule: "0 3 * * *"           # daily at 3 AM UTC

secrets:
  ssh_key_path: ~/.ssh/id_ed25519
  ssh_user: admin
```

The `iac` CLI reads this file and generates all tool-specific configs (Ansible inventory, Terraform tfvars) into `.generated/`. You never edit those files directly.

## Usage

### Bootstrap infrastructure

Provision both servers — this hardens the OS, installs KVM/libvirt, creates workload VMs with Docker, and configures NAT networking:

```bash
iac provision host --name homelab-host-01
iac provision host --name homelab-host-02
```

### Deploy VPN

Install WireGuard on the designated host for secure remote access:

```bash
iac deploy vpn
```

Generate client configs for your team:

```bash
iac generate vpn-client --name "danny-laptop"
iac generate vpn-client --name "danny-phone"    # includes QR code
```

Client configs are saved to `.generated/vpn-clients/`. Import them into the WireGuard app on each device, then delete the `.conf` files from your workstation — they contain the client's private key and preshared key. The `.generated/` directory is gitignored and the files are created with `0600` permissions, but they should be treated as sensitive and not kept around longer than needed.

### Deploy applications

Deploy the reverse proxy, then each media app:

```bash
iac deploy proxy
iac deploy app --name plex
iac deploy app --name sonarr
iac deploy app --name radarr
iac deploy app --name bazarr
```

Each command runs an Ansible playbook that deploys the container into the VM with a read-only root filesystem, restart policy, and proper volume mounts. Traefik routes traffic to the containers automatically.

### Verify

```bash
iac status                  # cluster overview: hosts, containers, VPN
iac audit security          # run security checks across all infrastructure
```

After deployment, access apps from the home LAN (or over VPN):

| App    | URL                             |
|--------|---------------------------------|
| Plex   | `http://<host-ip>:32400`        |
| Sonarr | `http://<host-ip>:8989`         |
| Radarr | `http://<host-ip>:7878`         |
| Bazarr | `http://<host-ip>:6767`         |

### Manage auto-updates

Apps are updated automatically by a systemd timer that compares Docker image digests and recreates containers when a new image is available.

```bash
iac update status            # show update status for all hosts
```

### Tear down

```bash
iac teardown                 # destroy all VMs via Terraform
```

## CLI reference

```
iac <command> [options]

Commands:
  provision host       Provision a physical host (hardening, hypervisor, VM)
  deploy vpn           Deploy WireGuard VPN on designated host
  deploy proxy         Deploy reverse proxy into workload VMs
  deploy app           Deploy a media application (plex, sonarr, radarr, bazarr)
  generate configs     Generate all tool-specific configs from homelab.yml
  generate vpn-client  Generate a WireGuard client config
  status               Show cluster status (hosts, containers, VPN)
  update status        Show auto-update status for all hosts
  audit security       Run security audit across all infrastructure
  teardown             Tear down all VMs
  version              Print version
```

All commands read configuration from `homelab.yml` at the repository root.

## Project structure

```
homelab.yml.example          # example config — copy to homelab.yml
iac/
  cli/                       # Go CLI source
    main.go                  # entrypoint and command routing
    cmd/                     # command implementations
    config/                  # config loader and validator
    generators/              # Ansible/Terraform config generators
    wireguard/               # WireGuard key/peer management
  ansible/
    playbooks/               # setup-host, configure-vm, deploy-app,
                             # deploy-proxy, deploy-vpn, deploy-updater,
                             # security-audit
    roles/                   # hardening, hypervisor, vm-guest, wireguard,
                             # app-container, proxy-container, auto-updater
  terraform/                 # libvirt VM lifecycle
tests/
  integration/               # e2e deployment test script
.generated/                  # auto-generated configs (gitignored)
```

## Scaling to new hardware

When you migrate to beefier servers:

1. Update `homelab.yml` with the new host IPs and storage paths
2. Run `iac provision host` for each new host
3. Deploy apps to the new hosts
4. No changes to application configs needed

## Tech stack

| Component   | Tool                  | Why                                          |
|-------------|-----------------------|----------------------------------------------|
| CLI         | Go                    | Single binary, no runtime dependencies       |
| Provisioning| Ansible               | Agentless, SSH-based, idempotent             |
| VM lifecycle| Terraform + libvirt   | Declarative, reproducible                    |
| Containers  | Docker                | Industry standard, LinuxServer.io images     |
| VPN         | WireGuard             | Kernel-level, no third-party trust           |
| Proxy       | Traefik               | Auto-discovery, TLS, security headers        |
| OS          | Debian 12             | Stable, security updates, KVM support        |
