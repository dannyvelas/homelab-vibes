# Tasks: IaC Homelab Media

**Input**: Design documents from `/specs/001-iac-homelab-media/`
**Prerequisites**: plan.md, spec.md, research.md, data-model.md, contracts/README.md, quickstart.md

**Organization**: Tasks grouped by user story for independent implementation and testing.

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Can run in parallel (different files, no dependencies)
- **[Story]**: Which user story this task belongs to (US1–US5)
- All paths are relative to repository root

---

## Phase 1: Setup

**Purpose**: Repository structure, tooling initialization, example configuration

- [ ] T001 Create directory structure: `iac/terraform/`, `iac/ansible/playbooks/`, `iac/ansible/roles/`, `iac/nomad/jobs/`, `iac/templates/`, `iac/cli/`, `iac/docs/`, `tests/integration/`, `tests/unit/`
- [ ] T002 Create `.gitignore` with entries for `.generated/`, `.secrets/`, `*.tfstate`, `*.tfstate.backup`, `.terraform/`
- [ ] T003 [P] Initialize Go module for the `iac` CLI in `iac/cli/go.mod`
- [ ] T004 [P] Create `homelab.yml.example` at repo root with all configuration sections (cluster, hosts, vpn, storage, apps, auto_update, secrets) per the schema in contracts/README.md
- [ ] T005 [P] Create Terraform provider configuration for `dmacvicar/libvirt` and `hashicorp/nomad` in `iac/terraform/providers.tf`
- [ ] T006 [P] Create Terraform variable declarations (matching homelab.yml schema) in `iac/terraform/variables.tf`

---

## Phase 2: Foundational (Blocking Prerequisites)

**Purpose**: Ansible roles and Terraform configs that ALL user stories depend on. Must complete before any story.

**⚠️ CRITICAL**: No user story work can begin until this phase is complete.

### Host Provisioning Roles

- [ ] T007 Create Ansible role for Debian OS hardening (UFW, SSH key-only, disable root login, auto-updates, least privilege users) in `iac/ansible/roles/hardening/`
- [ ] T008 [P] Create Ansible role for KVM/libvirt hypervisor installation on Debian in `iac/ansible/roles/hypervisor/`
- [ ] T009 [P] Create Ansible role for Nomad server+client installation on Debian host in `iac/ansible/roles/nomad/`
- [ ] T010 Create Ansible role for NAT networking (Linux bridge, private subnet, iptables masquerade, egress blocking, port forwarding) in `iac/ansible/roles/nat-network/`
- [ ] T011 Create Ansible role for workload VM creation via libvirt (Debian guest, resource allocation, cloud-init) in `iac/ansible/roles/vm-create/`
- [ ] T012 Create Ansible role for VM guest setup (OS hardening, Nomad client, Docker engine installation) in `iac/ansible/roles/vm-guest/`

### Terraform Infrastructure

- [ ] T013 Create Terraform resource definitions for libvirt VMs (domain, volume, cloud-init) in `iac/terraform/vm.tf`
- [ ] T014 [P] Create Terraform resource definitions for Nomad jobs (using `hashicorp/nomad` provider) in `iac/terraform/nomad-jobs.tf`

### Host Provisioning Playbook

- [ ] T015 Create Ansible playbook that orchestrates all host provisioning roles (hardening → hypervisor → Nomad → NAT network → VM creation → VM guest setup) in `iac/ansible/playbooks/provision-host.yml`

**Checkpoint**: Running the provisioning playbook against a Debian server should produce a hardened host with KVM, Nomad, NAT networking, and a workload VM with Docker and Nomad client.

---

## Phase 3: User Story 5 — Engineer Manages All Configuration from a Single Source of Truth (Priority: P1) 🎯 MVP

**Goal**: Engineer edits only `homelab.yml`; the `iac` CLI generates all tool-specific configs and orchestrates Terraform/Ansible/Nomad.

**Independent Test**: Change a server IP in `homelab.yml`, run `iac provision host`, verify Ansible inventory and Terraform tfvars both reflect the new IP without manual edits.

### Implementation

