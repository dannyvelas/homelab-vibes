# Tasks: Simplify IaC Media (Remove Nomad)

**Input**: Design documents from `/specs/002-simplify-iac-media/`
**Prerequisites**: plan.md (required), spec.md (required for user stories), research.md, data-model.md, contracts/

**Tests**: Not requested in the feature specification.

**Organization**: Tasks are grouped by user story to enable independent implementation and testing of each story. Since this is a refactoring feature (modifying existing code, not building from scratch), US5 (Single Config) and US6 (Fewer Resources) are achieved through the foundational Nomad removal phase rather than having separate implementation.

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Can run in parallel (different files, no dependencies)
- **[Story]**: Which user story this task belongs to (e.g., US1, US2, US3)
- Include exact file paths in descriptions

---

## Phase 1: Setup

**Purpose**: Prepare for Nomad removal and verify current state

- [X] T001 Verify current Go CLI compiles cleanly with `cd iac/cli && go build ./...`
- [X] T002 [P] Verify homelab.yml.example exists and note current schema at iac/cli/homelab.yml

---

## Phase 2: Foundational - Remove Nomad (Blocking Prerequisites) [US5, US6]

**Purpose**: Delete all Nomad-related code and configuration. MUST complete before any user story implementation since Go code won't compile with partial Nomad references.

**Warning**: No user story work can begin until this phase is complete

### Delete Nomad Files

- [X] T003 [P] Delete Nomad Ansible server role directory at iac/ansible/roles/nomad/
- [X] T004 [P] Delete Nomad job files directory at iac/nomad/
- [X] T005 [P] Delete Nomad HCL templates: iac/templates/nomad-app.hcl.tmpl, iac/templates/nomad-proxy.hcl.tmpl, iac/templates/nomad-updater.hcl.tmpl
- [X] T006 [P] Delete Nomad Terraform resource file at iac/terraform/nomad-jobs.tf
- [X] T007 [P] Delete Nomad Go generator at iac/cli/generators/nomad.go
- [X] T008 [P] Delete Nomad Go updater package at iac/cli/updater/updater.go
- [X] T009 [P] Delete Go CLI files iac/cli/cmd/update_trigger.go and iac/cli/cmd/update_toggle.go

### Update Go Config Schema

- [X] T010 [US5] Remove `Datacenter` field from `ClusterConfig` struct in iac/cli/config/config.go
- [X] T011 [US5] Remove `AutoRevert` and `HealthCheckTimeout` fields from `AutoUpdateConfig` struct in iac/cli/config/config.go
- [X] T012 [US5] Remove Datacenter and HealthCheckTimeout defaults from `applyDefaults()` in iac/cli/config/config.go
- [X] T013 [US5] Remove datacenter validation (if any) from iac/cli/config/validate.go

### Update Go Generators

- [X] T014 [US5] Remove `GenerateNomadJobs` call from `generateConfigs()` in iac/cli/cmd/generate.go
- [X] T015 [US5] Remove Nomad cluster vars (`nomad_server_bootstrap_expect`, `nomad_server_ips`) from `generateGroupVars()` in iac/cli/generators/ansible.go
- [X] T016 [US6] Remove `nomad_address` variable generation from `generateHostTfvars()` in iac/cli/generators/terraform.go

### Update Go CLI Commands (Remove Nomad References)

- [X] T017 Remove `GenerateNomadJobs` call and `nomad job run` execution from `deployApp()` in iac/cli/cmd/deploy_app.go (stub with Ansible playbook call placeholder)
- [X] T018 Remove `GenerateNomadJobs` call and `nomad job run` execution from `deployProxy()` in iac/cli/cmd/deploy_proxy.go (stub with Ansible playbook call placeholder)
- [X] T019 Remove `GenerateNomadJobs` call from `provisionHost()` in iac/cli/cmd/provision.go and remove Nomad status message
- [X] T020 Replace `printNomadJobs()` function with `printContainerStatus()` stub in iac/cli/cmd/status.go
- [X] T021 Rewrite `updateStatus()` in iac/cli/cmd/update_status.go to stub SSH-based systemd timer query instead of Nomad API
- [X] T022 Remove `trigger`, `enable`, `disable` subcommand routing from `Update()` in iac/cli/cmd/update.go
- [X] T023 Remove Nomad references from usage text and routing in iac/cli/main.go
- [X] T024 Remove Nomad reference from completion message in iac/cli/cmd/teardown.go (line 44: "All VMs and Nomad jobs destroyed")

### Update Ansible (Remove Nomad)

- [X] T025 Remove Nomad client installation tasks (lines 72-111) and Nomad UFW rule (lines 25-29) from iac/ansible/roles/vm-guest/tasks/main.yml
- [X] T026 Delete Nomad client template at iac/ansible/roles/vm-guest/templates/nomad-client.hcl.j2
- [X] T027 Remove `nomad` role inclusion (lines 24-26) from iac/ansible/playbooks/provision-host.yml and update comments

