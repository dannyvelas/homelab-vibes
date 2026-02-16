# Feature Specification: Simplify IaC Media (Remove Nomad)

**Feature Branch**: `002-simplify-iac-media`
**Created**: 2026-02-09
**Status**: Draft
**Input**: User description: "Improve on 001-iac-homelab-media by making it simpler and using less resources. Retain all the same functionality except remove Nomad. The current implementation chose to use Nomad when in reality it wasn't necessary for the use case."

## Motivation

The previous implementation (001-iac-homelab-media) introduced a container scheduler to manage application workloads. While a scheduler provides automatic workload placement and multi-node scheduling, these capabilities are unnecessary for a homelab with 2-3 servers running four known applications with predetermined placement. The scheduler consumes significant RAM per host and adds operational complexity (cluster formation, consensus protocols, join configuration) that outweighs its benefits at this scale.

This feature removes the container scheduler and replaces it with direct container management, where the provisioning tool manages containers on each VM without an intermediary. All other functionality — VPN, defense-in-depth security, single config file, portability — is retained unchanged. Auto-updates are deferred to a future implementation when a container scheduler (Nomad or k3s) is introduced.

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Remote Engineer Securely Accesses Home LAN Resources via VPN (Priority: P1)

An engineer from the team, working remotely, needs to access resources located within the home LAN. They use a pre-configured VPN client on their computer to establish a secure connection, making it appear as if they are physically present in the home network. This allows them to seamlessly access any internal resource.

**Why this priority**: Essential for remote work and team collaboration, providing secure access to internal tools and resources from outside the home network. It establishes foundational networking required for other services.

**Independent Test**: The engineer can connect to the VPN from an external network. After connecting, they can successfully ping a known internal IP address (e.g., the server hosting Plex) and access a simple web service running on a non-public port within the home LAN.

**Acceptance Scenarios**:

1.  **Given** an engineer is outside the home LAN with a VPN client, **When** they connect to the VPN, **Then** their device is assigned an IP address within the home LAN subnet and they can access local resources.
2.  **Given** a VPN connection is established, **When** the engineer attempts to access a local resource (e.g., SSH into a server, access a web UI), **Then** the access is successful and indistinguishable from being physically on the home network.

---

### User Story 2 - Engineer Deploys Media Stack Applications (Plex, Sonarr, Radarr, Bazarr) with Individual Commands (Priority: P1)

An engineer on the team can execute individual, automated commands for each media application (Plex, Sonarr, Radarr, and Bazarr) to deploy and configure them onto the designated servers. Upon successful deployment, these applications are accessible within the home LAN. Application containers are managed directly by the provisioning tool without an intermediary scheduler, reducing resource overhead and operational complexity.

**Why this priority**: This fulfills a core functional requirement of the project, providing immediate value by setting up key applications. It demonstrates the effectiveness of the IaC approach for application deployment.

**Independent Test**: After running the deployment command, the engineer can access Plex at `http://<server-ip>:32400` from within the home LAN, and verify that Sonarr, Radarr, and Bazarr are also accessible via their respective ports/paths. Basic functionality (e.g., Plex dashboard loads, Sonarr/Radarr UI is accessible) is confirmed.

**Acceptance Scenarios**:

1.  **Given** bare-metal servers are provisioned, **When** an engineer executes the deployment command for media apps, **Then** Plex, Sonarr, Radarr, and Bazarr are installed, configured, and running on the servers.
2.  **Given** the media apps are deployed, **When** a user accesses Plex at `http://<server-ip>:32400` from the home LAN, **Then** the Plex web interface loads successfully.
3.  **Given** the media apps are deployed, **When** a user attempts to access Sonarr, Radarr, and Bazarr from the home LAN, **Then** their respective web interfaces load successfully.

---

### User Story 3 - Robust Security Measures Protect Infrastructure from Application Vulnerabilities and External Threats (Priority: P2)

The entire infrastructure, including the host OS, network, and deployed applications, is configured with "defense-in-depth" security measures. This ensures that even if an application (like Plex) is compromised or a malicious actor gains access, the risk of escalation to the host kernel or lateral movement within the home network is minimized.

