# Implementation Plan: IaC Homelab Media

**Branch**: `001-iac-homelab-media` | **Date**: 2026-02-07 | **Spec**: /Users/dannyvelasquez/RemoteGit/MyGithub/homelab-vibe/specs/001-iac-homelab-media/spec.md
**Input**: Feature specification from `/specs/001-iac-homelab-media/spec.md`

## Summary

This project develops an Infrastructure as Code (IaC) solution for a homelab environment, initially on two low-resource laptops, with a clear migration path to on-premise bare-metal servers. The architecture uses a two-layer design: trusted infrastructure (Nomad, Tailscale) runs directly on the host, while user-facing media applications (Plex, Sonarr, Radarr, Bazarr) run as Docker containers inside a KVM/libvirt VM per host, scheduled by Nomad. Defense-in-depth security is implemented across multiple independent layers: NAT networking with egress blocking, reverse proxy ingress, VM kernel isolation, and read-only container filesystems. Ansible handles configuration management and deployment, Terraform handles provisioning, and Nomad handles workload scheduling -- built now on constrained hardware so the scheduling infrastructure is ready when the cluster grows.

## Technical Context

**Language/Version**: Go (for custom scripting/logic within IaC if needed, per user preference for Go, Nim, OCaml)
**Primary Dependencies**: Ansible, Terraform, Tailscale, Nomad, KVM/libvirt, Docker
**Storage**: Host-level storage with passthrough to VM and containers (media libraries, config, downloads)
**Testing**: Unit testing for IaC code (e.g., Terratest, Molecule), integration testing for deployed services
**Target Platform**: Debian servers (bare-metal, 8GB RAM, dual-core Intel i7, 128GB/256GB storage)
**Project Type**: Single (IaC repository for multiple services)
**Performance Goals**: Fast, automated deployment; reliable Tailscale VPN access; responsive media applications within hardware constraints.
**Constraints**: Low-resource initial hardware (laptops), IaC portability for future bare-metal migration, extreme defense-in-depth security.
**Scale/Scope**: Two laptops initially, transitioning to dedicated bare-metal. Focus on VPN and media stack. Monitoring deferred.

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

-   **I. Easy/Automated End-User Experience**: **PASS** - Plan emphasizes IaC and automation for setup and deployment. Nomad job files provide declarative application definitions. Single-command engineer experience via wrapper scripts.
-   **II. Production Quality**: **PASS** - Focus on security, IaC testing, and future monitoring aligns. Nomad provides self-healing workload management. Acknowledgement of current hardware constraints, but striving for best possible reliability/performance.
-   **III. Scalability**: **PASS** - Nomad scales from 2 nodes to thousands without architectural changes. IaC with Ansible/Terraform facilitates portability to bare-metal. Adding capacity = adding Nomad clients, no job definition changes.
-   **IV. Defense-in-Depth Security**: **PASS** - Multiple independent security layers: NAT networking with egress blocking, reverse proxy ingress, VM kernel isolation, read-only container filesystems, OS hardening, least privilege. No single layer's failure compromises the system.

## Project Structure

### Documentation (this feature)

```text
specs/001-iac-homelab-media/
├── plan.md              # This file (/speckit.plan command output)
├── research.md          # Phase 0 output (/speckit.plan command)
├── data-model.md        # Phase 1 output (/speckit.plan command)
├── quickstart.md        # Phase 1 output (/speckit.plan command)
├── contracts/           # Phase 1 output (/speckit.plan command)
└── tasks.md             # Phase 2 output (/speckit.tasks command - NOT created by /speckit.plan)
```

### Source Code (repository root)

```text
iac/
├── terraform/           # Terraform configurations for server provisioning and VM management
├── ansible/
│   ├── inventory/       # Host inventory and group vars (application-to-VM assignments)
│   ├── playbooks/       # Playbooks for host setup, VM setup, app deployment
│   └── roles/           # Reusable roles (hardening, docker, nomad, tailscale, etc.)
├── nomad/
│   └── jobs/            # Nomad job files (HCL) for media apps and reverse proxy
├── scripts/             # Custom scripts (Go) for orchestration or specific tasks
└── docs/                # IaC-specific documentation

tests/
├── integration/         # Integration tests for deployed services
└── unit/                # Unit tests for IaC (Terraform, Ansible)
```

**Structure Decision**: A single project IaC repository structure. Terraform handles provisioning the initial OS-booted state and managing VMs via KVM/libvirt. Ansible handles host OS hardening, hypervisor setup, Nomad server/client installation on host, Tailscale deployment on host, VM creation (1 per host for user-facing workloads), VM OS hardening, Nomad client + Docker installation within VMs, and NAT networking configuration (iptables rules, port forwarding). Application definitions are Nomad job files (HCL) that define media containers with read-only root filesystems and writable data volume mounts. A reverse proxy (Traefik/Caddy) runs as a Nomad-scheduled container inside each VM. Custom Go scripts may be used for glue logic or specific automation tasks.