### Verify Compilation

- [X] T028 Verify Go CLI compiles cleanly with `cd iac/cli && go build ./...`

**Checkpoint**: All Nomad code is removed. Go compiles cleanly. Config schema is updated (US5). No Nomad processes will run (US6). Foundation ready for new container deployment code.

---

## Phase 3: User Story 2 - Engineer Deploys Media Stack Applications (Priority: P1)

**Goal**: Engineer can deploy Plex, Sonarr, Radarr, Bazarr as Docker containers via Ansible `docker_container` module with a single CLI command per app. Containers have read-only root filesystems, writable data volumes, and `restart_policy: unless-stopped`.

**Independent Test**: Run `iac deploy app --name plex`, then verify a Plex container is running inside the VM with the correct image, read-only root, volume mounts, port mapping, and restart policy.

### Ansible Roles

- [X] T029 [P] [US2] Create `app-container` role defaults at iac/ansible/roles/app-container/defaults/main.yml with default app config variables
- [X] T030 [P] [US2] Create `app-container` role tasks at iac/ansible/roles/app-container/tasks/main.yml using `community.docker.docker_image_pull` and `community.docker.docker_container` with read_only, restart_policy, volumes, ports
- [X] T031 [US2] Create per-container run script template at iac/ansible/roles/app-container/templates/run-app.sh.j2 defining full `docker run` invocation for auto-updater use
- [X] T032 [P] [US2] Create `proxy-container` role defaults at iac/ansible/roles/proxy-container/defaults/main.yml with Traefik config variables
- [X] T033 [P] [US2] Create `proxy-container` role tasks at iac/ansible/roles/proxy-container/tasks/main.yml using `docker_image_pull` and `docker_container` for Traefik
- [X] T034 [US2] Create Traefik static config template at iac/ansible/roles/proxy-container/templates/traefik.yml.j2 with entrypoints and provider config
- [X] T035 [US2] Create proxy run script template at iac/ansible/roles/proxy-container/templates/run-proxy.sh.j2

### Ansible Playbooks

- [X] T036 [P] [US2] Create deploy-app playbook at iac/ansible/playbooks/deploy-app.yml targeting VM hosts, applying app-container role with app-specific extra vars
- [X] T037 [P] [US2] Create deploy-proxy playbook at iac/ansible/playbooks/deploy-proxy.yml targeting VM hosts, applying proxy-container role

### Go CLI Updates

- [X] T038 [US2] Implement `deployApp()` in iac/cli/cmd/deploy_app.go to run `ansible-playbook deploy-app.yml` with extra vars for app name, image, port, storage paths
- [X] T039 [US2] Implement `deployProxy()` in iac/cli/cmd/deploy_proxy.go to run `ansible-playbook deploy-proxy.yml`

### Build Verification

- [X] T040 [US2] Verify Go CLI compiles cleanly with `cd iac/cli && go build ./...`

**Checkpoint**: `iac deploy app --name <app>` and `iac deploy proxy` work via Ansible. Containers run with read-only root, unless-stopped restart, proper volumes and ports.

---

## Phase 4: User Story 1 - Remote Engineer Securely Accesses Home LAN via VPN (Priority: P1)

**Goal**: VPN deployment works correctly without any Nomad dependencies. This story was already Nomad-free in 001 (WireGuard runs on the host via Ansible), but we verify provisioning still works after removing the Nomad role from the provision playbook.

**Independent Test**: Run `iac provision host --name <host> --ip <ip>`, then `iac deploy vpn --host <host>`. Verify WireGuard is running and VPN client can connect.

### Verification

- [X] T041 [US1] Verify `iac/ansible/playbooks/provision-host.yml` no longer references Nomad role and still includes all other roles (hardening, hypervisor, nat-network, vm-create, vm-guest, wireguard)
- [X] T042 [US1] Verify `iac/cli/cmd/deploy_vpn.go` has no Nomad references (should already be clean)
- [X] T043 [US1] Verify provision.go workflow outputs correct status messages without Nomad references

**Checkpoint**: Host provisioning and VPN deployment work end-to-end without Nomad. WireGuard is deployed on the designated host.

---

## Phase 5: User Story 5 - Single Source of Truth Configuration (Priority: P1)

**Goal**: The `homelab.yml` schema is updated (no `cluster.datacenter`, no `auto_update.auto_revert`/`health_check_timeout`), and `iac generate` produces correct Ansible inventory and Terraform tfvars without Nomad artifacts.

**Independent Test**: Edit `homelab.yml`, run `iac generate`, verify `.generated/ansible/` and `.generated/terraform/` contain correct configs and no `.generated/nomad/` directory is produced.

### Verification

