# Implementation Plan: Simplify IaC Media (Remove Nomad)

**Branch**: `002-simplify-iac-media` | **Date**: 2026-02-09 | **Spec**: /Users/dannyvelasquez/RemoteGit/MyGithub/homelab-vibe/specs/002-simplify-iac-media/spec.md
**Input**: Feature specification from `/specs/002-simplify-iac-media/spec.md`

## Summary

This feature removes the Nomad container scheduler from the homelab IaC system and replaces it with direct container management via Ansible's `community.docker.docker_container` module, plus a systemd timer-based auto-update mechanism. The two-layer architecture (trusted infra on host, apps in VMs via NAT) and all security layers are preserved. The result is a simpler system that saves ~500-750MB RAM per host by eliminating the Nomad server and client processes. Rollback and self-healing are explicitly deferred to a future feature when a scheduler may be reintroduced.

## Technical Context

**Language/Version**: Go (for `iac` CLI)
**Primary Dependencies**: Ansible (with `community.docker` collection), Terraform (with `dmacvicar/libvirt` provider), WireGuard, KVM/libvirt, Docker
**Storage**: Host-level storage with passthrough to VM and containers (media libraries, config, downloads)
**Testing**: `go build` for CLI compilation; manual integration testing on deployed infrastructure
**Target Platform**: Debian 12 servers (bare-metal, 8GB RAM, dual-core Intel i7)
**Project Type**: Single (IaC repository)
**Performance Goals**: Fast automated deployment; at least 500MB less RAM overhead per host compared to 001
**Constraints**: 8GB RAM per host; no container scheduler; auto-updates must work autonomously
**Scale/Scope**: 1-3 hosts; four media apps (Plex, Sonarr, Radarr, Bazarr) + reverse proxy per VM
**Configuration**: Single `homelab.yml` at repo root; `iac` CLI generates Ansible inventory and Terraform tfvars from it

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

-   **I. Easy/Automated End-User Experience**: **PASS** -- Simpler than 001. Same single-command experience via `iac` CLI. Fewer moving parts (no Nomad cluster formation, no Nomad join). Single `homelab.yml` config file. Auto-updates via systemd timer require zero operator interaction.
-   **II. Production Quality**: **PASS** -- Docker `restart_policy: unless-stopped` handles crash/reboot recovery. systemd timer with `Persistent=true` catches up on missed update runs. Rollback deferred per user decision (acceptable for homelab context).
-   **III. Scalability**: **PASS** -- Per-host Terraform state and Ansible inventory support 1-N hosts. Adding a host = adding an entry to `homelab.yml` and running `iac provision host`. No scheduler needed at this scale; scheduler can be reintroduced when the cluster grows.
-   **IV. Defense-in-Depth Security**: **PASS** -- All six security layers preserved unchanged: OS hardening, NAT networking, egress blocking, VM kernel isolation, reverse proxy, read-only containers. Removing Nomad actually reduces attack surface (fewer network-exposed services).

## Project Structure

### Documentation (this feature)

```text
specs/002-simplify-iac-media/
├── plan.md              # This file
├── research.md          # Phase 0 output
├── data-model.md        # Phase 1 output
├── quickstart.md        # Phase 1 output
├── contracts/           # Phase 1 output
│   └── README.md        # CLI command contracts
└── tasks.md             # Phase 2 output (/speckit.tasks command)
```

### Source Code (repository root)