- [ ] T016 [US5] Implement `homelab.yml` parser in Go — read and validate the YAML schema (cluster, hosts, vpn, storage, apps, auto_update, secrets sections) in `iac/cli/config/config.go`
- [ ] T017 [P] [US5] Implement Ansible inventory generator — read `homelab.yml` hosts section and output `hosts.yml` + `group_vars/` into `.generated/ansible/inventory/` in `iac/cli/generators/ansible.go`
- [ ] T018 [P] [US5] Implement Terraform tfvars generator — read `homelab.yml` and output `terraform.tfvars` into `.generated/terraform/` in `iac/cli/generators/terraform.go`
- [ ] T019 [P] [US5] Implement Nomad HCL job file generator — read `homelab.yml` apps section and template Nomad job files into `.generated/nomad/` in `iac/cli/generators/nomad.go`
- [ ] T020 [P] [US5] Create Nomad job templates (Go templates) for media apps in `iac/templates/nomad-app.hcl.tmpl`
- [ ] T021 [P] [US5] Create Nomad job template for reverse proxy in `iac/templates/nomad-proxy.hcl.tmpl`
- [ ] T022 [P] [US5] Create Nomad job template for auto-updater batch job in `iac/templates/nomad-updater.hcl.tmpl`
- [ ] T023 [US5] Implement `iac` CLI main entrypoint with subcommand routing (provision, deploy, generate, status, update) in `iac/cli/main.go`
- [ ] T024 [US5] Implement `iac provision host` subcommand — runs config generation then orchestrates Terraform apply + Ansible playbook in `iac/cli/cmd/provision.go`
- [ ] T025 [US5] Implement `iac status` subcommand — queries Nomad API, WireGuard status, host health and displays formatted output in `iac/cli/cmd/status.go`
- [ ] T026 [US5] Implement config validation — verify `homelab.yml` schema, required fields, valid IP ranges, no duplicate host names in `iac/cli/config/validate.go`
- [ ] T027 [US5] Build and test the `iac` binary — `go build` in `iac/cli/`, verify `iac provision host` generates correct configs in `.generated/` from a test `homelab.yml`

**Checkpoint**: Engineer can edit `homelab.yml`, run `iac provision host --name <host>`, and have all tool-specific configs auto-generated and the provisioning playbook executed. No manual editing of Ansible inventory, Terraform tfvars, or Nomad HCL files.

---

## Phase 4: User Story 1 — Remote Engineer Securely Accesses Home LAN Resources via VPN (Priority: P1)

**Goal**: Engineer deploys WireGuard on a host with one command. Remote team members connect via generated client configs and access all LAN resources.

**Independent Test**: Run `iac deploy vpn`, then `iac generate vpn-client --name test`. Import the generated config into a WireGuard client outside the LAN. Verify: connected client can ping the server's LAN IP and access a web service on a non-public port.

### Implementation

- [ ] T028 [US1] Create Ansible role for WireGuard server deployment (kernel module, key generation, wg0 interface, systemd service, iptables forwarding rules) in `iac/ansible/roles/wireguard/`
- [ ] T029 [US1] Create Ansible playbook for VPN deployment that applies the WireGuard role in `iac/ansible/playbooks/deploy-vpn.yml`
- [ ] T030 [US1] Implement `iac deploy vpn` subcommand — reads VPN config from `homelab.yml`, runs config generation, executes Ansible playbook in `iac/cli/cmd/deploy_vpn.go`
- [ ] T031 [US1] Implement `iac generate vpn-client` subcommand — generates client key pair, assigns VPN IP, creates client config file, generates QR code, adds peer to server config in `iac/cli/cmd/generate_vpn_client.go`
- [ ] T032 [P] [US1] Implement WireGuard peer management utilities (add peer, remove peer, list peers) in `iac/cli/wireguard/peers.go`
- [ ] T033 [P] [US1] Create Ansible task for optional ddclient (dynamic DNS) deployment within the WireGuard role in `iac/ansible/roles/wireguard/tasks/ddns.yml`

**Checkpoint**: Engineer runs `iac deploy vpn --host homelab-host-01`, then `iac generate vpn-client --name danny-laptop`. Remote engineer imports the config into WireGuard client and can access all home LAN resources.

---

## Phase 5: User Story 2 — Engineer Deploys Media Stack Applications (Priority: P1)

