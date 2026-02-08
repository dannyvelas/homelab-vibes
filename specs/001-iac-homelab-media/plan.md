# Implementation Plan: IaC Homelab Media

**Branch**: `001-iac-homelab-media` | **Date**: 2026-02-07 | **Spec**: /Users/dannyvelasquez/RemoteGit/MyGithub/homelab-vibe/specs/001-iac-homelab-media/spec.md
**Input**: Feature specification from `/specs/001-iac-homelab-media/spec.md`

## Summary

This project aims to develop an Infrastructure as Code (IaC) solution for a homelab environment, initially on two low-resource laptops, with a clear migration path to on-premise bare-metal servers. The core functionality includes automated server setup (Debian OS), secure remote access via Tailscale VPN, and deployment of a media stack (Plex, Sonarr, Radarr, Bazarr) using Ansible and Terraform. Extreme security, automation, and a user-friendly experience for engineers are paramount.

## Technical Context

**Language/Version**: Go (for custom scripting/logic within IaC if needed, per user preference for Go, Nim, OCaml)  
**Primary Dependencies**: Ansible, Terraform, Tailscale  
**Storage**: Host-level storage for media applications (details to be determined)  
**Testing**: Unit testing for IaC code (e.g., Terratest, Molecule), integration testing for deployed services  
**Target Platform**: Debian servers (bare-metal, 8GB RAM, dual-core Intel i7, 128GB/256GB storage)  
**Project Type**: Single (IaC repository for multiple services)  
**Performance Goals**: Fast, automated deployment; reliable Tailscale VPN access; responsive media applications within hardware constraints.  
**Constraints**: Low-resource initial hardware (laptops), IaC portability for future bare-metal migration, extreme defense-in-depth security.  
**Scale/Scope**: Two laptops initially, transitioning to dedicated bare-metal. Focus on VPN and media stack. Monitoring deferred.

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

-   **I. Easy/Automated End-User Experience**: **PASS** - Plan emphasizes IaC and automation for setup and deployment, aiming for minimal engineer effort.
-   **II. Production Quality**: **PASS** - Focus on security, IaC testing, and future monitoring aligns. Acknowledgement of current hardware constraints, but striving for best possible reliability/performance.
-   **III. Scalability**: **PASS** - IaC with Ansible/Terraform facilitates portability to bare-metal, supporting future growth without significant architectural changes.
-   **IV. Defense-in-Depth Security**: **PASS** - Extreme security is a core requirement, embedded across all layers of the plan.

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
# Option 1: Single project (DEFAULT)
iac/
├── terraform/           # Terraform configurations for server provisioning
├── ansible/             # Ansible playbooks for OS setup and application deployment
├── scripts/             # Custom scripts (e.g., Go) for orchestration or specific tasks
└── docs/                # IaC-specific documentation

tests/
├── integration/         # Integration tests for deployed services
└── unit/                # Unit tests for IaC (Terraform, Ansible)
```

**Structure Decision**: A single project IaC repository structure. Terraform will handle provisioning the initial OS-booted state. Ansible will handle configuration, hardening, Tailscale, and media stack deployment. Custom Go scripts may be used for glue logic or specific automation tasks.

## Complexity Tracking

> **Fill ONLY if Constitution Check has violations that must be justified**

(No violations to justify at this stage.)

## Phase 0: Outline & Research

**Goal**: Establish foundational IaC strategy, secure OS configuration, and optimal media stack deployment methods given constraints.

1.  **IaC Strategy and Integration**:
    -   Research best practices for integrating Terraform (provisioning) and Ansible (configuration) in a homelab context.
    -   Determine how to manage state and secrets securely across both tools.
    -   Explore methods for creating a simple, single-command engineer experience using these tools.

2.  **Debian Server Hardening**:
    -   Research essential security configurations for Debian servers (firewall, SSH hardening, user management, update policies).
    -   Identify minimal Debian installation requirements for this use case.

3.  **Tailscale Deployment and Management**:
    -   Research automated deployment of Tailscale via Ansible/Terraform, including node authentication and subnets.
    -   Investigate secure methods for Tailscale key management within the IaC framework.

4.  **Media Stack Deployment and Isolation**:
    -   Research best practices for deploying Plex, Sonarr, Radarr, Bazarr on Debian.
    -   Evaluate containerization (Docker/Podman) for application isolation and resource management on low-resource hardware, addressing security concerns.
    -   Determine optimal configurations for media storage.

5.  **Custom Logic Language Integration**:
    -   If custom scripting is required for orchestration or specific tasks, establish best practices for writing maintainable Go applications within the IaC repository.

**Output**: research.md with all NEEDS CLARIFICATION resolved, providing specific decisions and rationales for the above research areas.

## Phase 1: Design & Contracts

**Prerequisites:** `research.md` complete

(Content to be filled after Phase 0)

## Key rules

- Use absolute paths
- ERROR on gate failures or unresolved clarifications