**Why this priority**: Security is a non-negotiable principle for this project, vital for protecting sensitive data and the integrity of the home network. While difficult on limited hardware, it's a core requirement to strive for.

**Independent Test**: A security audit (manual or automated) can be performed to verify network segmentation, least privilege configurations for applications, firewall rules, and isolation mechanisms are in place according to best practices, demonstrating a hardened attack surface.

**Acceptance Scenarios**:

1.  **Given** applications are running on the servers, **When** a potential vulnerability in an application is exploited, **Then** the attack surface is minimized, preventing direct access to the host kernel or other devices on the home network.
2.  **Given** a malicious actor attempts unauthorized access, **When** security controls are in place (e.g., firewalls, access controls, containerization), **Then** the system actively resists intrusion and limits potential damage.
3.  **Given** application processes are running, **When** examining their permissions, **Then** they operate with the principle of least privilege, having only the necessary access to system resources.

---

### User Story 4 - Media Applications Automatically Update to Latest Versions (Priority: P2) — DEFERRED

> **Status**: Deferred to a future implementation. Auto-updates will be re-introduced when migrating to a container scheduler (Nomad or k3s) that provides built-in rolling update and rollback capabilities. Until then, updates are performed manually by re-running `iac deploy app`.

---

### User Story 5 - Engineer Manages All Configuration from a Single Source of Truth (Priority: P1)

An engineer configures the entire infrastructure and application stack by editing a single, unified configuration file. They never need to duplicate values across multiple tools, copy-paste settings between provisioning and configuration files, or worry about keeping multiple config sources in sync. The system's tooling reads from this single source and generates whatever tool-specific configuration is needed internally.

**Why this priority**: Configuration duplication is a major source of errors and friction in IaC systems. A single source of truth directly supports easy/automated UX and portability — one config to update when migrating to new hardware.

**Independent Test**: The engineer changes a single value (e.g., a server IP address or a media storage path) in one place, runs the deployment, and verifies that all affected components (provisioning, configuration, application deployment) correctly reflect the change without any additional manual edits.

**Acceptance Scenarios**:

1.  **Given** an engineer needs to configure the infrastructure, **When** they look for where to define settings, **Then** there is one clearly documented location for all configuration values.
2.  **Given** a configuration value (e.g., server IP, storage path, VPN subnet) is used by multiple underlying tools, **When** the engineer updates that value in the single source of truth, **Then** all tools consume the updated value automatically on the next deployment.
3.  **Given** the engineer has never used the system before, **When** they follow the quickstart guide, **Then** they only need to edit configuration in one location before running deployment commands.

---

### User Story 6 - Simplified Architecture Uses Fewer Resources Than Previous Implementation (Priority: P1)

The system achieves the same functionality as the previous implementation while consuming significantly less RAM and CPU by eliminating the container scheduler. This frees resources for the actual media applications and leaves headroom for future services.

**Why this priority**: The target hardware has only 8 GB of RAM per host. Every megabyte consumed by infrastructure overhead is a megabyte unavailable to the applications that provide actual user value. Reducing overhead directly improves application performance and enables future expansion.

**Independent Test**: After full deployment, measure the total RAM consumed by infrastructure components (excluding application containers). Compare against the previous implementation's infrastructure overhead. The new implementation uses measurably less RAM.

**Acceptance Scenarios**:

1.  **Given** the full stack is deployed on a host, **When** measuring infrastructure overhead (everything except application containers), **Then** the overhead is at least 500 MB less per host than the previous implementation.
2.  **Given** the system is running, **When** an engineer reviews the running processes, **Then** there is no container scheduler process consuming resources.
3.  **Given** the system is running on an 8 GB host, **When** all four media applications are deployed, **Then** at least 2 GB of RAM remains available for operating system cache and future services.

---

### Edge Cases

