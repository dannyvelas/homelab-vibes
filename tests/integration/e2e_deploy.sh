#!/bin/bash
# End-to-end deployment test script
# Provisions hosts, deploys VPN, deploys media stack, runs security audit
#
# Prerequisites:
# - homelab.yml configured at repo root
# - SSH access to target hosts
# - iac CLI built (go build -o iac ./iac/cli/)
#
# Usage: ./tests/integration/e2e_deploy.sh [--config path/to/homelab.yml]

set -euo pipefail

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m'

CONFIG="${1:-homelab.yml}"
IAC="./iac"
PASS=0
FAIL=0

log_step() {
  echo -e "\n${YELLOW}==== $1 ====${NC}\n"
}

log_pass() {
  echo -e "  ${GREEN}PASS${NC}: $1"
  PASS=$((PASS + 1))
}

log_fail() {
  echo -e "  ${RED}FAIL${NC}: $1"
  FAIL=$((FAIL + 1))
}

check_cmd() {
  local desc="$1"
  shift
  if "$@" > /dev/null 2>&1; then
    log_pass "$desc"
  else
    log_fail "$desc"
  fi
}

# ----------------------------------------------------------
# 0. Pre-flight checks
# ----------------------------------------------------------
log_step "Pre-flight checks"

if [ ! -f "$CONFIG" ]; then
  echo "Error: Config file not found: $CONFIG"
  exit 1
fi
log_pass "Config file exists: $CONFIG"

if [ ! -f "$IAC" ]; then
  echo "Building iac CLI..."
  (cd iac/cli && go build -o ../../iac .) || {
    log_fail "Failed to build iac CLI"
    exit 1
  }
fi
log_pass "iac CLI binary exists"

check_cmd "terraform is installed" terraform version
check_cmd "ansible-playbook is installed" ansible-playbook --version

# ----------------------------------------------------------
# 1. Config validation
# ----------------------------------------------------------
log_step "Step 1: Config validation"

# The CLI will validate the config when loading
check_cmd "iac status loads config" $IAC status --config "$CONFIG"

# ----------------------------------------------------------
# 2. Generate configs
# ----------------------------------------------------------
log_step "Step 2: Generate tool-specific configs"

check_cmd "iac generate configs" $IAC generate configs --config "$CONFIG"

# Verify generated files exist
check_cmd "Ansible inventory generated" test -f .generated/ansible/inventory/hosts.yml
check_cmd "Terraform tfvars directory generated" test -d .generated/terraform/
check_cmd "group_vars generated" test -f .generated/ansible/inventory/group_vars/all.yml

# ----------------------------------------------------------
# 3. Provision hosts
# ----------------------------------------------------------
log_step "Step 3: Provision hosts"

# Extract host names from config (simple grep approach)
HOSTS=$(grep -E '^\s+\w+.*:$' "$CONFIG" | head -5 | sed 's/://;s/^ *//')

for HOST in $HOSTS; do
  echo "  Provisioning host: $HOST"
  if $IAC provision host --name "$HOST" --config "$CONFIG" 2>&1; then
    log_pass "Provisioned host: $HOST"
  else
    log_fail "Failed to provision host: $HOST"
  fi
done

# ----------------------------------------------------------
# 4. Deploy VPN
# ----------------------------------------------------------
log_step "Step 4: Deploy VPN"

if $IAC deploy vpn --config "$CONFIG" 2>&1; then
  log_pass "VPN deployed"
else
  log_fail "VPN deployment failed"
fi

# Generate a test VPN client
if $IAC generate vpn-client --name e2e-test --config "$CONFIG" 2>&1; then
  log_pass "VPN client config generated"
  check_cmd "VPN client config file exists" test -f .generated/vpn-clients/e2e-test.conf
else
  log_fail "VPN client config generation failed"
fi

# ----------------------------------------------------------
# 5. Deploy reverse proxy
# ----------------------------------------------------------
log_step "Step 5: Deploy reverse proxy"

if $IAC deploy proxy --config "$CONFIG" 2>&1; then
  log_pass "Reverse proxy deployed"
else
  log_fail "Reverse proxy deployment failed"
fi

# ----------------------------------------------------------
# 6. Deploy media apps
# ----------------------------------------------------------
log_step "Step 6: Deploy media stack"

for APP in plex sonarr radarr bazarr; do
  echo "  Deploying $APP..."
  if $IAC deploy app --name "$APP" --config "$CONFIG" 2>&1; then
    log_pass "Deployed: $APP"
  else
    log_fail "Failed to deploy: $APP"
  fi
done

# ----------------------------------------------------------
# 7. Verify services are accessible
# ----------------------------------------------------------
log_step "Step 7: Verify services accessible"

# Give containers a moment to start
echo "  Waiting 30s for services to start..."
sleep 30

# Check container status
$IAC status --config "$CONFIG" 2>&1 && log_pass "iac status shows cluster info" || log_fail "iac status failed"

# ----------------------------------------------------------
# 8. Security audit
# ----------------------------------------------------------
log_step "Step 8: Security audit"

if $IAC audit security --config "$CONFIG" 2>&1; then
  log_pass "Security audit passed"
else
  log_fail "Security audit found issues"
fi

# ----------------------------------------------------------
# Summary
# ----------------------------------------------------------
log_step "E2E Test Summary"

TOTAL=$((PASS + FAIL))
echo -e "  Total checks: $TOTAL"
echo -e "  ${GREEN}Passed: $PASS${NC}"
echo -e "  ${RED}Failed: $FAIL${NC}"

if [ $FAIL -eq 0 ]; then
  echo -e "\n${GREEN}ALL E2E TESTS PASSED${NC}\n"
  exit 0
else
  echo -e "\n${RED}$FAIL E2E TESTS FAILED${NC}\n"
  exit 1
fi