```text
homelab.yml              # Single source of truth (engineer edits only this)

iac/
├── terraform/           # Terraform configurations for VM lifecycle
│   ├── providers.tf
│   ├── variables.tf
│   └── vm.tf
├── ansible/
│   ├── playbooks/
│   │   ├── provision-host.yml
│   │   ├── deploy-vpn.yml
│   │   ├── deploy-app.yml        # NEW: deploys a container via docker_container
│   │   ├── deploy-proxy.yml      # NEW: deploys Traefik via docker_container
│   │   ├── deploy-updater.yml    # NEW: deploys systemd timer + update script
│   │   └── security-audit.yml
│   └── roles/
│       ├── hardening/            # UNCHANGED: OS hardening
│       ├── hypervisor/           # UNCHANGED: KVM/libvirt setup
│       ├── nat-network/          # UNCHANGED: NAT + iptables
│       ├── vm-create/            # UNCHANGED: VM creation via cloud-init
│       ├── vm-guest/             # MODIFIED: remove Nomad client, keep Docker
│       ├── wireguard/            # UNCHANGED: WireGuard VPN
│       ├── app-container/        # NEW: deploy a Docker container via docker_container
│       ├── proxy-container/      # NEW: deploy Traefik via docker_container
│       └── auto-updater/         # NEW: systemd timer + update script
├── cli/
│   ├── cmd/
│   │   ├── audit.go              # UNCHANGED
│   │   ├── deploy_app.go         # MODIFIED: run Ansible playbook instead of Nomad
│   │   ├── deploy_proxy.go       # MODIFIED: run Ansible playbook instead of Nomad
│   │   ├── deploy_vpn.go         # UNCHANGED
│   │   ├── generate.go           # MODIFIED: remove Nomad generation
│   │   ├── generate_vpn_client.go # UNCHANGED
│   │   ├── provision.go          # MODIFIED: remove Nomad provisioning step
│   │   ├── status.go             # MODIFIED: query Docker/systemd instead of Nomad
│   │   ├── teardown.go           # UNCHANGED (already per-host from scaling fix)
│   │   ├── update.go             # MODIFIED: remove trigger/toggle subcommands
│   │   ├── update_status.go      # MODIFIED: query systemd/journald instead of Nomad
│   │   └── [REMOVED: update_trigger.go, update_toggle.go]
│   ├── config/
│   │   ├── config.go             # MODIFIED: remove Datacenter field
│   │   └── validate.go           # MODIFIED: remove datacenter validation
│   ├── generators/
│   │   ├── ansible.go            # MODIFIED: remove Nomad cluster vars from group_vars
│   │   ├── terraform.go          # UNCHANGED
│   │   └── [REMOVED: nomad.go]
│   ├── updater/
│   │   └── [REMOVED: updater.go]
│   ├── wireguard/
│   │   └── peers.go              # UNCHANGED
│   ├── go.mod
│   ├── go.sum
│   └── main.go                   # MODIFIED: remove update trigger/toggle routing
├── [REMOVED: nomad/]             # Entire directory removed
└── [REMOVED: templates/nomad-*.hcl.tmpl]

.generated/                       # Generated configs (gitignored)
├── ansible/inventory/            # Generated Ansible inventory
└── terraform/<hostname>/         # Per-host Terraform tfvars and state
```

**Structure Decision**: Same single-project IaC repository structure as 001. Nomad-specific directories (`iac/nomad/`, `iac/templates/nomad-*.hcl.tmpl`) are removed. Three new Ansible roles (`app-container`, `proxy-container`, `auto-updater`) replace all Nomad job management. The `.generated/nomad/` output directory is no longer produced.

## Complexity Tracking

(No violations to justify.)

## Phase 0: Outline & Research

**Goal**: Determine how to replace Nomad's container scheduling, auto-updates, and health-checked deployments with simpler alternatives.

1. **Container Management Without a Scheduler**: Research Ansible `docker_container` module + Docker `restart_policy: unless-stopped` as replacement for Nomad service jobs.
2. **Auto-Updates Without Nomad**: Research systemd timers + shell scripts as replacement for Nomad periodic batch job. Design per-container run scripts for consistent recreation.
3. **Rollback Strategy**: Confirm deferral to future feature per user decision.
4. **Ansible Deployment Pattern**: Research `docker_image_pull` + conditional `recreate` pattern.
5. **Config Schema Changes**: Identify fields to remove (`cluster.datacenter`, `auto_update.auto_revert`, `auto_update.health_check_timeout`).

**Output**: research.md (completed)

