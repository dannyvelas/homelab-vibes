# Research Findings: IaC Homelab Media

**Feature Branch**: `001-iac-homelab-media`  
**Created**: 2026-02-07  
**Status**: Completed  
**Input**: Phase 0 Research tasks from plan.md, user clarifications from spec.md

## Resolved Clarifications & Research Decisions

### 1. IaC Strategy and Integration

-   **Decision**: Terraform will be used for initial server provisioning and management of infrastructure resources, while Ansible will be utilized for configuration management, OS hardening, Nomad/Docker setup, and application deployment.
-   **Rationale**: This combination leverages Terraform's strengths in declarative infrastructure state management and Ansible's powerful capabilities for idempotent software configuration and orchestration. It provides a robust, flexible, and well-supported IaC solution that aligns with the user's preference for these tools.
-   **Alternatives Considered**: Pure Ansible (less suitable for initial bare-metal provisioning and state management); Pure Terraform (less suitable for intricate software configurations and application deployments); Custom scripting (higher initial development effort to match capabilities of established tools).

### 2. Debian Server Hardening

-   **Decision**: Comprehensive Debian server hardening will be implemented via Ansible playbooks. This includes:
    -   Configuring Uncomplicated Firewall (UFW) to limit inbound/outbound traffic.
    -   Securing SSH access (disabling password authentication, preventing root login, enforcing key-based authentication).
    -   Implementing automatic security updates.
    -   Managing users and groups with the principle of least privilege.
-   **Rationale**: Directly addresses the "Defense-in-Depth Security" principle, enhancing the security posture of the base operating system. Ansible provides an idempotent and auditable way to enforce these configurations.
-   **Alternatives Considered**: Manual hardening (prone to human error, not repeatable); other hardening tools (Ansible provides sufficient capabilities for this scope).

### 2a. Network Segmentation and Traffic Control

-   **Decision**: All user-facing/internet-facing applications will be placed on a dedicated VLAN, isolated from other LAN devices. Egress traffic from the application VLAN will be blocked from reaching other devices on the home network, preventing lateral movement in the event of a compromise. VLAN configuration will be managed via Ansible, with firewall rules (UFW or iptables) enforcing egress restrictions at the VM and/or host level.
-   **Rationale**: VLAN isolation provides network-layer segmentation that is independent of the VM/container boundary. Even if an attacker escapes the container, they are still confined to a network segment that cannot reach other devices. Egress blocking specifically addresses lateral movement -- a compromised Plex instance cannot scan or attack other hosts on the LAN. Together, these form a critical layer in the defense-in-depth strategy.
-   **Alternatives Considered**: Flat network with firewall rules only (less robust -- a misconfigured rule exposes the entire LAN); network-level micro-segmentation per container (overly complex for this scale, better suited to service mesh architectures).

### 3. Tailscale Deployment and Management

-   **Decision**: Tailscale will be deployed and configured on Debian servers using Ansible. Node authentication will initially leverage pre-authenticated keys for automation, with plans to explore more dynamic/secure non-interactive authentication methods for long-term production use. Subnet routers will be configured to provide remote access to the home LAN resources.
-   **Rationale**: Tailscale was explicitly requested by the user, aligns with the "Easy/Automated End-User Experience" principle, and provides a modern, secure VPN solution. Ansible is well-suited for installing and managing Tailscale.
-   **Alternatives Considered**: OpenVPN, WireGuard (rejected as Tailscale was preferred by the user for simplicity and ease of use).

### 4. Media Stack Deployment, Isolation, and Scheduling

-   **Decision**: Each physical host runs two layers:
    1.  **Host layer**: Nomad agent (server + client), Tailscale, and system infrastructure run directly on the host OS. These are trusted, operator-controlled services with no user-facing attack surface.
    2.  **VM layer**: One VM per host, provisioned via KVM/libvirt, placed on a dedicated VLAN (see 2a). The VM runs a Nomad client and Docker. All user-facing media applications (Plex, Sonarr, Radarr, Bazarr) run as Docker containers inside the VM, scheduled by Nomad. The VM provides kernel-level isolation between user-facing applications and the host.

    All user-facing applications are additionally:
    -   **Behind a reverse proxy**: A reverse proxy (e.g., Traefik or Caddy) runs inside the VM as a Nomad-scheduled container, providing a single ingress point for all media services. It terminates TLS, enforces rate limiting, and prevents direct exposure of application ports.
    -   **On a dedicated VLAN**: The VM's network interface is on an isolated VLAN, segmented from other LAN devices.
    -   **Egress-blocked**: Firewall rules prevent the application VLAN from initiating outbound connections to other devices on the home network, blocking lateral movement.
    -   **Read-only filesystem**: Docker containers are run with `--read-only` flag. Only explicitly required data directories (media libraries, config, databases) are mounted as writable volumes. This limits post-compromise persistence -- an attacker cannot drop binaries or modify application code.

    Nomad handles container scheduling and placement across the cluster. On the current 2-host setup, Nomad targets the VM's Docker daemon to schedule media workloads. Application definitions are Nomad job files (HCL). Ansible manages host OS hardening, hypervisor setup, VM creation, VM OS hardening, Nomad installation (on both host and VM), and Docker installation within the VM.
