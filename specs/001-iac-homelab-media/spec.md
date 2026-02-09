# Feature Specification: IaC Homelab Media

**Feature Branch**: `001-iac-homelab-media`  
**Created**: 2026-02-07  
**Status**: Draft  
**Input**: User description: "create infrastucture as code for a tech startup that will soon grow rapidly. the only thing we have right now is two laptops that we are using as servers. we have physical access to those laptops and can turn them off or on whenever we want. we can choose whatever OS that we want. these servers are connected to a home internet gateway via ethernet. both laptops have 8GB of RAM and dual-core Intel i7 chips. one laptop has 128GB of storage, the other has 256GB of storage. we will soon migrate to having proper bare metal servers that we will run on-premise. the reason we want to do IAAC is that we don't want your work here to go to waste. ideally, once you're done, after we migrate, an engineer from our team will have a very automated, easy, and quick user-experience to get the identical infrastructure/software that is running on the laptops, to start running on the real on-premise servers. of course, it will be difficult, if not impossible, to build ideal production quality software/infrastructure (with good reliability/scalability) on two old and low-resource laptops. however, the goal of this project is to see how close we can get, given these hardware constraints. assuming that some bare-metal servers are remotely accessible and booted with an OS, this repo should allow an engineer to: 1. easily (e.g. minimal effort whether its one button click, one simple command, or one api request) set up those servers so that when anyone from our team is away and wants to access resources in the home LAN, they can use a VPN client on their computer which will it look to them like they are actually in the actual home network, and can access any resource in the home network. 2. easily (e.g. minimal effort whether its one button click, one simple command, or one api request) deploy plex, sonarr, radarr, and bazarr to these servers. After this step, a working instance of plex should be available in my LAN at port 32400. radarr, sonarr, and bazarr should also be available in the home LAN. 3. Also, all of these things should be done with extreme security, so that if there is a vulnerability in any of these apps, or if there is a bad actor using plex, there is minimal chance that they will be able to access the kernel of one of our servers, or connect with local devices in our home network. we want redundant production-level security across all layers. 4. In the future, we will want to also deploy a monitoring/alerting system so that we can monitor the traffic, health, and performance of these applications but for now we can start with the other three requirements"

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

An engineer on the team can execute individual, automated commands or API requests for each media application (Plex, Sonarr, Radarr, and Bazarr) to deploy and configure them onto the designated servers. Upon successful deployment, these applications are accessible within the home LAN.

**Why this priority**: This fulfills a core functional requirement of the project, providing immediate value by setting up key applications. It demonstrates the effectiveness of the IaC approach for application deployment.

**Independent Test**: After running the deployment command, the engineer can access Plex at `http://<server-ip>:32400` from within the home LAN, and verify that Sonarr, Radarr, and Bazarr are also accessible via their respective ports/paths. Basic functionality (e.g., Plex dashboard loads, Sonarr/Radarr UI is accessible) is confirmed.

**Acceptance Scenarios**:

1.  **Given** bare-metal servers are set up (from US1 pre-req), **When** an engineer executes the deployment command for media apps, **Then** Plex, Sonarr, Radarr, and Bazarr are installed, configured, and running on the servers.
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

### User Story 4 - Media Applications Automatically Update to Latest Versions (Priority: P2)

Deployed media applications (Plex, Sonarr, Radarr, Bazarr) automatically detect and apply new version updates without requiring manual intervention from an engineer. The update process preserves application data and configuration. Applications remain available during the update with minimal downtime.

**Why this priority**: Keeping applications up-to-date is critical for security (patching vulnerabilities) and functionality (new features, bug fixes). Automating this removes ongoing maintenance burden from the engineering team, aligning with the "Easy/Automated End-User Experience" principle.

**Independent Test**: Deploy a media application at a known version. Wait for or simulate a new version release. Verify that the application is automatically updated to the new version, that its data and configuration are preserved, and that it remains accessible after the update.

**Acceptance Scenarios**:

1.  **Given** a media application is deployed and running, **When** a new version of that application is released, **Then** the system automatically updates the application to the new version within 24 hours.
2.  **Given** an automatic update is in progress, **When** the update completes, **Then** all application data (libraries, settings, databases) is preserved and the application is fully functional.
3.  **Given** an automatic update fails (e.g., new version crashes on startup), **When** the failure is detected, **Then** the system rolls back to the previous working version automatically.

---

### User Story 5 - Engineer Manages All Configuration from a Single Source of Truth (Priority: P1)

An engineer configures the entire infrastructure and application stack by editing a single, unified configuration file (or a small set of clearly organized files in one location). They never need to duplicate values across multiple tools, copy-paste settings between provisioning and configuration files, or worry about keeping multiple config sources in sync. The system's tooling reads from this single source and generates whatever tool-specific configuration is needed internally.

**Why this priority**: Configuration duplication is a major source of errors and friction in IaC systems. A single source of truth directly supports FR-011 (easy/automated UX) and FR-010 (portability — one config to update when migrating to new hardware).