- [X] T044 [US5] Update homelab.yml example at iac/cli/homelab.yml to remove `cluster.datacenter` field and `auto_update.auto_revert`/`health_check_timeout` fields
- [X] T045 [US5] Verify `iac generate` does not produce `.generated/nomad/` output (remove any code that creates this directory)

**Checkpoint**: Single config file generates all tool-specific configs correctly without any Nomad artifacts.

---

## Phase 6: User Story 6 - Simplified Architecture Uses Fewer Resources (Priority: P1)

**Goal**: Status command shows Docker container status instead of Nomad job status. No scheduler process runs. Infrastructure overhead is reduced.

**Independent Test**: Run `iac status` and verify it shows host table, Docker container table, and WireGuard status -- no Nomad section.

### Implementation

- [X] T046 [US6] Implement `printContainerStatus()` in iac/cli/cmd/status.go to query Docker container status via SSH (`docker ps --format` on each VM)
- [X] T047 [US6] Update `Status()` in iac/cli/cmd/status.go to call `printContainerStatus()` and display container table (name, image, status, ports, restart policy)

**Checkpoint**: `iac status` shows Docker containers instead of Nomad jobs. No scheduler overhead.

---

## Phase 7: User Story 4 - Media Applications Automatically Update (Priority: P2)

**Goal**: A systemd timer inside each VM runs a shell script on a configurable schedule. The script iterates over configured containers, pulls latest images, compares digests, and recreates containers when a new image is available. Per-container run scripts (templated by Ansible) ensure consistent recreation.

**Independent Test**: Deploy an app, verify the systemd timer is active, manually trigger the timer, and verify the update script runs and logs output to journald.

### Ansible Role

- [X] T048 [P] [US4] Create auto-updater role defaults at iac/ansible/roles/auto-updater/defaults/main.yml with schedule, container list, script path defaults
- [X] T049 [US4] Create update script template at iac/ansible/roles/auto-updater/templates/update-containers.sh.j2 that iterates containers, pulls images, compares digests, recreates via run scripts
- [X] T050 [P] [US4] Create systemd service unit template at iac/ansible/roles/auto-updater/templates/homelab-updater.service.j2
- [X] T051 [P] [US4] Create systemd timer unit template at iac/ansible/roles/auto-updater/templates/homelab-updater.timer.j2 with Persistent=true
- [X] T052 [US4] Create auto-updater role tasks at iac/ansible/roles/auto-updater/tasks/main.yml to template script and units, enable timer

### Ansible Playbook

- [X] T053 [US4] Create deploy-updater playbook at iac/ansible/playbooks/deploy-updater.yml targeting VM hosts, applying auto-updater role

### Go CLI Updates

- [X] T054 [US4] Update `deployApp()` in iac/cli/cmd/deploy_app.go to also run deploy-updater.yml when `auto_update.enabled` is true
- [X] T055 [US4] Implement `updateStatus()` in iac/cli/cmd/update_status.go to query systemd timer status and recent journald logs via SSH
- [X] T056 [US4] Simplify `Update()` routing in iac/cli/cmd/update.go to only support `status` subcommand
- [X] T057 [US4] Update `main.go` routing to remove `update trigger`, `update enable`, `update disable` paths if still present

### Build Verification

- [X] T058 [US4] Verify Go CLI compiles cleanly with `cd iac/cli && go build ./...`

**Checkpoint**: Auto-updater systemd timer is deployed to VMs. `iac update status` shows timer status and last run. Containers auto-update on schedule.

---

## Phase 8: User Story 3 - Defense-in-Depth Security (Priority: P2)

**Goal**: Security audit playbook is updated to verify Docker container security (read-only root, restart policy, no privileged mode) instead of Nomad job security. All six security layers are validated.

**Independent Test**: Run `iac audit` and verify it checks all security layers including container-specific checks (read-only root, non-privileged, restart policy).

### Implementation

- [X] T059 [US3] Update security audit playbook at iac/ansible/playbooks/security-audit.yml to remove Nomad port 4646 from expected VM ports and add Docker container security checks (read-only root, restart policy, no privileged flag)
- [X] T060 [US3] Verify audit.go in iac/cli/cmd/audit.go has no Nomad references

**Checkpoint**: Security audit validates all six defense-in-depth layers without Nomad. Container security posture is verified.

---

## Phase 9: Polish & Cross-Cutting Concerns

**Purpose**: Final cleanup, documentation, and validation

- [X] T061 [P] Update README.md to remove Nomad from prerequisites, architecture diagram, and descriptions
- [X] T062 [P] Grep entire codebase for remaining "nomad" or "Nomad" references and clean up any stragglers
- [X] T063 Final Go build verification with `cd iac/cli && go build ./...`
- [X] T064 Verify `.gitignore` does not reference `.generated/nomad/` (clean up if present)
- [X] T065 Run quickstart.md validation: verify all CLI commands referenced in quickstart.md are implemented