-   **Rationale**: This architecture implements defense-in-depth across multiple independent layers: network (VLAN + egress blocking), application ingress (reverse proxy), kernel (VM boundary), container (Docker isolation + read-only filesystem). Each layer mitigates a different class of attack, and no single layer's failure compromises the entire system. The VM boundary ensures that a container escape exploit (e.g., a zero-day in Plex) lands in the VM's kernel, not the host's -- an attacker would need a second exploit (VM escape) to reach the host, which is significantly harder. VLAN isolation and egress blocking ensure that even a VM-level compromise cannot reach other LAN devices. The reverse proxy eliminates direct application port exposure and provides centralized TLS and rate limiting. Read-only containers prevent persistent malware installation. Running Nomad and Tailscale directly on the host avoids wasting VM overhead on trusted infrastructure. Nomad adds ~500-750MB RAM overhead, which is affordable on 8GB when only one VM is running. Critically, Nomad is being built now to support future scaling: when more hardware is available, additional VMs or physical nodes each run a Nomad client, and Nomad distributes workloads across them automatically. The Nomad job definitions remain unchanged -- only the cluster membership grows.
-   **Alternatives Considered**: One VM per application (not viable on constrained hardware -- excessive RAM/CPU/storage overhead); Docker containers directly on host without VM (no kernel isolation -- unacceptable given threat model of malicious users and zero-day exploits in user-facing apps); Docker Compose without a scheduler (simpler but does not scale -- static assignment cannot automatically redistribute workloads as the cluster grows); Kubernetes/k3s (heavier than Nomad at ~750MB-1.5GB, more complex for this scale, though viable on future hardware); Native package installation (less isolated, dependency conflicts).

### 5. Custom Logic Language Integration

-   **Decision**: Go will be the primary language for any custom scripting or orchestration logic required beyond the capabilities of Ansible and Terraform.
-   **Rationale**: Go was one of the user's preferred languages, and its performance, static typing, and strong ecosystem for CLI tools and backend services make it a pragmatic choice for infrastructure automation within this context.
-   **Alternatives Considered**: Nim, OCaml (also user preferences, but Go has more established libraries and community support in the infrastructure automation space).

### 6. Container Scheduling

-   **Decision**: HashiCorp Nomad will serve as the container scheduler/orchestrator.
-   **Rationale**: Nomad is lightweight (~500-750MB RAM for server + client), has a native Docker task driver, mature Terraform and Ansible integration, and scales from 2 nodes to thousands without architectural changes. Its HCL-based job files are simpler than Kubernetes YAML manifests. Nomad is purpose-built for workload scheduling and avoids the operational complexity of a full Kubernetes deployment, making it ideal for a resource-constrained homelab that needs to scale later.
-   **Alternatives Considered**: Kubernetes/k3s (heavier resource footprint, more operational complexity than warranted at this scale); Docker Compose with static assignment (simpler but does not scale -- no automatic workload distribution); Docker Swarm (essentially in maintenance mode, limited future investment); Custom Go scheduler (significant development effort to replicate what Nomad provides out of the box).

## Key Decisions Summary

-   **OS**: Debian
-   **IaC Tools**: Terraform (provisioning), Ansible (configuration/deployment)
-   **VPN**: Tailscale (runs on host)
-   **Container Scheduler**: Nomad (runs on host; client also runs inside VM)
-   **Media Stack Deployment**: Docker containers (read-only) inside VM, scheduled by Nomad (1 VM per host for user-facing workloads)
-   **Network Security**: Dedicated VLAN for application VM, egress blocking to prevent lateral movement
-   **Ingress**: Reverse proxy (Traefik/Caddy) inside VM for TLS termination, rate limiting, single ingress point
-   **Custom Scripting**: Go
