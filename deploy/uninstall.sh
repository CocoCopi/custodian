#!/usr/bin/env bash
# Custodian PaaS Automated Server Uninstaller
# Usage: curl -fsSL https://raw.githubusercontent.com/CocoCopi/custodian/main/deploy/uninstall.sh | sudo bash

set -euo pipefail

# Visual colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
CYAN='\033[0;36m'
BOLD='\033[1m'
NC='\033[0m' # No Color

log_info() { echo -e "${BLUE}[INFO]${NC} $1" >&2; }
log_success() { echo -e "${GREEN}[SUCCESS]${NC} $1" >&2; }
log_warn() { echo -e "${YELLOW}[WARN]${NC} $1" >&2; }
log_error() { echo -e "${RED}[ERROR]${NC} $1" >&2; }

echo -e "${CYAN}${BOLD}"
echo "  CUSTODIAN PAAS UNINSTALLER"
echo "  Removing Custodian PaaS control plane and services."
echo -e "${NC}"

# 1. Require root or sudo privileges
if [ "$EUID" -ne 0 ]; then
  log_error "Please run uninstaller as root or with sudo:"
  echo -e "  ${BOLD}curl -fsSL https://raw.githubusercontent.com/CocoCopi/custodian/main/deploy/uninstall.sh | sudo bash${NC}"
  exit 1
fi

INSTALL_DIR="/opt/custodian"

# If current directory is already a custodian repository, use it as fallback
if [ -f "./deploy/docker-compose.yml" ] && [ -f "./go.mod" ]; then
  INSTALL_DIR="$(pwd)"
fi

# 2. Stop and disable Systemd Service
SERVICE_FILE="/etc/systemd/system/custodian.service"
if [ -f "${SERVICE_FILE}" ]; then
  log_info "Stopping and disabling systemd custodian.service..."
  systemctl stop custodian.service 2>/dev/null || true
  systemctl disable custodian.service 2>/dev/null || true
  rm -f "${SERVICE_FILE}"
  systemctl daemon-reload
  log_success "Removed custodian.service systemd unit."
fi

# 3. Stop Docker Compose Stack
if [ -d "${INSTALL_DIR}" ] && [ -f "${INSTALL_DIR}/deploy/docker-compose.yml" ]; then
  log_info "Stopping and removing Custodian Docker containers, networks, and volumes..."
  if command -v docker &>/dev/null; then
    ENV_ARG=""
    if [ -f "${INSTALL_DIR}/deploy/.env" ]; then
      ENV_ARG="--env-file ${INSTALL_DIR}/deploy/.env"
    fi
    docker compose -f "${INSTALL_DIR}/deploy/docker-compose.yml" ${ENV_ARG} down -v --remove-orphans 2>/dev/null || true
  fi
fi

# 4. Remove Installation Directory
if [ -d "${INSTALL_DIR}" ]; then
  log_info "Removing installation directory ${INSTALL_DIR}..."
  rm -rf "${INSTALL_DIR}"
  log_success "Removed ${INSTALL_DIR}."
fi

echo ""
echo -e "${GREEN}${BOLD}================================================================"
echo -e "  CUSTODIAN PAAS HAS BEEN UNINSTALLED"
echo -e "================================================================${NC}"
echo ""
echo -e "  All Custodian control plane services, systemd units, and files"
echo -e "  have been completely removed from this server."
echo ""
echo -e "${GREEN}================================================================"
