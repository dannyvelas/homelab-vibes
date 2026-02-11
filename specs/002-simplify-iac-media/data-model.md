# Data Model: Simplify IaC Media (Remove Nomad)

**Feature Branch**: `002-simplify-iac-media`
**Created**: 2026-02-09
**Status**: Draft
**Input**: Feature specification from `/specs/002-simplify-iac-media/spec.md`, Research findings from `/specs/002-simplify-iac-media/research.md`

This data model defines the key entities and their attributes for the simplified homelab IaC system. Entities are managed by Terraform and Ansible across a two-layer architecture (host + VM). Compared to 001, Nomad_Cluster, Nomad_Job, and Auto_Update_Job are removed. Container management is handled directly by Ansible's `docker_container` module, and auto-updates are handled by a systemd timer running a shell script inside each VM.

## Key Entities

### Physical_Host
Represents a physical server (laptop or bare-metal) running the host OS.
-   **Name**: String (e.g., `homelab-host-01`) - Unique identifier.
-   **IP_Address**: String (e.g., `192.168.1.10`) - IP address on the home LAN.
-   **OS**: String (e.g., `Debian 12`) - Operating system.
-   **RAM_GB**: Integer (e.g., `8`) - Total RAM in GB.
-   **CPU_Cores**: Integer (e.g., `2`) - Number of CPU cores.
-   **Storage_GB**: Integer (e.g., `128`) - Total usable storage in GB.
-   **SSH_User**: String (e.g., `admin`) - User for SSH access.
-   **SSH_Key_Path**: String - Path to SSH private key for automation.
-   **Status**: Enum (`online`, `offline`, `provisioning`, `decommissioned`) - Operational status.
-   **Host Services**: WireGuard (if designated VPN endpoint), KVM/libvirt hypervisor -- trusted infrastructure running directly on the host OS.

### Virtual_Machine
Represents a KVM/libvirt VM running on a Physical_Host, providing kernel isolation for user-facing workloads.
-   **Name**: String (e.g., `homelab-vm-media-01`) - Unique identifier.
-   **Host_Name**: String - References the `Physical_Host.Name` this VM runs on.
-   **NAT_IP**: String (e.g., `192.168.122.50`) - IP on the private NAT subnet.
-   **NAT_Subnet**: String (e.g., `192.168.122.0/24`) - The private subnet this VM is on.
-   **OS**: String (e.g., `Debian 12`) - Guest operating system.
-   **RAM_MB**: Integer (e.g., `5120`) - Allocated RAM in MB.
-   **vCPUs**: Integer (e.g., `2`) - Allocated virtual CPU cores.
-   **Disk_GB**: Integer (e.g., `40`) - Allocated disk in GB.
-   **Status**: Enum (`running`, `stopped`, `creating`, `error`) - VM state.
-   **VM Services**: Docker engine -- containers managed directly by Ansible.

### NAT_Network
Represents the NAT networking configuration between host and VMs.
-   **Bridge_Name**: String (e.g., `virbr0`) - Linux bridge name on the host.
-   **Subnet**: String (e.g., `192.168.122.0/24`) - Private subnet CIDR.
-   **Gateway_IP**: String (e.g., `192.168.122.1`) - Host's IP on the bridge (gateway for VMs).
-   **DHCP_Range**: String (e.g., `192.168.122.50-192.168.122.200`) - DHCP range for VMs.
-   **Ingress_Rules**: List of Strings - Iptables port forwarding rules (e.g., `host:443 -> vm:443`).
-   **Egress_Rules**: List of Strings - Iptables egress rules (e.g., `deny vm -> home_lan`, `allow vm -> internet`).

### Docker_Container
Represents a Docker container running inside a VM, managed by Ansible's `docker_container` module.
-   **Container_Name**: String (e.g., `plex`, `traefik`) - Unique identifier within the VM.
-   **VM_Name**: String - References the `Virtual_Machine.Name` this container runs in.
-   **Docker_Image**: String (e.g., `linuxserver/plex:latest`) - Container image.
-   **Read_Only_Root**: Boolean (`true`) - Whether container root filesystem is read-only.
-   **Writable_Volumes**: List of Strings (e.g., `["/config", "/downloads", "/media"]`) - Explicitly writable mount paths.
-   **Port_Mappings**: List of Strings (e.g., `["32400:32400/tcp"]`) - Container port mappings.
-   **Environment_Variables**: Map of Strings - Key-value pairs for container environment.
-   **Restart_Policy**: String (`unless-stopped`) - Docker restart policy for automatic recovery on crash or reboot.
-   **Status**: Enum (`running`, `stopped`, `restarting`, `created`) - Container state.
-   **Current_Image_Digest**: String - SHA256 digest of the currently running Docker image.
-   **Run_Script_Path**: String (e.g., `/opt/homelab/run-plex.sh`) - Path to per-container run script inside the VM, templated by Ansible.

### WireGuard_Server
Represents the WireGuard VPN server running on the designated Physical_Host.
-   **Host_Name**: String - References the `Physical_Host.Name` this server runs on.
-   **Interface_Name**: String (e.g., `wg0`) - WireGuard network interface name.
-   **Listen_Port**: Integer (default: `51820`) - UDP port WireGuard listens on.
-   **Server_Private_Key**: String (Sensitive) - Server's WireGuard private key.
-   **Server_Public_Key**: String - Server's WireGuard public key (distributed to clients).
-   **VPN_Subnet**: String (e.g., `10.0.0.0/24`) - Subnet assigned to VPN clients.
-   **Server_VPN_IP**: String (e.g., `10.0.0.1`) - Server's IP within the VPN subnet.
-   **Endpoint**: String (e.g., `home.example.com:51820` or `<public-ip>:51820`) - Public-facing endpoint for remote clients.
-   **Allowed_Routes**: List of Strings (e.g., `["192.168.1.0/24"]`) - Home LAN subnets routed to VPN clients.