## Complexity Tracking

> **Fill ONLY if Constitution Check has violations that must be justified**

(No violations to justify at this stage.)

## Phase 0: Outline & Research

**Goal**: Establish foundational IaC strategy, secure OS configuration, and optimal media stack deployment methods given constraints.

1.  **IaC Strategy and Integration**:
    -   Research best practices for integrating Terraform (provisioning) and Ansible (configuration) in a homelab context.
    -   Determine how to manage state and secrets securely across both tools.
    -   Explore methods for creating a simple, single-command engineer experience using these tools.

2.  **Debian Server Hardening and Network Segmentation**:
    -   Research essential security configurations for Debian servers (firewall, SSH hardening, user management, update policies).
    -   Identify minimal Debian installation requirements for this use case.
    -   Design NAT networking for VMs: private subnet (e.g., `192.168.122.0/24`) distinct from the home LAN CIDR, host acts as gateway/router with iptables masquerading.
    -   Research iptables rules for ingress control (port forwarding from host to VM reverse proxy), egress blocking (deny VM-to-LAN by default, allow only internet/DNS/NTP), and lateral movement prevention.
    -   Determine NAT networking configuration approach via Ansible (libvirt network definitions, iptables rules).

3.  **Tailscale Deployment and Management**:
    -   Research automated deployment of Tailscale via Ansible/Terraform, including node authentication and subnets.
    -   Investigate secure methods for Tailscale key management within the IaC framework.

4.  **Hypervisor, Media Stack Isolation, and Scheduling**:
    -   Research lightweight hypervisor technologies suitable for Debian on limited hardware (KVM/libvirt preferred, explicitly excluding Proxmox per user request).
    -   Design two-layer architecture: trusted infrastructure (Nomad server/client, Tailscale) on host; user-facing apps as Docker containers inside 1 VM per host, scheduled by Nomad.
    -   Investigate Nomad deployment on Debian: server + client on host, client + Docker inside VM, Nomad targeting the VM's Docker daemon for workload scheduling.
    -   Research Nomad job file structure for media applications (Plex, Sonarr, Radarr, Bazarr), including read-only root filesystem (`--read-only`) with explicit writable volume mounts for required data directories (`/config`, `/downloads`, `/media`).
    -   Research reverse proxy deployment (Traefik or Caddy) as a Nomad-scheduled container inside the VM: TLS termination, rate limiting, single ingress point for all media services.
    -   Investigate automated VM creation and management using Terraform and Ansible.
    -   Research optimal configurations for media storage passthrough from host to VM to containers.
    -   Validate resource feasibility: Nomad (~500-750MB) + 1 VM (~512MB-1GB) + media apps (~2-3GB) on 8GB RAM hosts.

5.  **Custom Logic Language Integration**:
    -   If custom scripting is required for orchestration or specific tasks, establish best practices for writing maintainable Go applications within the IaC repository.

**Output**: research.md with all NEEDS CLARIFICATION resolved, providing specific decisions and rationales for the above research areas.

## Phase 1: Design & Contracts

**Prerequisites:** `research.md` complete

1.  **Data Model** (`data-model.md`):
    -   Define entities: Physical Host, Virtual Machine, Nomad Cluster, Nomad Job, Docker Container, Tailscale Node, Reverse Proxy, NAT Network, Media Application Config.
    -   Map relationships: Host -> VM (1:1 for now), VM -> Containers (1:N), Nomad Cluster -> Nomad Jobs (1:N), Reverse Proxy -> Containers (routes).
    -   Define state transitions for hosts, VMs, and containers.

2.  **Contracts** (`contracts/`):
    -   CLI command contracts for engineer workflow: provision, deploy-vpn, deploy-app, harden, status.
    -   Nomad job file schemas for each media application.
    -   Ansible inventory/group_vars structure for application-to-VM assignment.

3.  **Quickstart** (`quickstart.md`):
    -   Step-by-step guide for an engineer to go from bare Debian hosts to a fully running media stack.
    -   Prerequisites, configuration, and single-command deployment workflow.

4.  **Agent context update**:
    -   Run `.specify/scripts/bash/update-agent-context.sh gemini`.

5.  **Post-Design Constitution Re-Check**:
    -   Verify all four principles still pass after detailed design.

**Output**: data-model.md, contracts/*, quickstart.md, agent-specific context file.

## Key rules

- Use absolute paths
- ERROR on gate failures or unresolved clarifications