**Goal**: Engineer deploys Plex, Sonarr, Radarr, Bazarr with individual `iac deploy app` commands. Each app is accessible via its web UI within the LAN.

**Independent Test**: Run `iac deploy app --name plex`. Access `http://<server-ip>:32400` from the LAN and verify the Plex web UI loads. Repeat for Sonarr (8989), Radarr (7878), Bazarr (6767).

**Depends on**: Phase 2 (Nomad cluster, VMs with Docker), Phase 3 (config generation from homelab.yml)

### Implementation

- [ ] T034 [US2] Create Nomad job HCL for reverse proxy (Traefik or Caddy) with TLS termination, rate limiting, and dynamic routing in `iac/nomad/jobs/proxy.hcl`
- [ ] T035 [US2] Implement `iac deploy proxy` subcommand — submits reverse proxy Nomad job, configures host iptables port forwarding (443 → VM proxy) in `iac/cli/cmd/deploy_proxy.go`
- [ ] T036 [P] [US2] Create Nomad job HCL for Plex with read-only root, writable /config /media volumes, port 32400 mapping in `iac/nomad/jobs/plex.hcl`
- [ ] T037 [P] [US2] Create Nomad job HCL for Sonarr with read-only root, writable /config /downloads /media volumes, port 8989 mapping in `iac/nomad/jobs/sonarr.hcl`
- [ ] T038 [P] [US2] Create Nomad job HCL for Radarr with read-only root, writable /config /downloads /media volumes, port 7878 mapping in `iac/nomad/jobs/radarr.hcl`
- [ ] T039 [P] [US2] Create Nomad job HCL for Bazarr with read-only root, writable /config /downloads /media volumes, port 6767 mapping in `iac/nomad/jobs/bazarr.hcl`
- [ ] T040 [US2] Implement `iac deploy app` subcommand — reads app config from `homelab.yml`, generates Nomad HCL from template, submits job to Nomad cluster in `iac/cli/cmd/deploy_app.go`
- [ ] T041 [US2] Configure storage passthrough: Ansible tasks to create host mount points, libvirt storage pool for VM passthrough, Docker volume mounts in Nomad jobs in `iac/ansible/roles/vm-guest/tasks/storage.yml`
- [ ] T042 [US2] Configure reverse proxy routes for each media app (path-based routing: /plex, /sonarr, /radarr, /bazarr) in reverse proxy config within `iac/nomad/jobs/proxy.hcl`

**Checkpoint**: Engineer runs `iac deploy proxy` then `iac deploy app --name plex` (and sonarr/radarr/bazarr). All four media app web UIs are accessible from the LAN via their respective ports and via reverse proxy paths.

---

## Phase 6: User Story 3 — Robust Security Measures Protect Infrastructure (Priority: P2)

**Goal**: Verify and harden all defense-in-depth layers: OS hardening, NAT isolation, egress blocking, VM kernel isolation, reverse proxy hardening, read-only containers, least privilege.

**Independent Test**: Run a security audit script. Verify: VMs cannot reach LAN devices (egress blocked), containers run read-only, all services run as non-root, firewall rules match expected policy, reverse proxy has TLS and rate limiting active.

**Depends on**: Phase 2 (foundational infrastructure), Phase 5 (media apps deployed to audit)

### Implementation

- [ ] T043 [US3] Harden iptables egress rules — verify and tighten rules to deny all VM-to-LAN traffic except explicitly allowed (internet, DNS, NTP) in `iac/ansible/roles/nat-network/tasks/egress.yml`
- [ ] T044 [P] [US3] Harden reverse proxy — enforce TLS-only, add security headers (HSTS, X-Frame-Options, CSP), configure rate limiting in `iac/nomad/jobs/proxy.hcl`
- [ ] T045 [P] [US3] Verify all Nomad job HCL files use `readonly_rootfs = true` and explicit writable volume mounts only for required paths (/config, /downloads, /media) — audit `iac/nomad/jobs/*.hcl`
- [ ] T046 [P] [US3] Verify all containers and services run as non-root users — add `user` directive to Nomad job HCL task configs in `iac/nomad/jobs/*.hcl`
- [ ] T047 [US3] Create security audit playbook that checks: UFW status, SSH config, iptables rules, container read-only status, user privileges, open ports, egress connectivity from VM in `iac/ansible/playbooks/security-audit.yml`
- [ ] T048 [US3] Implement `iac audit security` subcommand — runs the security audit playbook and reports pass/fail for each check in `iac/cli/cmd/audit.go`