### WireGuard_Peer
Represents a VPN client (team member's device) authorized to connect.
-   **Name**: String (e.g., `danny-laptop`) - Human-readable peer identifier.
-   **Public_Key**: String - Peer's WireGuard public key.
-   **Private_Key**: String (Sensitive) - Peer's WireGuard private key (stored in generated client config only).
-   **VPN_IP**: String (e.g., `10.0.0.2`) - Peer's assigned IP within the VPN subnet.
-   **Config_File**: String - Path to the generated client config file.

### Reverse_Proxy
Represents the reverse proxy running inside a VM as an Ansible-managed Docker container.
-   **VM_Name**: String - References the `Virtual_Machine.Name` where it runs.
-   **Container_Name**: String (e.g., `traefik`) - References the `Docker_Container.Container_Name`.
-   **Listen_Port**: Integer (e.g., `443`) - Port the proxy listens on inside the VM.
-   **TLS_Enabled**: Boolean - Whether TLS termination is active.
-   **Routes**: List of Route objects - Mapping of paths/hostnames to backend containers.
-   **Rate_Limiting**: Boolean - Whether rate limiting is enabled.

### Media_Application_Config
Represents application-specific configuration for each media service.
-   **Application_Name**: String (e.g., `Plex`, `Sonarr`, `Radarr`, `Bazarr`) - Application name.
-   **Container_Name**: String - References the `Docker_Container.Container_Name`.
-   **Web_UI_Port**: Integer (e.g., `32400` for Plex, `8989` for Sonarr) - Application web UI port.
-   **Reverse_Proxy_Route**: String (e.g., `/sonarr`) - Path on reverse proxy.
-   **API_Key**: String (Sensitive) - API key for inter-application communication.
-   **Media_Root_Path**: String (e.g., `/media/movies`) - Path to media library inside container.
-   **Config_Path**: String (e.g., `/config`) - Path to persistent config/database inside container.

### Auto_Update_Timer
Represents the systemd timer and update script inside each VM that checks for and applies container image updates.
-   **VM_Name**: String - References the `Virtual_Machine.Name` where it runs.
-   **Timer_Name**: String (`homelab-updater.timer`) - systemd timer unit name.
-   **Service_Name**: String (`homelab-updater.service`) - systemd service unit name.
-   **Script_Path**: String (`/opt/homelab/update-containers.sh`) - Path to the update shell script inside the VM.
-   **Schedule**: String (systemd calendar expression, e.g., `*-*-* 03:00:00`) - When to check for updates.
-   **Monitored_Containers**: List of Strings - Container names to check for updates.
-   **Last_Run**: Timestamp - When the updater last ran (available via `journalctl`).
-   **Persistent**: Boolean (`true`) - Whether missed runs are caught up after VM reboot.

### Unified_Configuration
Represents the single source of truth configuration file (`homelab.yml`) that all tools consume.
-   **File_Path**: String (`homelab.yml`) - Location at repository root.
-   **Cluster_Config**: Object - Cluster name.
-   **Hosts_Config**: Map of `Physical_Host` configurations (IPs, NAT subnets, roles).
-   **VPN_Config**: Object - VPN subnet, endpoint, port, LAN routes.
-   **Storage_Config**: Object - Media, downloads, and config paths on the host.
-   **Apps_Config**: Map of application configurations (enabled, image, port per app).
-   **Auto_Update_Config**: Object - Enabled flag, schedule.
-   **Secrets_Config**: Object - SSH key path, SSH user (secrets referenced by path, not stored inline).

The `iac` CLI reads this entity and generates:
-   Ansible inventory + group_vars -> `.generated/ansible/`
-   Terraform tfvars (per-host) -> `.generated/terraform/<hostname>/`

## Relationships

-   **Physical_Host** 1--1 **Virtual_Machine**: Each host runs one VM for user-facing workloads (current constraint; scales to 1:N with more hardware).
-   **Physical_Host** 1--0..1 **WireGuard_Server**: One host is designated as the WireGuard VPN endpoint.
-   **WireGuard_Server** 1--N **WireGuard_Peer**: The server manages multiple authorized VPN peers (team members' devices).
-   **Physical_Host** 1--1 **NAT_Network**: Each host has one NAT network configuration for its VMs.
-   **Virtual_Machine** 1--N **Docker_Container**: Each VM runs multiple Docker containers (media apps + reverse proxy).
-   **Virtual_Machine** 1--1 **Reverse_Proxy**: Each VM has one reverse proxy container.
-   **Virtual_Machine** 1--1 **Auto_Update_Timer**: Each VM has one auto-update timer checking all containers.
-   **Docker_Container** 1--1 **Media_Application_Config**: Each media app container has one application config.
-   **Reverse_Proxy** 1--N **Docker_Container**: Reverse proxy routes to multiple media app containers.
-   **Auto_Update_Timer** 1--N **Docker_Container**: The auto-updater monitors and updates all media app containers.
-   **Unified_Configuration** 1--1 **System**: Single config file generates all tool-specific configs.

## State Transitions

### Physical_Host
```
offline -> provisioning -> online -> decommissioned
```

### Virtual_Machine
```
(none) -> creating -> running -> stopped -> running
                   -> error -> creating (recreate)
```

### Docker_Container
```
(none) -> created -> running -> stopped -> running (restart_policy: unless-stopped)
                             -> running (auto-update: new image detected, container recreated)
```
