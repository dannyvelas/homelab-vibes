# IaC System Contracts

**Feature Branch**: `001-iac-homelab-media`  
**Created**: 2026-02-07  
**Status**: Conceptual API Outline  
**Input**: Functional Requirements from `/specs/001-iac-homelab-media/spec.md`, Research findings from `/specs/001-iac-homelab-media/research.md`

This document outlines the conceptual API (Application Programming Interface) for interacting with the IaC system. Given the project's focus on Infrastructure as Code, the primary "API" for engineers will be command-line interface (CLI) interactions with Terraform and Ansible, potentially wrapped in custom Go scripts for a simplified user experience. No formal HTTP-based OpenAPI or GraphQL schema is generated at this stage, as the interaction model is CLI-centric.

## IaC Operations (Conceptual Endpoints / CLI Commands)

### 1. Server Provisioning and Initial OS Setup

**Purpose**: Automate the initial setup and configuration of bare-metal servers from an OS-booted state (Debian). This includes basic OS configuration and user management.

-   **Functional Requirement**: FR-001 (Automate server setup)
-   **Conceptual Method**: `POST /servers/{server_name}/provision`
-   **CLI Command Equivalent**: `iac provision server --name <server_name> --ip <ip_address> --os debian`
-   **Inputs**:
    -   `server_name`: String (Path parameter/CLI argument) - Unique identifier for the server.
    -   `ip_address`: String (e.g., `192.168.1.10`) - Target IP address of the server.
    -   `ssh_user`: String (e.g., `admin`) - Initial SSH user.
    -   `ssh_key_path`: String - Path to SSH private key for initial connection.
-   **Outputs**:
    -   Confirmation of successful provisioning.
    -   Updated inventory for configuration management.

### 2. VPN Server Deployment and Configuration

**Purpose**: Deploy and configure the Tailscale VPN server on a designated server.

-   **Functional Requirement**: FR-002 (Deploy VPN server)
-   **Conceptual Method**: `POST /servers/{server_name}/deploy-vpn`
-   **CLI Command Equivalent**: `iac deploy vpn --server <server_name> --auth-key <tailscale_auth_key>`
-   **Inputs**:
    -   `server_name`: String (Path parameter/CLI argument) - Target server for VPN deployment.
    -   `tailscale_auth_key`: String (Sensitive) - Pre-authenticated key for Tailscale node registration.
    -   `advertised_routes`: List of Strings (Optional) - Subnets to advertise (e.g., `192.168.1.0/24`).
-   **Outputs**:
    -   Confirmation of Tailscale installation and node registration.
    -   Tailscale IP address of the server.

### 3. Media Application Deployment (Plex, Sonarr, Radarr, Bazarr)

**Purpose**: Automate the deployment and configuration of individual media applications using Docker Compose.

-   **Functional Requirement**: FR-004 (Automate media app deployment)
-   **Conceptual Method**: `POST /servers/{server_name}/deploy-app/{app_name}`
-   **CLI Command Equivalent (for each app)**:
    -   `iac deploy plex --server <server_name> --media-path <path>`
    -   `iac deploy sonarr --server <server_name> --media-path <path> --api-key <key>`
    -   `iac deploy radarr --server <server_name> --media-path <path> --api-key <key>`
    -   `iac deploy bazarr --server <server_name> --media-path <path> --api-key <key>`
-   **Inputs**:
    -   `server_name`: String (Path parameter/CLI argument) - Target server for app deployment.
    -   `app_name`: String (Path parameter/CLI argument) - Specific application (`plex`, `sonarr`, `radarr`, `bazarr`).
    -   `media_path`: String (Optional) - Host path for media storage (e.g., `/mnt/media/movies`).
    -   `api_key`: String (Sensitive, for Sonarr/Radarr/Bazarr) - API key for application communication.
    -   `port`: Integer (Optional) - Custom port for web UI.
-   **Outputs**:
    -   Confirmation of Docker container deployment and configuration.
    -   Accessibility details (e.g., `http://<server_ip>:<port>`).

### 4. System Security Hardening

**Purpose**: Apply defense-in-depth security measures to the server infrastructure.

-   **Functional Requirement**: FR-007, FR-008, FR-009 (Implement defense-in-depth security, least privilege, prevent lateral movement)
-   **Conceptual Method**: `POST /servers/{server_name}/harden-security` (often integrated into provisioning/deployment)
-   **CLI Command Equivalent**: `iac harden security --server <server_name>` (can be implicit in `provision` or `deploy` commands)
-   **Inputs**:
    -   `server_name`: String (Path parameter/CLI argument) - Target server.
-   **Outputs**:
    -   Confirmation of security measures applied (e.g., firewall rules, SSH hardening).

## Interaction Model

The primary interaction model is a CLI-driven workflow, where an engineer executes commands to trigger IaC operations. These commands will internally invoke Terraform and Ansible, providing a simplified abstraction. Custom Go scripts may wrap these calls for enhanced user experience and error handling.