**Checkpoint**: Running `iac audit security` produces a report showing all defense-in-depth layers are correctly configured. VMs cannot reach LAN devices, containers are read-only, all services run non-root, firewall rules are tight.

---

## Phase 7: User Story 4 — Media Applications Automatically Update to Latest Versions (Priority: P2)

**Goal**: Deployed media apps automatically detect new Docker image versions and update with health-checked rolling deployments and automatic rollback on failure.

**Independent Test**: Deploy Plex at a known image digest. Manually trigger an update check. Verify: new image is pulled, Nomad performs a deployment, health check passes, app runs the new version. Then simulate a failure (bad health check) and verify auto-rollback to the previous version.

**Depends on**: Phase 5 (media apps deployed)

### Implementation

- [ ] T049 [US4] Add Nomad `update` stanza with `auto_revert = true`, health checks (HTTP on web UI port), and configurable timeout to all media app job templates in `iac/templates/nomad-app.hcl.tmpl`
- [ ] T050 [US4] Create Nomad periodic batch job for auto-updater — pulls latest images, compares digests, triggers `nomad job run` for apps with new images in `iac/nomad/jobs/auto-updater.hcl`
- [ ] T051 [US4] Implement the auto-updater logic as a Go script or shell script invoked by the Nomad batch job — Docker image digest comparison, Nomad API calls for redeployment in `iac/cli/updater/updater.go`
- [ ] T052 [US4] Implement `iac update status` subcommand — show auto-update status for all apps (current digest, last check, last result) in `iac/cli/cmd/update_status.go`
- [ ] T053 [P] [US4] Implement `iac update trigger --name <app>` subcommand — manually trigger an update check for a specific app in `iac/cli/cmd/update_trigger.go`
- [ ] T054 [P] [US4] Implement `iac update enable/disable --name <app>` subcommands — toggle auto-updates per app in `iac/cli/cmd/update_toggle.go`

**Checkpoint**: Auto-updater runs on schedule (configurable in homelab.yml). When a new image is available, Nomad deploys it with health checks. If the new version fails, Nomad reverts automatically. `iac update status` shows current state.

---

## Phase 8: Polish & Cross-Cutting Concerns

**Purpose**: End-to-end validation, documentation, and cleanup

- [ ] T055 Implement `iac teardown` subcommand — runs `terraform destroy` to cleanly remove all VMs and Nomad jobs in `iac/cli/cmd/teardown.go`
- [ ] T056 [P] Add `homelab.yml` schema validation with helpful error messages for common mistakes (missing required fields, invalid IP format, duplicate host names) in `iac/cli/config/validate.go`
- [ ] T057 [P] Create end-to-end deployment test script — provisions hosts, deploys VPN, deploys media stack, runs security audit, verifies all services accessible in `tests/integration/e2e_deploy.sh`
- [ ] T058 Validate quickstart.md — follow the guide from scratch on a test environment and verify all steps work as documented
- [ ] T059 [P] Add `iac` CLI help text and usage documentation for all subcommands in `iac/cli/cmd/*.go`

---

## Dependencies & Execution Order

### Phase Dependencies

- **Setup (Phase 1)**: No dependencies — start immediately
- **Foundational (Phase 2)**: Depends on Phase 1 — BLOCKS all user stories
- **US5 - Unified Config (Phase 3)**: Depends on Phase 2 — BLOCKS US1, US2, US4 (they use the `iac` CLI)
- **US1 - VPN (Phase 4)**: Depends on Phase 3 (`iac deploy vpn` reads from `homelab.yml`)
- **US2 - Media Stack (Phase 5)**: Depends on Phase 3 (`iac deploy app` reads from `homelab.yml`)
- **US3 - Security (Phase 6)**: Depends on Phases 2 + 5 (need deployed infrastructure to audit)
- **US4 - Auto-Updates (Phase 7)**: Depends on Phase 5 (need deployed apps to update)
- **Polish (Phase 8)**: Depends on all prior phases

