# Research Findings: IaC Homelab Media

**Feature Branch**: `001-iac-homelab-media`  
**Created**: 2026-02-07  
**Status**: Completed  
**Input**: Phase 0 Research tasks from plan.md, user clarifications from spec.md

## Resolved Clarifications & Research Decisions

### 1. IaC Strategy and Integration

-   **Decision**: The IaC uses a two-layer architecture with clear separation of concerns:

    **Terraform (lifecycle management)**: Manages the lifecycle of infrastructure resources that expose APIs — specifically VMs via the `dmacvicar/libvirt` Terraform provider, and Nomad workloads via the `hashicorp/nomad` Terraform provider. Terraform runs on the engineer's workstation (zero RAM cost on the servers) and communicates with libvirt and Nomad APIs over the network. This gives the system declarative create-and-destroy semantics: `terraform apply` creates VMs and deploys Nomad jobs, `terraform destroy` cleanly tears them down, and Terraform state tracks exactly what resources exist.

    **Ansible (configuration management)**: Manages everything that doesn't have an API and requires SSH-based convergence — OS hardening, package installation, service configuration, iptables rules, WireGuard setup, and initial bootstrapping of KVM/libvirt and Nomad on fresh hosts. Ansible playbooks are idempotent and can be safely re-run to converge a host to the desired state.

    The two layers map to the infrastructure like this:
    ```
    ┌─────────────────────────────────────────────────┐
    │  Terraform (engineer's workstation)             │
    │  ├── libvirt provider  → VM lifecycle           │
    │  └── nomad provider    → workload lifecycle     │
    └──────────────┬──────────────────┬───────────────┘
                   │ libvirt API      │ Nomad API
    ┌──────────────▼──────────────────▼───────────────┐
    │  Host layer (acts as "cloud provider")          │
    │  ├── KVM/libvirt  → VM create/destroy/snapshot  │
    │  ├── Nomad        → workload scheduling         │
    │  └── WireGuard    → VPN                         │
    └──────────────┬──────────────────────────────────┘
                   │ SSH
    ┌──────────────▼──────────────────────────────────┐
    │  Ansible (engineer's workstation)               │
    │  └── OS hardening, packages, services, iptables │
    └─────────────────────────────────────────────────┘
    ```

    This mirrors how companies use Terraform with cloud providers — except here, libvirt and Nomad serve as the on-premise "cloud provider" APIs. The same Terraform workflow (`plan`, `apply`, `destroy`) works regardless of whether the backing infrastructure is AWS, GCP, or a homelab running libvirt. This means the Terraform layer is portable: if the team later adopts a bare-metal provisioning system (e.g., MAAS) or moves workloads to a cloud provider, they add a new Terraform provider without changing the Ansible configuration layer.

-   **Rationale**: The key reason for using both tools (rather than Ansible alone) is **lifecycle management**. Ansible excels at configuring systems but has no built-in concept of "I own these resources and can cleanly destroy them." To tear down a VM or remove a Nomad job with Ansible, you'd need to write and maintain separate teardown playbooks — the reverse of every creation step. Terraform solves this with state-tracked lifecycle: it knows what it created and can surgically remove it via `terraform destroy`. This directly supports FR-011 (easy/automated UX) — tearing down infrastructure should be as easy as standing it up. The separation also makes the codebase clearer: Terraform files declare *what resources should exist*, Ansible playbooks declare *how those resources should be configured*.

