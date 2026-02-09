# Data Model: IaC Homelab Media

**Feature Branch**: `001-iac-homelab-media`
**Created**: 2026-02-07
**Status**: Draft
**Input**: Feature specification from `/specs/001-iac-homelab-media/spec.md`, Research findings from `/specs/001-iac-homelab-media/research.md`

This data model defines the key entities and their attributes for the homelab IaC system. Entities are managed by Terraform, Ansible, and Nomad across a two-layer architecture (host + VM).

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
-   **Host Services**: Nomad (server + client), WireGuard (if designated VPN endpoint), KVM/libvirt hypervisor -- trusted infrastructure running directly on the host OS.

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
-   **VM Services**: Nomad client, Docker engine -- workload execution layer.

### NAT_Network
Represents the NAT networking configuration between host and VMs.
-   **Bridge_Name**: String (e.g., `virbr0`) - Linux bridge name on the host.
-   **Subnet**: String (e.g., `192.168.122.0/24`) - Private subnet CIDR.
-   **Gateway_IP**: String (e.g., `192.168.122.1`) - Host's IP on the bridge (gateway for VMs).
-   **DHCP_Range**: String (e.g., `192.168.122.50-192.168.122.200`) - DHCP range for VMs.
-   **Ingress_Rules**: List of Strings - Iptables port forwarding rules (e.g., `host:443 -> vm:443`).
-   **Egress_Rules**: List of Strings - Iptables egress rules (e.g., `deny vm -> home_lan`, `allow vm -> internet`).

### Nomad_Cluster
Represents the Nomad scheduling cluster spanning all hosts.
-   **Cluster_Name**: String (e.g., `homelab`) - Cluster identifier.
-   **Server_Nodes**: List of Strings - `Physical_Host.Name` references running Nomad server.
-   **Client_Nodes**: List of Strings - `Virtual_Machine.Name` references running Nomad client.
-   **Datacenter**: String (e.g., `dc1`) - Nomad datacenter name.

### Nomad_Job
Represents a Nomad job definition for a scheduled workload.
-   **Job_Name**: String (e.g., `plex`, `traefik`) - Unique job identifier.
-   **Job_File**: String (e.g., `iac/nomad/jobs/plex.hcl`) - Path to HCL job file.
-   **Task_Driver**: String (e.g., `docker`) - Nomad task driver.
-   **Docker_Image**: String (e.g., `plexinc/pms-docker:latest`) - Container image.
-   **Read_Only_Root**: Boolean (`true`) - Whether container root filesystem is read-only.
-   **Writable_Volumes**: List of Strings (e.g., `["/config", "/downloads", "/media"]`) - Explicitly writable mount paths.
-   **Port_Mappings**: List of Strings (e.g., `["32400:32400/tcp"]`) - Container port mappings.
-   **Environment_Variables**: Map of Strings - Key-value pairs for container environment.
-   **Status**: Enum (`running`, `pending`, `dead`) - Job status from Nomad.
-   **Auto_Update_Enabled**: Boolean - Whether this job participates in automatic image updates.
-   **Update_Strategy**: Object - Nomad `update` stanza configuration:
    -   **Health_Check**: String (e.g., `checks`) - Health check method.
    -   **Health_Check_Timeout**: String (e.g., `5m`) - Time to wait for health check to pass.
    -   **Auto_Revert**: Boolean (`true`) - Whether to automatically revert on failed deployment.
-   **Current_Image_Digest**: String - SHA256 digest of the currently running Docker image.
-   **Last_Updated**: Timestamp - When the job was last updated (manually or via auto-update).

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
Represents the reverse proxy running inside a VM as a Nomad-scheduled container.
-   **VM_Name**: String - References the `Virtual_Machine.Name` where it runs.
-   **Nomad_Job_Name**: String (e.g., `traefik`) - References the `Nomad_Job.Job_Name`.
-   **Listen_Port**: Integer (e.g., `443`) - Port the proxy listens on inside the VM.
-   **TLS_Enabled**: Boolean - Whether TLS termination is active.
-   **Routes**: List of Route objects - Mapping of paths/hostnames to backend containers.
-   **Rate_Limiting**: Boolean - Whether rate limiting is enabled.

