# Research Findings: Simplify IaC Media (Remove Nomad)

**Feature Branch**: `002-simplify-iac-media`
**Created**: 2026-02-09
**Status**: Completed
**Input**: Phase 0 Research tasks from plan.md, findings from 001-iac-homelab-media

## Carried Forward from 001

The following decisions from `specs/001-iac-homelab-media/research.md` remain unchanged:

- **IaC Strategy**: Terraform (VM lifecycle via libvirt provider) + Ansible (configuration management via SSH). Same two-layer approach.
- **Debian Server Hardening**: UFW, SSH key-only, auto-updates, least privilege. Same Ansible roles.
- **Network Segmentation**: NAT with private subnets, egress blocking, iptables. Unchanged.
- **VPN**: WireGuard on host, kernel module, Ansible-deployed. Unchanged.
- **VM Isolation**: One KVM/libvirt VM per host for user-facing workloads. Unchanged.
- **Reverse Proxy**: Traefik inside VM as single ingress point. Unchanged (but deployed via Ansible `docker_container` instead of Nomad).
- **Read-only Containers**: `--read-only` flag with explicit writable volume mounts. Unchanged.
- **Unified Configuration**: Single `homelab.yml`, `iac` CLI generates tool-specific configs. Unchanged.
- **Custom Scripting**: Go for the `iac` CLI. Unchanged.

## New Research: Replacing Nomad

### 1. Container Management Without a Scheduler

- **Decision**: Use Ansible's `community.docker.docker_container` module for initial deployment and configuration. Use Docker's `restart_policy: unless-stopped` for automatic restart on crash or reboot. No scheduler needed.

- **Rationale**: For 2-3 hosts running four known applications with predetermined placement, a scheduler adds complexity without value. The `docker_container` module is idempotent — re-running the playbook converges to the desired state. Docker's native restart policies handle the most common failure mode (container crash or VM reboot) without any external process. The `unless-stopped` policy restarts containers after the Docker daemon starts, unless the container was explicitly stopped by the operator.

- **Alternatives Considered**: Docker Compose (adds a dependency and file format on top of Docker, doesn't integrate naturally with Ansible's variable system); Podman with systemd-generated units (viable but adds complexity switching from Docker, and the Ansible collection is less mature); keeping Nomad (rejected per feature scope — unnecessary resource and complexity overhead).

### 2. Auto-Updates Without Nomad

- **Decision**: A shell script run by a systemd timer inside each VM. The script iterates over configured containers, pulls the latest image, compares the digest to the running container's image, and recreates the container if the image has changed. Per-container run scripts (templated by Ansible) define the full `docker run` invocation so the update script only needs to call them.

- **Rationale**: systemd timers are preferred over cron because: (1) `Persistent=true` catches up on missed runs after VM reboot, (2) built-in single-instance guarantee prevents overlapping runs, (3) all output is captured by journald with no log management needed, (4) no additional packages required (systemd is already the init system on Debian 12). The per-container run scripts ensure the recreated container has exactly the same configuration as the Ansible-managed definition — no risk of drift between the update script and the Ansible playbook.

- **Alternatives Considered**: Watchtower (black box, adds another running container, not customizable for deferred rollback); running `ansible-playbook` on a timer inside the VM (requires installing Ansible in every VM — heavyweight); running Ansible from the engineer's workstation on a schedule (requires always-on workstation); using `community.docker.docker_container` with `pull: always` (only works when Ansible is invoked, not for autonomous periodic updates).

### 3. Rollback Strategy

- **Decision**: Deferred to a future feature. The current implementation does not implement automatic rollback. If an update causes issues, the engineer can manually run `iac deploy app --name <app>` to redeploy the previous known-good configuration.

- **Rationale**: The user explicitly deferred rollback and self-healing to a future feature when a scheduler may be reintroduced. Implementing rollback without a scheduler adds significant complexity (image tagging, health checks, restore logic) for marginal benefit in a homelab context where downtime is acceptable and the engineer can intervene manually.

### 4. Ansible Deployment Pattern for Containers

- **Decision**: Use a two-step pattern: `community.docker.docker_image_pull` to pull the image, then `community.docker.docker_container` with conditional `recreate` based on whether the pull detected a change. This avoids edge cases with the `pull: always` parameter on `docker_container`.

- **Rationale**: The `docker_image_pull` module returns `changed: true` only when a new image was actually downloaded. Using this to conditionally set `recreate: true` on the `docker_container` task ensures containers are only recreated when the image actually changed, not on every Ansible run.

### 5. Config Schema Changes

- **Decision**: Remove `cluster.datacenter` from `homelab.yml` (this was a Nomad concept). Remove Nomad-related references from the `iac` CLI. The rest of the schema remains identical.

- **Rationale**: The datacenter concept was used for Nomad job placement. Without Nomad, it serves no purpose. All other configuration (hosts, VPN, storage, apps, auto_update, secrets) is consumed by Ansible and Terraform directly.

## Key Decisions Summary

- **Container Management**: Ansible `docker_container` module + Docker `restart_policy: unless-stopped`
- **Auto-Updates**: Shell script + systemd timer inside each VM; per-container run scripts templated by Ansible
- **Rollback**: Deferred to future feature
- **Config Schema**: Remove `cluster.datacenter`, keep everything else
- **Everything else**: Unchanged from 001