- What happens if a server's network connection is interrupted during deployment or operation?
- How does the system handle power loss or unexpected reboots of the laptops/servers?
- What is the rollback strategy if an automated deployment fails or introduces a critical error?
- How are application configurations (e.g., Plex libraries, Sonarr/Radarr settings) persisted and backed up?
- What happens if one of the old laptops fails completely, considering the resource limitations?
- What happens if an automatic application update introduces a breaking change or is incompatible with the current configuration?
- What happens if a container crashes or exits unexpectedly — how is it restarted without a scheduler?
- What happens when deploying to a host that already has a previous version of the application running?

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: System MUST automate the setup and configuration of bare-metal servers from an initial OS-booted state.
- **FR-002**: System MUST deploy and configure a VPN server on a designated server.
- **FR-003**: System MUST enable remote users to connect securely to the home LAN via VPN clients.
- **FR-004**: System MUST automate the deployment and configuration of Plex, Sonarr, Radarr, and Bazarr as containers within isolated virtual machines.
- **FR-005**: Plex MUST be accessible within the home LAN at port 32400 after deployment.
- **FR-006**: Sonarr, Radarr, and Bazarr MUST be accessible within the home LAN after deployment.
- **FR-007**: System MUST implement defense-in-depth security measures to isolate applications and protect the host OS, including: NAT networking, egress blocking, VM kernel isolation, reverse proxy ingress, read-only container filesystems, and OS hardening.
- **FR-008**: System MUST apply the principle of least privilege for all deployed applications and services.
- **FR-009**: System MUST prevent lateral movement from compromised applications to other network devices.
- **FR-010**: The infrastructure as code solution MUST be reusable and portable for future migration to different bare-metal servers.
- **FR-011**: The setup and deployment processes MUST provide an automated, easy, and quick user experience for an engineer.
- **FR-012**: ~~Deployed media applications MUST be automatically updated to the latest version when a new release is available, without manual intervention from the engineer. Failed updates MUST be automatically rolled back.~~ *DEFERRED — will be re-introduced with a container scheduler (Nomad or k3s).*
- **FR-013**: The system MUST provide a single source of truth for all infrastructure and application configuration. Engineers MUST NOT need to duplicate or synchronize configuration values across multiple tools or files.
- **FR-014**: Application containers MUST be managed directly by the provisioning tool without an intermediary container scheduler, to minimize resource overhead and operational complexity.
- **FR-015**: Application containers MUST automatically restart on failure or host reboot without requiring a scheduler.

### Key Entities

- **Server**: A physical machine with an OS, hosting infrastructure services and virtual machines.
- **Virtual Machine**: An isolated compute environment on each server, running containers for applications.
- **VPN Server**: Software component providing secure remote access to the home LAN.
- **VPN Client**: Software on a remote user's device for connecting to the VPN.
- **Reverse Proxy**: Single ingress point inside each VM, handling TLS termination and routing to application containers.
- **Plex**: Media server application.
- **Sonarr**: TV show management application.
- **Radarr**: Movie management application.
- **Bazarr**: Subtitle management application.
- **Home LAN Resource**: Any device or service within the home network.
- **Engineer**: User responsible for setting up and deploying the infrastructure.
- **Remote User**: User accessing LAN resources via VPN.
- **Unified Configuration**: The single source of truth for all infrastructure and application settings, consumed by all underlying tools.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: An engineer can fully provision the base server setup (including VPN) from scratch on a new bare-metal server using a single command within 5 minutes.
- **SC-002**: An engineer can fully deploy the media stack (Plex, Sonarr, Radarr, Bazarr) from scratch using individual commands for each application within 5 minutes on a provisioned server.
- **SC-003**: Remote users can successfully connect to the VPN and access internal home LAN resources with 100% reliability over a 24-hour test period.
- **SC-004**: A security audit reports zero critical or high-severity vulnerabilities directly attributable to misconfigurations or lack of isolation in the deployed infrastructure or applications.
- **SC-005**: All deployed applications run within isolated environments (containers inside VMs) with restricted network access, achieving a minimal attack surface.
- **SC-006**: The IaC codebase can be executed on a separate bare-metal environment (simulating future migration) and successfully reproduce the identical infrastructure setup.
- **SC-007**: ~~All deployed media applications are automatically updated to the latest version within 24 hours of a new release, with zero data loss and automatic rollback on failure.~~ *DEFERRED.*
- **SC-008**: An engineer can change any shared configuration value (e.g., server IP, storage path) in exactly one location, and all affected infrastructure and application components reflect the change on the next deployment run.
- **SC-009**: Infrastructure overhead (excluding application containers) consumes at least 500 MB less RAM per host compared to the previous scheduler-based implementation.