### Media_Application_Config
Represents application-specific configuration for each media service.
-   **Application_Name**: String (e.g., `Plex`, `Sonarr`, `Radarr`, `Bazarr`) - Application name.
-   **Nomad_Job_Name**: String - References the `Nomad_Job.Job_Name`.
-   **Web_UI_Port**: Integer (e.g., `32400` for Plex, `8989` for Sonarr) - Application web UI port.
-   **Reverse_Proxy_Route**: String (e.g., `/sonarr`) - Path on reverse proxy.
-   **API_Key**: String (Sensitive) - API key for inter-application communication.
-   **Media_Root_Path**: String (e.g., `/media/movies`) - Path to media library inside container.
-   **Config_Path**: String (e.g., `/config`) - Path to persistent config/database inside container.

### Unified_Configuration
Represents the single source of truth configuration file (`homelab.yml`) that all tools consume.
-   **File_Path**: String (`homelab.yml`) - Location at repository root.
-   **Cluster_Config**: Object - Cluster name, datacenter.
-   **Hosts_Config**: Map of `Physical_Host` configurations (IPs, NAT subnets, roles).
-   **VPN_Config**: Object - VPN subnet, endpoint, port, LAN routes.
-   **Storage_Config**: Object - Media, downloads, and config paths on the host.
-   **Apps_Config**: Map of application configurations (enabled, image, port per app).
-   **Auto_Update_Config**: Object - Enabled flag, cron schedule, auto-revert setting, health check timeout.
-   **Secrets_Config**: Object - SSH key path, SSH user (secrets referenced by path, not stored inline).

The `iac` CLI reads this entity and generates:
-   Ansible inventory + group_vars → `.generated/ansible/`
-   Terraform tfvars → `.generated/terraform/`
-   Nomad HCL job files → `.generated/nomad/`

### Auto_Update_Job
Represents the periodic Nomad batch job that checks for and applies application updates.
-   **Job_Name**: String (`auto-updater`) - Nomad job identifier.
-   **Schedule**: String (cron expression, e.g., `0 3 * * *`) - How often to check for updates.
-   **Monitored_Apps**: List of Strings - `Nomad_Job.Job_Name` references for apps to check.
-   **Last_Run**: Timestamp - When the updater last ran.
-   **Last_Result**: Enum (`no_updates`, `updated`, `update_failed_reverted`) - Outcome of the last run.

## Relationships

-   **Physical_Host** 1--1 **Virtual_Machine**: Each host runs one VM for user-facing workloads (current constraint; scales to 1:N with more hardware).
-   **Physical_Host** 1--0..1 **WireGuard_Server**: One host is designated as the WireGuard VPN endpoint.
-   **WireGuard_Server** 1--N **WireGuard_Peer**: The server manages multiple authorized VPN peers (team members' devices).
-   **Physical_Host** 1--1 **NAT_Network**: Each host has one NAT network configuration for its VMs.
-   **Nomad_Cluster** 1--N **Physical_Host**: Cluster spans all hosts (Nomad servers).
-   **Nomad_Cluster** 1--N **Virtual_Machine**: Cluster spans all VMs (Nomad clients).
-   **Virtual_Machine** 1--N **Nomad_Job**: Each VM runs multiple Nomad jobs (media apps + reverse proxy).
-   **Virtual_Machine** 1--1 **Reverse_Proxy**: Each VM has one reverse proxy.
-   **Nomad_Job** 1--1 **Media_Application_Config**: Each media app job has one application config.
-   **Reverse_Proxy** 1--N **Nomad_Job**: Reverse proxy routes to multiple media app containers.
-   **Unified_Configuration** 1--1 **System**: Single config file generates all tool-specific configs.
-   **Auto_Update_Job** 1--N **Nomad_Job**: The auto-updater monitors and triggers updates for media app jobs.

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

### Nomad_Job
```
(none) -> pending -> running -> dead
                  -> running (self-healing restart)
                  -> running (auto-update: new image detected)
                     -> health_check_pass -> running (new version)
                     -> health_check_fail -> running (auto-reverted to previous version)
```