---

## Dependencies & Execution Order

### Phase Dependencies

- **Setup (Phase 1)**: No dependencies - start immediately
- **Foundational (Phase 2)**: Depends on Setup - BLOCKS all user stories
- **US2 (Phase 3)**: Depends on Foundational phase completion
- **US1 (Phase 4)**: Depends on Foundational phase completion - can run in parallel with US2
- **US5 (Phase 5)**: Depends on Foundational phase completion - can run in parallel with US2, US1
- **US6 (Phase 6)**: Depends on Foundational phase completion - can run in parallel with US2, US1, US5
- **US4 (Phase 7)**: Depends on US2 (needs app-container role's run scripts to exist)
- **US3 (Phase 8)**: Depends on US2 (needs containers deployed to audit)
- **Polish (Phase 9)**: Depends on all user stories being complete

### User Story Dependencies

- **US2 (Deploy Apps) P1**: Can start after Foundational (Phase 2) - No dependencies on other stories
- **US1 (VPN Access) P1**: Can start after Foundational (Phase 2) - No dependencies on other stories
- **US5 (Single Config) P1**: Mostly achieved in Foundational phase. Phase 5 is verification only.
- **US6 (Fewer Resources) P1**: Can start after Foundational (Phase 2) - Needs status.go updates.
- **US4 (Auto-Updates) P2**: Depends on US2 (uses per-container run scripts from app-container role)
- **US3 (Security) P2**: Depends on US2 (needs deployed containers to audit)

### Within Each User Story

- Ansible roles before playbooks
- Playbooks before Go CLI updates
- Defaults before tasks
- Templates before tasks that reference them
- All implementation before build verification

### Parallel Opportunities

- T003-T009 (delete Nomad files) can all run in parallel
- T010-T013 (config updates) can run in parallel
- T014-T016 (generator updates) can run in parallel after config changes
- T029-T037 (Ansible roles + playbooks for US2) - roles can be built in parallel, playbooks in parallel
- T048-T053 (auto-updater role + playbook) - defaults, service, timer templates in parallel
- T061-T062 (polish) can run in parallel
- US1, US5, US6 can all run in parallel with US2 (after foundational phase)

---

## Parallel Example: Phase 2 (Delete Nomad Files)

```bash
# These 7 delete tasks can all run in parallel (independent files):
Task T003: "Delete iac/ansible/roles/nomad/"
Task T004: "Delete iac/nomad/"
Task T005: "Delete iac/templates/nomad-*.hcl.tmpl"
Task T006: "Delete iac/terraform/nomad-jobs.tf"
Task T007: "Delete iac/cli/generators/nomad.go"
Task T008: "Delete iac/cli/updater/updater.go"
Task T009: "Delete iac/cli/cmd/update_trigger.go and update_toggle.go"
```

## Parallel Example: Phase 3 (US2 Ansible Roles)

```bash
# App container and proxy container roles can be built in parallel:
Task T029: "Create app-container role defaults"
Task T032: "Create proxy-container role defaults"

# After defaults, tasks and templates can proceed:
Task T030: "Create app-container role tasks"
Task T033: "Create proxy-container role tasks"

# Playbooks can be created in parallel:
Task T036: "Create deploy-app playbook"
Task T037: "Create deploy-proxy playbook"
```

---

## Implementation Strategy

### MVP First (US2 Only)

1. Complete Phase 1: Setup
2. Complete Phase 2: Foundational - Remove Nomad (CRITICAL - blocks all stories)
3. Complete Phase 3: US2 - Deploy Media Apps
4. **STOP and VALIDATE**: Verify `iac deploy app --name plex` creates a running container
5. The system can deploy apps at this point

### Incremental Delivery

1. Complete Setup + Foundational -> Nomad removed, code compiles
2. Add US2 (Deploy Apps) -> Apps deploy via Ansible -> **MVP!**
3. Add US1 (VPN) verification -> VPN confirmed working
4. Add US5 (Config) verification -> Config pipeline confirmed clean
5. Add US6 (Status) -> Status shows Docker containers
6. Add US4 (Auto-Updates) -> Autonomous updates working
7. Add US3 (Security) -> Audit validates all layers
8. Polish -> README updated, codebase clean

---

## Notes

- [P] tasks = different files, no dependencies
- [Story] label maps task to specific user story for traceability
- This is a refactoring feature: most work is removing/replacing existing code, not building from scratch
- US5 and US6 are largely achieved through the foundational Nomad removal (Phase 2)
- Rollback is explicitly deferred - do not implement health checks or auto-revert
- All containers use `restart_policy: unless-stopped` for crash/reboot recovery
- Per-container run scripts ensure the auto-updater recreates containers with the same config as Ansible
- Total: 65 tasks across 9 phases
