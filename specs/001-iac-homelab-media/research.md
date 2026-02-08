# Research Findings: IaC Homelab Media

**Feature Branch**: `001-iac-homelab-media`  
**Created**: 2026-02-07  
**Status**: Completed  
**Input**: Phase 0 Research tasks from plan.md, user clarifications from spec.md

## Resolved Clarifications & Research Decisions

### 1. IaC Strategy and Integration

-   **Decision**: Terraform will be used for initial server provisioning and management of infrastructure resources, while Ansible will be utilized for configuration management, OS hardening, and application deployment (including Docker/Docker Compose for the media stack).
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

### 3. Tailscale Deployment and Management

-   **Decision**: Tailscale will be deployed and configured on Debian servers using Ansible. Node authentication will initially leverage pre-authenticated keys for automation, with plans to explore more dynamic/secure non-interactive authentication methods for long-term production use. Subnet routers will be configured to provide remote access to the home LAN resources.
-   **Rationale**: Tailscale was explicitly requested by the user, aligns with the "Easy/Automated End-User Experience" principle, and provides a modern, secure VPN solution. Ansible is well-suited for installing and managing Tailscale.
-   **Alternatives Considered**: OpenVPN, WireGuard (rejected as Tailscale was preferred by the user for simplicity and ease of use).

### 4. Media Stack Deployment and Isolation

-   **Decision**: A small number of VMs (1-2 per physical host) will be provisioned using a lightweight hypervisor (KVM/libvirt). Media applications (Plex, Sonarr, Radarr, Bazarr) will run as Docker containers inside these VMs, managed via Docker Compose. Application-to-VM assignment is statically defined in Ansible inventory/group vars -- no runtime scheduler is needed at this scale. Ansible will manage VM creation, VM OS hardening, Docker installation, and Docker Compose deployments within each VM.
-   **Rationale**: The original one-VM-per-app approach is not viable on the target hardware (8GB RAM, dual-core i7, 128-256GB storage). Each VM incurs ~512MB-1GB of overhead for its kernel and base OS services; four or more VMs would consume most available RAM before any applications run, and CPU context-switching across many VMs on a dual-core would degrade performance. The revised approach preserves the key defense-in-depth benefit -- applications are isolated from the host kernel by the VM boundary -- while keeping resource usage practical. Docker Compose inside VMs provides easy application lifecycle management (start, stop, update, rollback). When migrating to beefier bare-metal hardware, the architecture scales naturally: add more VMs or redistribute containers by updating Ansible inventory, with no structural changes required.
-   **Alternatives Considered**: One VM per application (not viable on constrained hardware -- excessive RAM/CPU/storage overhead); Docker containers directly on the host (no VM-level kernel isolation); Native package installation (less isolated, dependency conflicts); Runtime container scheduler/orchestrator (over-engineering for 2 hosts and ~4 applications -- Kubernetes/Nomad patterns are appropriate at larger scale).

### 5. Custom Logic Language Integration

-   **Decision**: Go will be the primary language for any custom scripting or orchestration logic required beyond the capabilities of Ansible and Terraform.
-   **Rationale**: Go was one of the user's preferred languages, and its performance, static typing, and strong ecosystem for CLI tools and backend services make it a pragmatic choice for infrastructure automation within this context.
-   **Alternatives Considered**: Nim, OCaml (also user preferences, but Go has more established libraries and community support in the infrastructure automation space).

## Key Decisions Summary

-   **OS**: Debian
-   **IaC Tools**: Terraform (provisioning), Ansible (configuration/deployment)
-   **VPN**: Tailscale
-   **Media Stack Deployment**: Docker containers inside VMs (1-2 VMs per host, static assignment via Ansible)
-   **Custom Scripting**: Go