## Phase 1: Design & Contracts

**Prerequisites**: `research.md` complete

1. **Data Model** (`data-model.md`):
    - Remove Nomad_Cluster, Nomad_Job, Auto_Update_Job entities.
    - Add Docker_Container entity (with restart_policy, run_script_path).
    - Add Auto_Update_Timer entity (systemd timer/service/script).
    - Update relationships: VM -> Docker_Container (1:N) replaces Nomad_Cluster -> Nomad_Job.

2. **Contracts** (`contracts/`):
    - Update CLI command contracts: deploy app/proxy use Ansible instead of Nomad.
    - Remove Nomad job file schema section.
    - Update config schema: remove `cluster.datacenter`, `auto_update.auto_revert`, `auto_update.health_check_timeout`.
    - Add generate configs contract and teardown contract.

3. **Quickstart** (`quickstart.md`):
    - Remove all Nomad references (Nomad CLI prerequisite, Nomad in "what happens" descriptions).
    - Update architecture diagram: no Nomad server/client, containers managed by Ansible.
    - Update auto-update description: systemd timer instead of Nomad periodic job.

4. **Agent context update**:
    - Run `.specify/scripts/bash/update-agent-context.sh gemini`.

5. **Post-Design Constitution Re-Check**:
    - Verify all four principles still pass after detailed design.