### User Story Dependencies

- **US5 (P1)**: Foundational dependency — implements the `iac` CLI that all other stories use
- **US1 (P1)**: Depends on US5 only. Can run in parallel with US2.
- **US2 (P1)**: Depends on US5 only. Can run in parallel with US1.
- **US3 (P2)**: Depends on US2 (need deployed apps to audit security posture)
- **US4 (P2)**: Depends on US2 (need deployed apps to auto-update)

### Within Each User Story

- Config/schema work before CLI commands
- CLI commands before verification/testing
- Templates before generators that consume them

### Parallel Opportunities

- **Phase 1**: T003, T004, T005, T006 can all run in parallel
- **Phase 2**: T008, T009 can run in parallel; T013, T014 can run in parallel
- **Phase 3**: T017, T018, T019 (generators) can run in parallel; T020, T021, T022 (templates) can run in parallel
- **Phase 4**: T032, T033 can run in parallel
- **Phase 5**: T036, T037, T038, T039 (all Nomad job HCLs) can run in parallel
- **Phase 6**: T044, T045, T046 can run in parallel
- **Phase 7**: T053, T054 can run in parallel
- **US1 and US2**: Can be worked on in parallel after US5 is complete

---

## Parallel Example: Phase 5 (Media Stack)

```bash
# Launch all Nomad job HCL files in parallel (different files, no dependencies):
Task: T036 "Create Plex Nomad job in iac/nomad/jobs/plex.hcl"
Task: T037 "Create Sonarr Nomad job in iac/nomad/jobs/sonarr.hcl"
Task: T038 "Create Radarr Nomad job in iac/nomad/jobs/radarr.hcl"
Task: T039 "Create Bazarr Nomad job in iac/nomad/jobs/bazarr.hcl"

# Then sequentially (depends on HCL files existing):
Task: T040 "Implement iac deploy app subcommand"
```

---

## Implementation Strategy

### MVP First (US5 + US1 + US2)

1. Complete Phase 1: Setup
2. Complete Phase 2: Foundational Ansible roles + Terraform configs
3. Complete Phase 3: US5 — `iac` CLI + `homelab.yml` + config generation
4. Complete Phase 4: US1 — VPN deployment + client config generation
5. Complete Phase 5: US2 — Media stack deployment
6. **STOP and VALIDATE**: Can an engineer edit `homelab.yml`, provision hosts, deploy VPN, and deploy all four media apps? Are they accessible?

### Incremental Delivery

1. Setup + Foundational → Infrastructure ready
2. Add US5 (Unified Config) → Single config file working, `iac` CLI functional (MVP foundation!)
3. Add US1 (VPN) → Remote access working → Test independently
4. Add US2 (Media Stack) → Apps deployed and accessible → Test independently
5. Add US3 (Security) → Hardened and auditable → Test independently
6. Add US4 (Auto-Updates) → Hands-off maintenance → Test independently
7. Each story adds value without breaking previous stories

---

## Summary

| Metric | Value |
|--------|-------|
| Total tasks | 59 |
| Phase 1 (Setup) | 6 |
| Phase 2 (Foundational) | 9 |
| Phase 3 (US5 - Unified Config) | 12 |
| Phase 4 (US1 - VPN) | 6 |
| Phase 5 (US2 - Media Stack) | 9 |
| Phase 6 (US3 - Security) | 6 |
| Phase 7 (US4 - Auto-Updates) | 6 |
| Phase 8 (Polish) | 5 |
| Parallel opportunities | 8 groups of parallelizable tasks |
| Suggested MVP scope | US5 + US1 + US2 (Phases 1–5) |

## Notes

- [P] tasks = different files, no dependencies on incomplete tasks
- [Story] label maps task to specific user story for traceability
- Each user story is independently completable and testable after its checkpoint
- Commit after each task or logical group
- Stop at any checkpoint to validate the story independently
- US5 (Unified Config) is Phase 3 because it implements the `iac` CLI that all other stories depend on
- No test tasks generated — tests were not explicitly requested in the spec
