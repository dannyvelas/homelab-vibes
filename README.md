# homelab-vibe

Infrastructure as Code for a homelab environment. A single Go CLI (`iac`) reads one config file (`homelab.yml`) and orchestrates Terraform and Ansible to provision one or more servers, deploy a WireGuard VPN, configure OVN networking, and schedule containerized workloads via k3s. Engineers deploy any dockerized service by writing a simple manifest — no application-specific code in the platform itself.

## Architecture

```
Home LAN (192.168.1.0/24)
  |
  +-- Home Gateway/Router
  |     \-- UDP 51820 forwarded -> Host 01
  |         TCP 443 forwarded -> Host 01
  |
  +-- Host 01 (192.168.1.10)
  |     +-- WireGuard (wg0, VPN endpoint)
  |     +-- UFW (NAT, port forwarding, egress blocking)
  |     +-- OVN (overlay networking between hosts)
  |     +-- k3s server (workload scheduler)
  |     \-- VM (192.168.122.50) <- private subnet, not on LAN
  |           +-- k3s agent
  |           +-- Traefik (reverse proxy, subdomain routing)
  |           \-- workload containers (read-only root)
  |
  +-- Host 02 (192.168.1.11)
  |     +-- UFW (NAT, port forwarding, egress blocking)
  |     +-- OVN (overlay networking between hosts)
  |     +-- k3s server
  |     \-- VM (192.168.123.50) <- private subnet, not on LAN
  |           +-- k3s agent
  |           +-- Traefik (reverse proxy, subdomain routing)
  |           \-- workload containers (read-only root)
  |
  +-- ...more hosts as needed...
```

**Two-layer isolation**: Trusted infrastructure (WireGuard, KVM, OVN) runs on the host. Application workloads run inside VMs behind NAT. A container escape lands in the VM kernel, not the host.

### Security layers

1. **OS hardening** — UFW firewall, SSH key-only auth, unattended security updates
2. **NAT networking** — VMs on private subnets, invisible to the home LAN
3. **Egress blocking** — VMs cannot initiate connections to other LAN devices
4. **VM kernel isolation** — Container breakouts are contained by the VM boundary
5. **OVN overlay** — Encrypted east-west traffic between hosts, isolated from the physical network
6. **Reverse proxy** — Single ingress point with TLS, security headers, rate limiting
7. **Read-only containers** — Immutable root filesystem, writable only for data volumes

### Monitoring and alerting

- **Grafana dashboard** — real-time metrics for all deployed services (resource usage, uptime, response times)
- **Alerting** — automatic notifications when any service goes down or degrades (email, Slack, PagerDuty)
### Go links

Internal short URLs for quick access to services and dashboards:

| Link | Destination |
|------|-------------|
| `go/grafana` | Monitoring dashboard |
| `go/alerts` | Alert configuration |
| `go/vpn` | VPN client setup guide |

## Prerequisites

### Hardware

- One or more Debian 12 servers (laptops or bare-metal), each with:
  - 8 GB+ RAM
  - Dual-core CPU with Intel VT-x
  - 128 GB+ storage
  - Ethernet to home LAN

### Software (on your workstation)