**Independent Test**: The engineer changes a single value (e.g., a server IP address or a media storage path) in one place, runs the deployment, and verifies that all affected components (provisioning, configuration, application deployment) correctly reflect the change without any additional manual edits.

**Acceptance Scenarios**:

1.  **Given** an engineer needs to configure the infrastructure, **When** they look for where to define settings, **Then** there is one clearly documented location for all configuration values.
2.  **Given** a configuration value (e.g., server IP, storage path, VPN subnet) is used by multiple underlying tools, **When** the engineer updates that value in the single source of truth, **Then** all tools consume the updated value automatically on the next deployment.
3.  **Given** the engineer has never used the system before, **When** they follow the quickstart guide, **Then** they only need to edit configuration in one location before running deployment commands.

---

### Edge Cases

- What happens if a server's network connection is interrupted during deployment or operation?
- How does the system handle power loss or unexpected reboots of the laptops/servers?
- What is the rollback strategy if an automated deployment fails or introduces a critical error?
- How are application configurations (e.g., Plex libraries, Sonarr/Radarr settings) persisted and backed up?
- What happens if one of the old laptops fails completely, considering the resource limitations?
- What happens if an automatic application update introduces a breaking change or is incompatible with the current configuration?
- What happens if two applications have conflicting version requirements (e.g., Sonarr requires a specific Radarr API version)?

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: System MUST automate the setup and configuration of bare-metal servers from an initial OS-booted state.
- **FR-002**: System MUST deploy and configure a VPN server on a designated server.
- **FR-003**: System MUST enable remote users to connect securely to the home LAN via VPN clients.
- **FR-004**: System MUST automate the deployment and configuration of Plex, Sonarr, Radarr, and Bazarr.
- **FR-004a**: System MUST use a container scheduler to manage the placement and lifecycle of application workloads across available infrastructure, enabling automatic scaling as new hosts or VMs are added.
- **FR-005**: Plex MUST be accessible within the home LAN at port 32400 after deployment.
- **FR-006**: Sonarr, Radarr, and Bazarr MUST be accessible within the home LAN after deployment.
- **FR-007**: System MUST implement defense-in-depth security measures to isolate applications and protect the host OS.
- **FR-008**: System MUST apply the principle of least privilege for all deployed applications and services.
- **FR-009**: System MUST prevent lateral movement from compromised applications to other network devices.
- **FR-010**: The infrastructure as code solution MUST be reusable and portable for future migration to different bare-metal servers.
- **FR-011**: The setup and deployment processes MUST provide an automated, easy, and quick user experience for an engineer.
- **FR-012**: Deployed media applications (Plex, Sonarr, Radarr, Bazarr) MUST be automatically updated to the latest version when a new release is available, without manual intervention from the engineer.
- **FR-013**: The system MUST provide a single source of truth for all infrastructure and application configuration. Engineers MUST NOT need to duplicate or synchronize configuration values across multiple tools or files. All configuration MUST be defined in one place and consumed by the underlying provisioning and configuration tools transparently.

### Key Entities *(include if feature involves data)*

- **Server**: A physical or virtual machine with an OS, hosting services.
- **VPN Server**: Software component providing secure remote access.
- **VPN Client**: Software on a remote user's device for connecting to the VPN.
- **Plex**: Media server application.
- **Sonarr**: TV show management application.
- **Radarr**: Movie management application.
- **Bazarr**: Subtitle management application.
- **Home LAN Resource**: Any device or service within the home network.
- **Engineer**: User responsible for setting up and deploying the infrastructure.
- **Remote User**: User accessing LAN resources via VPN.
- **Container Scheduler**: Software that manages which containers run on which hosts/VMs and handles their lifecycle (start, stop, restart, placement).
- **IaC Configuration**: Code defining infrastructure and application deployments.
- **Unified Configuration**: The single source of truth for all infrastructure and application settings, consumed by all underlying tools.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: An engineer can fully provision the base server setup (including VPN) from scratch on a new bare-metal server using a single command/API request within 5 minutes.
- **SC-002**: An engineer can fully deploy the media stack (Plex, Sonarr, Radarr, Bazarr) from scratch using individual commands/API requests for each application within 5 minutes on a provisioned server.
- **SC-003**: Remote users can successfully connect to the VPN and access internal home LAN resources with 100% reliability over a 24-hour test period.
- **SC-004**: A security scan (e.g., Nessus, OpenVAS) reports zero critical or high-severity vulnerabilities directly attributable to misconfigurations or lack of isolation in the deployed infrastructure or applications.
- **SC-005**: All deployed applications run within isolated environments (e.g., containers, VMs) with restricted network access, achieving a minimal attack surface.
- **SC-006**: The IaC codebase can be executed on a separate bare-metal environment (simulating future migration) and successfully reproduce the identical infrastructure setup.
- **SC-007**: All deployed media applications are automatically updated to the latest version within 24 hours of a new release, with zero data loss and automatic rollback on failure.
- **SC-008**: An engineer can change any shared configuration value (e.g., server IP, storage path) in exactly one location, and all affected infrastructure and application components reflect the change on the next deployment run.

