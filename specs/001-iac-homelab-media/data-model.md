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
-   **Host Services**: Nomad (server + client), Tailscale, KVM/libvirt hypervisor -- trusted infrastructure running directly on the host OS.

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

### Tailscale_Node
Represents a host registered as a node in the Tailnet.
-   **Host_Name**: String - References the `Physical_Host.Name` it's associated with.
-   **Tailscale_IP**: String (e.g., `100.x.y.z`) - Tailscale IP address.
-   **Advertised_Routes**: List of Strings (e.g., `["192.168.1.0/24"]`) - Subnets advertised by this node.
-   **Auth_Key_ID**: String - Identifier for the pre-authenticated key used for registration.
-   **Roles**: List of Strings (e.g., `subnet-router`, `exit-node`) - Tailscale-specific roles.

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

## Relationships

-   **Physical_Host** 1--1 **Virtual_Machine**: Each host runs one VM for user-facing workloads (current constraint; scales to 1:N with more hardware).
-   **Physical_Host** 1--1 **Tailscale_Node**: Each host registers as one Tailscale node.
-   **Physical_Host** 1--1 **NAT_Network**: Each host has one NAT network configuration for its VMs.
-   **Nomad_Cluster** 1--N **Physical_Host**: Cluster spans all hosts (Nomad servers).
-   **Nomad_Cluster** 1--N **Virtual_Machine**: Cluster spans all VMs (Nomad clients).
-   **Virtual_Machine** 1--N **Nomad_Job**: Each VM runs multiple Nomad jobs (media apps + reverse proxy).
-   **Virtual_Machine** 1--1 **Reverse_Proxy**: Each VM has one reverse proxy.
-   **Nomad_Job** 1--1 **Media_Application_Config**: Each media app job has one application config.
-   **Reverse_Proxy** 1--N **Nomad_Job**: Reverse proxy routes to multiple media app containers.

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
```