- [Go](https://go.dev/) 1.21+ (to build the CLI)
- [Terraform](https://www.terraform.io/) (latest stable)
- [Ansible](https://docs.ansible.com/) (latest stable)
- [WireGuard client](https://www.wireguard.com/install/) (for VPN access)
- SSH client with key-based auth configured to all servers

### Network preparation

Forward UDP port **51820** and TCP port **443** on your home gateway/router to the server that will act as the WireGuard endpoint and ingress point. If your home IP is dynamic, set up a dynamic DNS hostname (e.g., DuckDNS, No-IP).

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

Edit `homelab.yml` — this is the **only file you need to touch** for infrastructure:

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
  # add as many hosts as needed

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

secrets:
  ssh_key_path: ~/.ssh/id_ed25519
  ssh_user: admin
```

The `iac` CLI reads this file and generates all tool-specific configs (Ansible inventory, Terraform tfvars) into `.generated/`. You never edit those files directly.

## Deploying services

This platform is application-agnostic. After infrastructure is set up, deploy any service using standard Kubernetes manifests and kubectl. See `services/` for full working examples.

k3s handles scheduling and restarts. Traefik picks up new services automatically via Ingress resources and routes by subdomain.

```bash
kubectl apply -f services/plex.yml
kubectl apply -f services/sonarr.yml
kubectl apply -f services/radarr.yml
kubectl apply -f services/bazarr.yml
kubectl apply -f services/grafana.yml
kubectl apply -f services/golinks.yml
```

## Usage

### Bootstrap infrastructure

Provision each server — this hardens the OS, installs KVM/libvirt, creates a workload VM, configures OVN overlay networking, and joins k3s:

```bash
iac provision host --name homelab-host-01
iac provision host --name homelab-host-02
# repeat for each host in homelab.yml
```

### Deploy VPN

Install WireGuard on the designated host for secure remote access:

```bash
iac deploy vpn
```

Generate client configs for your team:

```bash
iac generate vpn-client --name "danny-laptop"
iac generate vpn-client --name "danny-phone"
```

Client configs are saved to `.generated/vpn-clients/`. Import them into the WireGuard app on each device, then delete the `.conf` files from your workstation — they contain the client's private key and preshared key. The `.generated/` directory is gitignored and the files are created with `0600` permissions, but they should be treated as sensitive and not kept around longer than needed.

### Deploy reverse proxy

```bash
iac deploy proxy
```

Traefik routes all traffic through port 443, routing to services by subdomain.

### Verify

```bash
iac status         # cluster overview: hosts, services, VPN, k3s
iac audit security # run security checks across all infrastructure
```

### Tear down

```bash
iac teardown # destroy all VMs via Terraform
```

## CLI reference

```
iac <command> [options]

Commands:
  provision host       Provision a physical host (hardening, hypervisor, VM, OVN, k3s)
  deploy vpn           Deploy WireGuard VPN on designated host
  deploy proxy         Deploy reverse proxy into workload VMs
  generate configs     Generate all tool-specific configs from homelab.yml
  generate vpn-client  Generate a WireGuard client config
  status               Show cluster status (hosts, services, VPN, k3s)
  audit security       Run security audit across all infrastructure
  teardown             Tear down all VMs
  version              Print version
```

All commands read infrastructure configuration from `homelab.yml` at the repository root. Services are deployed with `kubectl` using standard Kubernetes manifests.

## Project structure

```
homelab.yml.example          # example config — copy to homelab.yml
services/                    # service manifests (one per app)
iac/
  cli/                       # Go CLI source
    main.go                  # entrypoint and command routing
    cmd/                     # command implementations
    config/                  # config loader and validator
    generators/              # Ansible/Terraform config generators
    wireguard/               # WireGuard key/peer management
  ansible/
    playbooks/               # setup-host, configure-vm, deploy-proxy,
                             # deploy-vpn, security-audit
    roles/                   # hardening-common, hardening-host, hypervisor,
                             # vm-guest, wireguard, proxy-container, k3s, ovn
  terraform/                 # libvirt VM lifecycle
tests/
  integration/               # e2e deployment test script
.generated/                  # auto-generated configs (gitignored)
```

## Scaling to new hardware

When you add or migrate servers:

1. Update `homelab.yml` with the new host IPs and storage paths
2. Run `iac provision host` for each new host
3. k3s automatically joins the new node to the cluster
4. OVN extends the overlay network to the new host
5. Deploy services with `kubectl` — no changes to manifests needed, k3s schedules across the cluster

## Tech stack

| Component    | Tool                  | Why                                          |
|--------------|-----------------------|----------------------------------------------|
| CLI          | Go                    | Single binary, no runtime dependencies       |
| Provisioning | Ansible               | Agentless, SSH-based, idempotent             |
| VM lifecycle | Terraform + libvirt   | Declarative, reproducible                    |
| Scheduling   | k3s                   | Lightweight Kubernetes, rolling updates, auto-restart |
| Networking   | OVN                   | Overlay networking, encrypted east-west traffic |
| Firewall     | UFW                   | Simple, readable firewall rules              |
| VPN          | WireGuard             | Kernel-level, no third-party trust           |
| Proxy        | Traefik               | Auto-discovery, TLS, subdomain routing       |
| Monitoring   | Grafana + Prometheus  | Dashboards, alerting, service health         |
| Go links     | golinks               | Internal short URLs for quick service access |
| OS           | Debian 12             | Stable, security updates, KVM support        |
