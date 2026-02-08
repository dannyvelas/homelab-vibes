# Data Model: IaC Homelab Media

**Feature Branch**: `001-iac-homelab-media`  
**Created**: 2026-02-07  
**Status**: Draft  
**Input**: Feature specification from `/specs/001-iac-homelab-media/spec.md`, Research findings from `/specs/001-iac-homelab-media/research.md`

This data model defines the key entities and their attributes relevant to the Infrastructure as Code (IaC) deployment for the homelab media server environment. These entities will be managed or defined by Terraform, Ansible, and Docker Compose.

## Key Entities

### Server
Represents a physical or virtual machine in the homelab environment.
-   **Name**: String (e.g., `homelab-server-01`) - Unique identifier for the server.
-   **IP_Address**: String (e.g., `192.168.1.10`) - Primary IP address within the home LAN.
-   **OS**: String (e.g., `Debian 12`) - Operating system installed on the server.
-   **RAM_GB**: Integer (e.g., `8`) - Total RAM in GB.
-   **CPU_Cores**: Integer (e.g., `2`) - Number of CPU cores.
-   **Storage_GB**: Integer (e.g., `128`) - Total usable storage in GB.
-   **SSH_User**: String (e.g., `admin`) - User for SSH access.
-   **SSH_Key_Path**: String - Path to the SSH private key used for automation.
-   **Role**: String (e.g., `media-host`, `vpn-gateway`) - Primary function of the server.
-   **Status**: Enum (`online`, `offline`, `provisioning`, `decommissioned`) - Current operational status.

### Tailscale_Node
Represents a server registered as a node in the Tailnet.
-   **Server_Name**: String - References the `Server.Name` it's associated with.
-   **Tailscale_IP**: String (e.g., `100.x.y.z`) - Tailscale IP address.
-   **Advertised_Routes**: List of Strings (e.g., `["192.168.1.0/24"]`) - Subnets advertised by this node.
-   **Auth_Key_ID**: String (UUID) - Identifier for the pre-authenticated key used for registration.
-   **Roles**: List of Strings (e.g., `subnet-router`, `exit-node`) - Tailscale-specific roles.

### Docker_Container
Represents a deployed application running as a Docker container.
-   **Name**: String (e.g., `plex`) - Unique name for the container.
-   **Server_Name**: String - References the `Server.Name` where it's deployed.
-   **Image**: String (e.g., `plexinc/pms-docker:latest`) - Docker image used.
-   **Version**: String (e.g., `1.32.8.7479`) - Version of the application/image.
-   **Port_Bindings**: List of Strings (e.g., `["32400:32400/tcp"]`) - Host:Container port mappings.
-   **Volume_Mappings**: List of Strings (e.g., `["/mnt/media:/data/media"]`) - Host:Container volume mappings for persistent data.
-   **Environment_Variables**: Map of Strings - Key-value pairs for container environment.
-   **Status**: Enum (`running`, `stopped`, `restarting`, `error`) - Current container state.

### Media_Application_Config
Represents specific configuration for each media application instance.
-   **Application_Name**: String (e.g., `Plex`, `Sonarr`, `Radarr`, `Bazarr`) - Name of the application.
-   **Container_Name**: String - References the `Docker_Container.Name` it's associated with.
-   **Web_UI_Port**: Integer (e.g., `8989` for Sonarr) - Port for the application's web interface.
-   **Base_URL**: String (e.g., `/sonarr`) - Base URL path if behind a reverse proxy.
-   **API_Key**: String (Sensitive) - API key for inter-application communication (e.g., Sonarr to Radarr).
-   **Media_Root_Path**: String (e.g., `/mnt/media/movies`) - Path to media library on the host.
-   **User_Access_Control**: Map of Strings - Defines access rules or authentication methods.

## Relationships

-   **Server** 1--N **Tailscale_Node**: A Server can host zero or one Tailscale Node (if not, it's just a regular server).
-   **Server** 1--N **Docker_Container**: A Server can host multiple Docker Containers.
-   **Docker_Container** 1--1 **Media_Application_Config**: Each Docker Container running a media application has one associated configuration.
-   **IaC_Configuration** 1--N **Server**: IaC configuration defines and manages multiple Servers.
-   **IaC_Configuration** 1--N **Tailscale_Node**: IaC configuration defines and manages Tailscale nodes.
-   **IaC_Configuration** 1--N **Docker_Container**: IaC configuration defines and manages Docker containers.
-   **IaC_Configuration** 1--N **Media_Application_Config**: IaC configuration defines and manages media application configurations.