**Output**: data-model.md, contracts/*, quickstart.md, agent-specific context file.

## Phase 2: Remove Nomad Infrastructure

**Goal**: Delete all Nomad-related code and configuration.

1. **Delete Nomad Ansible roles**:
    - Remove `iac/ansible/roles/nomad/` directory entirely.
    - Remove `iac/ansible/roles/vm-guest/templates/nomad-client.hcl.j2`.
    - Update `iac/ansible/roles/vm-guest/tasks/main.yml` to remove Nomad client installation tasks.

2. **Delete Nomad job files and templates**:
    - Remove `iac/nomad/` directory entirely.
    - Remove `iac/templates/nomad-app.hcl.tmpl`, `nomad-proxy.hcl.tmpl`, `nomad-updater.hcl.tmpl`.

3. **Delete Nomad Terraform resources**:
    - Remove `iac/terraform/nomad-jobs.tf`.

4. **Delete Nomad Go code**:
    - Remove `iac/cli/generators/nomad.go`.
    - Remove `iac/cli/updater/updater.go`.
    - Remove `iac/cli/cmd/update_trigger.go` and `iac/cli/cmd/update_toggle.go`.

5. **Update Go config**:
    - Remove `Datacenter` field from `config.go` struct.
    - Remove `auto_update.auto_revert` and `auto_update.health_check_timeout` from config struct.
    - Update `validate.go` to remove datacenter validation.

6. **Update Go generators**:
    - Remove `GenerateNomadJobs` call from `generate.go`.
    - Remove Nomad cluster vars (`nomad_server_bootstrap_expect`, `nomad_server_ips`) from `ansible.go` group_vars generation.

7. **Verify**: `go build ./...` compiles cleanly.

## Phase 3: Create Container Deployment via Ansible

**Goal**: Replace Nomad job submissions with Ansible-managed Docker containers.

1. **Create `app-container` Ansible role**:
    - `tasks/main.yml`: Pull image via `docker_image_pull`, deploy container via `docker_container` with `read_only: true`, `restart_policy: unless-stopped`, volume mounts, port mappings.
    - `defaults/main.yml`: Default values for app configuration.
    - `templates/run-app.sh.j2`: Per-container run script that defines the full `docker run` invocation (used by auto-updater).

2. **Create `proxy-container` Ansible role**:
    - `tasks/main.yml`: Pull Traefik image, deploy via `docker_container`.
    - `defaults/main.yml`: Default proxy configuration.
    - `templates/traefik.yml.j2`: Traefik static configuration.
    - `templates/run-proxy.sh.j2`: Per-container run script for proxy.

3. **Create deployment playbooks**:
    - `iac/ansible/playbooks/deploy-app.yml`: Targets the VM, applies `app-container` role with app-specific variables.
    - `iac/ansible/playbooks/deploy-proxy.yml`: Targets the VM, applies `proxy-container` role.

4. **Update Go CLI `deploy_app.go`**:
    - Replace Nomad job generation + submission with: generate configs, then run `ansible-playbook deploy-app.yml` with extra vars for the app name.

5. **Update Go CLI `deploy_proxy.go`**:
    - Replace Nomad job submission with: run `ansible-playbook deploy-proxy.yml`.

6. **Verify**: `go build ./...` compiles cleanly.

## Phase 4: Create Auto-Update System

**Goal**: Replace Nomad periodic batch job with systemd timer + shell script.

1. **Create `auto-updater` Ansible role**:
    - `templates/update-containers.sh.j2`: Shell script that iterates over configured containers, pulls latest image, compares digest, and recreates container if changed (using per-container run scripts).
    - `templates/homelab-updater.service.j2`: systemd service unit that runs the update script.
    - `templates/homelab-updater.timer.j2`: systemd timer unit with `Persistent=true` and schedule from config.
    - `tasks/main.yml`: Template the script and units, enable the timer.
    - `defaults/main.yml`: Default schedule, container list.

2. **Create deployment playbook**:
    - `iac/ansible/playbooks/deploy-updater.yml`: Targets the VM, applies `auto-updater` role.

3. **Update Go CLI `deploy_app.go`**:
    - After deploying a container, also deploy the auto-updater if `auto_update.enabled` is true.

4. **Update Go CLI `update_status.go`**:
    - Query systemd timer status and journald logs via SSH/Ansible instead of Nomad API.

5. **Update Go CLI `update.go`**:
    - Remove `trigger` and `toggle` subcommands (only `status` remains).

6. **Update Go CLI `main.go`**:
    - Remove routing for `update trigger`, `update enable`, `update disable`.

7. **Verify**: `go build ./...` compiles cleanly.

## Phase 5: Update Remaining Components

**Goal**: Update all remaining code that references Nomad.

1. **Update `provision.go`**:
    - Remove Nomad setup step from host provisioning workflow.
    - Remove any Nomad-related Ansible playbook invocations.

2. **Update `status.go`**:
    - Replace Nomad job status queries with Docker container status queries (via SSH or Ansible).
    - Add systemd timer status for auto-updater.

3. **Update `vm-guest` role**:
    - Remove Nomad client installation tasks.
    - Remove `nomad-client.hcl.j2` template reference.
    - Keep Docker installation tasks.

4. **Update `provision-host.yml` playbook**:
    - Remove `nomad` role from the play.

5. **Update `security-audit.yml` playbook**:
    - Remove Nomad-specific audit checks.
    - Add Docker container audit checks (read-only root, restart policy).

6. **Update `homelab.yml.example`** (if it exists):
    - Remove `cluster.datacenter` field.

7. **Verify**: `go build ./...` compiles cleanly.

## Phase 6: Validation and Cleanup

**Goal**: Ensure everything is consistent and the system compiles.

1. **Full build verification**: `cd iac/cli && go build ./...`
2. **Verify generated output**: Run `iac generate` with a test config and inspect:
    - `.generated/ansible/inventory/` has correct inventory without Nomad references
    - `.generated/terraform/<hostname>/` has per-host tfvars
    - No `.generated/nomad/` directory is produced
3. **Update README.md**: Remove Nomad from prerequisites and architecture descriptions.
4. **Final review**: Grep entire codebase for remaining "nomad" or "Nomad" references and clean up.

## Key Rules

- Use absolute paths
- ERROR on gate failures or unresolved clarifications
- Rollback is explicitly deferred -- do not implement health checks or auto-revert
- All containers use `restart_policy: unless-stopped` for crash/reboot recovery
- Per-container run scripts ensure the auto-updater recreates containers with the exact same configuration as Ansible