-   **Alternatives Considered**: Pure Ansible (can create resources but lacks clean destroy/lifecycle management — no state tracking means teardown requires manually maintained reverse playbooks; viable for simple setups but increasingly fragile as infrastructure grows); Pure Terraform (cannot SSH into servers and configure OS, install packages, or manage services — it's a provisioning tool, not a configuration management tool); NixOS declarative configuration (the OS itself becomes IaC — interesting but requires learning the Nix language, has a steep learning curve, and would replace Ansible entirely with a different paradigm); Custom Go tooling (high initial development effort to replicate what Terraform + Ansible provide out of the box); Pulumi (like Terraform but uses real programming languages like Go — viable but smaller ecosystem of providers and community content than Terraform).

### 2. Debian Server Hardening

-   **Decision**: Comprehensive Debian server hardening will be implemented via Ansible playbooks. This includes:
    -   Configuring Uncomplicated Firewall (UFW) to limit inbound/outbound traffic.
    -   Securing SSH access (disabling password authentication, preventing root login, enforcing key-based authentication).
    -   Implementing automatic security updates.
    -   Managing users and groups with the principle of least privilege.
-   **Rationale**: Directly addresses the "Defense-in-Depth Security" principle, enhancing the security posture of the base operating system. Ansible provides an idempotent and auditable way to enforce these configurations.
-   **Alternatives Considered**: Manual hardening (prone to human error, not repeatable); other hardening tools (Ansible provides sufficient capabilities for this scope).

### 2a. Network Segmentation and Traffic Control

-   **Decision**: VMs will be attached to a Linux virtual bridge on the host, with a private NAT subnet (e.g., `192.168.122.0/24`) distinct from the home LAN CIDR range. The host acts as the gateway/router for the VM subnet. Iptables rules on the host enforce:
    1.  **Ingress**: Only explicitly forwarded ports (e.g., 443) reach the VM's reverse proxy. VMs are not directly reachable from the LAN.
    2.  **Egress**: VM traffic to the home LAN is denied by default. Only explicitly allowed outbound traffic is permitted (internet for media downloads, DNS, NTP).
    3.  **Lateral movement prevention**: The VM cannot initiate connections to other devices on the home network.

    All network configuration (bridge setup, iptables rules, NAT, port forwarding) is managed via Ansible.
-   **Rationale**: A virtual bridge with NAT provides network-layer isolation without requiring any special network hardware (no managed switch, no 802.1Q VLAN support needed). Since the isolation boundary is host-to-VM, a software-defined bridge is sufficient and fully manageable via IaC. The host is the sole gateway for VM traffic, making it a natural enforcement point for all firewall rules -- a single place to audit and control access. VMs are invisible to the LAN by default (private subnet), so isolation is the starting state rather than something that must be configured on external equipment. Even if an attacker escapes the container and the VM, they land on a private subnet with no route to other LAN devices.
-   **Alternatives Considered**: VLAN (802.1Q) segmentation (requires managed switch hardware, more complex, overkill when the isolation boundary is a single host-to-VM relationship); flat network with firewall rules only (less robust -- VMs would be on the home LAN by default, and a misconfigured rule exposes the entire network); bridged networking where VMs get home LAN IPs (weaker isolation -- VMs directly reachable from LAN, egress control requires external switch ACLs).

### 3. VPN Deployment and Management

-   **Decision**: WireGuard will be deployed directly on Debian servers using Ansible. The server-side configuration (key generation, interface setup, peer management, iptables rules for forwarding) is fully automated. Client configuration files are generated by the IaC and distributed to team members. A single UDP port (default 51820) is forwarded on the home gateway to the designated WireGuard host. If the home IP is dynamic, a lightweight dynamic DNS solution (e.g., ddclient with a free DDNS provider) is deployed alongside WireGuard.
-   **Rationale**: WireGuard is the best fit for the spec's requirements because: (1) **Portability (FR-010)**: The entire VPN configuration lives in the IaC repo with zero external dependencies — no third-party accounts or SaaS services required to reproduce the setup on new hardware. (2) **Reliability (SC-003)**: No external coordination servers — if the host is up, VPN is up. (3) **Defense-in-depth (Constitution IV)**: No third-party trust in the network layer. WireGuard is a ~4000 LOC kernel module, formally verified (Noise protocol framework), with a minimal attack surface. Keys are generated and managed locally. (4) **Resource efficiency**: Kernel-level implementation means essentially zero RAM/CPU overhead, critical on 8GB hosts. (5) **Automation-friendly**: Fully configurable via Ansible (wg, wg-quick, systemd units). The home gateway port forward is the only manual step, and it's a one-time operation since the engineer has physical access to the gateway.
-   **Alternatives Considered**: Tailscale (excellent NAT traversal and client UX, but depends on a third-party SaaS control plane for coordination — conflicts with FR-010 portability, SC-003 reliability, and Constitution IV defense-in-depth; the NAT traversal advantage is unnecessary when the engineer controls the home gateway); Headscale (open-source Tailscale control plane — eliminates the SaaS dependency but adds a significant additional component to deploy and maintain, negating the simplicity advantage); OpenVPN (mature but heavier, slower than WireGuard, larger attack surface, more complex configuration); Netbird (open-source Tailscale alternative — less mature, smaller community).

### 4. Media Stack Deployment, Isolation, and Scheduling

-   **Decision**: Each physical host runs two layers:
    1.  **Host layer**: Nomad agent (server + client), WireGuard, and system infrastructure run directly on the host OS. These are trusted, operator-controlled services with no user-facing attack surface.
    2.  **VM layer**: One VM per host, provisioned via KVM/libvirt, attached to a private virtual bridge with NAT (see 2a). The VM runs a Nomad client and Docker. All user-facing media applications (Plex, Sonarr, Radarr, Bazarr) run as Docker containers inside the VM, scheduled by Nomad. The VM provides kernel-level isolation between user-facing applications and the host.

    All user-facing applications are additionally:
    -   **Behind a reverse proxy**: A reverse proxy (e.g., Traefik or Caddy) runs inside the VM as a Nomad-scheduled container, providing a single ingress point for all media services. It terminates TLS, enforces rate limiting, and prevents direct exposure of application ports. LAN/WireGuard users reach apps via the host IP; iptables on the host forwards designated ports to the reverse proxy inside the VM.
    -   **On an isolated private subnet**: The VM's network interface is on a virtual bridge with a private NAT subnet (e.g., `192.168.122.0/24`), not directly reachable from the home LAN. The host is the sole gateway.
    -   **Egress-blocked**: Iptables rules on the host deny VM-to-LAN traffic by default. Only explicitly allowed outbound traffic is permitted (internet for media downloads, DNS, NTP).
    -   **Read-only root filesystem**: Docker containers are run with `--read-only` flag, which makes the container's root filesystem (application binaries, system libraries) immutable. Data directories that applications need to write to (`/config` for databases/logs, `/downloads` for staging, `/media` for organized libraries) are mounted as explicit writable volumes. This limits post-compromise persistence -- an attacker cannot drop binaries or modify application code on the root filesystem.

    Nomad handles container scheduling and placement across the cluster. On the current 2-host setup, Nomad targets the VM's Docker daemon to schedule media workloads. Application definitions are Nomad job files (HCL). Ansible manages host OS hardening, hypervisor setup, VM creation, VM OS hardening, Nomad installation (on both host and VM), and Docker installation within the VM.
-   **Rationale**: This architecture implements defense-in-depth across multiple independent layers: network (private NAT subnet + egress blocking), application ingress (reverse proxy with host-level port forwarding), kernel (VM boundary), container (Docker isolation + read-only root filesystem). Each layer mitigates a different class of attack, and no single layer's failure compromises the entire system. The VM boundary ensures that a container escape exploit (e.g., a zero-day in Plex) lands in the VM's kernel, not the host's -- an attacker would need a second exploit (VM escape) to reach the host, which is significantly harder. The private NAT subnet and egress blocking ensure that even a VM-level compromise cannot reach other LAN devices. The reverse proxy eliminates direct application port exposure and provides centralized TLS and rate limiting. Read-only root filesystems prevent persistent malware installation on the container image layers while still allowing applications to write to their required data directories. Running Nomad and WireGuard directly on the host avoids wasting VM overhead on trusted infrastructure. WireGuard runs as a kernel module with essentially zero RAM overhead. Nomad adds ~500-750MB RAM overhead, which is affordable on 8GB when only one VM is running. Critically, Nomad is being built now to support future scaling: when more hardware is available, additional VMs or physical nodes each run a Nomad client, and Nomad distributes workloads across them automatically. The Nomad job definitions remain unchanged -- only the cluster membership grows.
-   **Alternatives Considered**: One VM per application (not viable on constrained hardware -- excessive RAM/CPU/storage overhead); Docker containers directly on host without VM (no kernel isolation -- unacceptable given threat model of malicious users and zero-day exploits in user-facing apps); Docker Compose without a scheduler (simpler but does not scale -- static assignment cannot automatically redistribute workloads as the cluster grows); Kubernetes/k3s (heavier than Nomad at ~750MB-1.5GB, more complex for this scale, though viable on future hardware); Native package installation (less isolated, dependency conflicts).

### 5. Custom Logic Language Integration

-   **Decision**: Go will be the primary language for any custom scripting or orchestration logic required beyond the capabilities of Ansible and Terraform.
-   **Rationale**: Go was one of the user's preferred languages, and its performance, static typing, and strong ecosystem for CLI tools and backend services make it a pragmatic choice for infrastructure automation within this context.
-   **Alternatives Considered**: Nim, OCaml (also user preferences, but Go has more established libraries and community support in the infrastructure automation space).

### 6. Container Scheduling

-   **Decision**: HashiCorp Nomad will serve as the container scheduler/orchestrator.
-   **Rationale**: Nomad is lightweight (~500-750MB RAM for server + client), has a native Docker task driver, mature Terraform and Ansible integration, and scales from 2 nodes to thousands without architectural changes. Its HCL-based job files are simpler than Kubernetes YAML manifests. Nomad is purpose-built for workload scheduling and avoids the operational complexity of a full Kubernetes deployment, making it ideal for a resource-constrained homelab that needs to scale later.
-   **Alternatives Considered**: Kubernetes/k3s (heavier resource footprint, more operational complexity than warranted at this scale); Docker Compose with static assignment (simpler but does not scale -- no automatic workload distribution); Docker Swarm (essentially in maintenance mode, limited future investment); Custom Go scheduler (significant development effort to replicate what Nomad provides out of the box).

### 7. Application Auto-Updates

-   **Decision**: Media applications (Plex, Sonarr, Radarr, Bazarr) are automatically updated using a Nomad-native approach. Each application's Nomad job uses a Docker image tag (e.g., `linuxserver/plex:latest`). A lightweight periodic Nomad batch job runs on a schedule (e.g., daily at 3 AM) that:
    1.  Pulls the latest image for each media app (`docker pull`).
    2.  Compares the new image digest to the currently running digest.
    3.  If a new version is detected, triggers a Nomad job deployment for the affected application.

    Nomad's built-in `update` stanza in each job definition controls the rollout:
    -   **Health check**: Nomad waits for the new container to pass health checks (HTTP check on the app's web UI port) before marking the deployment as successful.
    -   **Auto-revert**: If the new container fails health checks within a configurable timeout, Nomad automatically reverts to the previous working version.
    -   **Rolling update**: One application updates at a time, so the entire media stack is never down simultaneously.

    Application data (`/config`, `/downloads`, `/media` volumes) is mounted from the host and persists across container replacements, so updates never cause data loss.

-   **Rationale**: A Nomad-native approach is chosen over alternatives because Nomad already manages container lifecycle. Using Nomad's deployment mechanism ensures the scheduler is always aware of what's running (no out-of-band container restarts that confuse Nomad's state). The `update` stanza with `auto_revert = true` directly satisfies the spec's requirement for automatic rollback on failure (US4 acceptance scenario 3). Health checks ensure the new version actually works before the old one is removed. Running the check as a periodic Nomad batch job keeps the update mechanism within the same scheduling system — no additional cron infrastructure needed.

-   **Alternatives Considered**: Watchtower (runs as a Docker container that watches for image updates and restarts containers directly — but this bypasses Nomad entirely, causing Nomad to see unexpected container restarts and potentially fight with Watchtower over container state; unacceptable when Nomad is the workload scheduler); Manual `iac update apps` command (simple but doesn't meet FR-012's "without manual intervention" requirement); OS-level cron job (works but adds infrastructure outside Nomad's management; the periodic batch job keeps everything in one system).

### 8. Unified Configuration (Single Source of Truth)

-   **Decision**: All user-configurable values are defined in a single configuration file at the repository root: `homelab.yml`. This file is the sole place an engineer edits configuration. The `iac` CLI wrapper reads `homelab.yml` and generates or passes the appropriate values to each underlying tool:

    -   **Ansible**: The `iac` CLI generates the Ansible inventory (`hosts.yml`, `group_vars/`) from `homelab.yml` before running playbooks. The engineer never edits the Ansible inventory directly.
    -   **Terraform**: The `iac` CLI generates a `terraform.tfvars` file from `homelab.yml` before running `terraform apply`. The engineer never edits `.tfvars` files directly.
    -   **Nomad**: The `iac` CLI templates Nomad HCL job files from `homelab.yml` values (e.g., media paths, resource limits, image tags) before submitting them. The engineer never edits HCL files directly for configuration values.

    The `homelab.yml` file is structured by concern:

    ```yaml
    # homelab.yml — single source of truth
    cluster:
      name: homelab
      datacenter: dc1

    hosts:
      homelab-host-01:
        ip: 192.168.1.10
        nat_subnet: 192.168.122.0/24
        wireguard_endpoint: true
      homelab-host-02:
        ip: 192.168.1.11
        nat_subnet: 192.168.123.0/24

    vpn:
      subnet: 10.0.0.0/24
      endpoint: home.example.com
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
      schedule: "0 3 * * *"  # daily at 3 AM
      auto_revert: true
      health_check_timeout: 5m

    secrets:
      ssh_key_path: ~/.ssh/id_ed25519
      ssh_user: admin
    ```

    The tool-specific files (`hosts.yml`, `terraform.tfvars`, `*.hcl`) are generated artifacts — they live in a `.generated/` directory (or are generated in-memory) and should not be manually edited. The `iac` CLI regenerates them on every run from `homelab.yml`.

-   **Rationale**: This directly satisfies FR-013 (single source of truth) and SC-008 (change a value in one place, all tools reflect it). The engineer's mental model is simple: "edit `homelab.yml`, run `iac` commands." They never need to know that Ansible uses a different inventory format, Terraform uses `.tfvars`, and Nomad uses HCL — the `iac` wrapper abstracts that away. This also reduces configuration drift errors: if a server IP changes, the engineer updates it in one place, not in three separate tool-specific files. The `homelab.yml` structure mirrors the data model entities (hosts, VPN, apps, storage), making it intuitive. Secrets that should not be committed to git (e.g., WireGuard private keys) are stored separately and referenced by path, or injected via environment variables.

-   **Alternatives Considered**: Ansible inventory as the source of truth (forces the engineer to learn Ansible's inventory format, and Terraform/Nomad configs would still need separate files or complex variable passing); Terraform variables as the source of truth (similar issue — Ansible would need to read Terraform output, adding a dependency chain); environment variables for everything (no single file to review, harder to version control, error-prone); separate config files per tool with a sync script (fragile, the sync script becomes a maintenance burden).

## Key Decisions Summary

-   **OS**: Debian
-   **IaC Tools**: Terraform (lifecycle management via libvirt + Nomad providers), Ansible (configuration management via SSH)
-   **VPN**: WireGuard (kernel module, runs on host)
-   **Container Scheduler**: Nomad (runs on host; client also runs inside VM)
-   **Media Stack Deployment**: Docker containers (read-only) inside VM, scheduled by Nomad (1 VM per host for user-facing workloads)
-   **Network Security**: Virtual bridge with private NAT subnet for application VMs, iptables egress blocking to prevent lateral movement, host-level port forwarding for ingress
-   **Ingress**: Reverse proxy (Traefik/Caddy) inside VM for TLS termination, rate limiting, single ingress point
-   **Auto-Updates**: Nomad-native periodic batch job + health-checked deployments with auto-revert
-   **Configuration**: Single `homelab.yml` file at repo root; `iac` CLI generates tool-specific configs
-   **Custom Scripting**: Go
