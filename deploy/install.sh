#!/usr/bin/env bg
# Custodian PaaS Automated Server Installer
# Usage: curl -fsSL https://raw.githubusercontent.com/CocoCopi/custodian/main/deploy/install.sh | sudo bash

set -euo pipefail

# Visual colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
CYAN='\033[0;36m'
BOLD='\033[1m'
NC='\033[0m' # No Color

log_info() { echo -e "${BLUE}[INFO]${NC} $1"; }
log_success() { echo -e "${GREEN}[SUCCESS]${NC} $1"; }
log_warn() { echo -e "${YELLOW}[WARN]${NC} $1"; }
log_error() { echo -e "${RED}[ERROR]${NC} $1"; }

echo -e "${CYAN}${BOLD}"
echo "  CUSTODIAN PAAS INSTALLER"
echo "  The self-hosted control plane for your infrastructure."
echo -e "${NC}"

# 1. Require root or sudo privileges
if [ "$EUID" -ne 0 ]; then
  log_error "Please run installer as root or with sudo:"
  echo -e "  ${BOLD}curl -fsSL https://raw.githubusercontent.com/CocoCopi/custodian/main/deploy/install.sh | sudo bash${NC}"
  exit 1
fi

# 2. Detect Server IP address safely
SERVER_IP=$(curl -s --max-time 5 https://api.ipify.org 2>/dev/null | tr -d '\r\n' | awk '{print $1}')
if [ -z "${SERVER_IP}" ]; then
  SERVER_IP=$(hostname -I 2>/dev/null | awk '{print $1}')
fi
if [ -z "${SERVER_IP}" ]; then
  SERVER_IP="localhost"
fi
log_info "Detected server public IP: ${BOLD}${SERVER_IP}${NC}"

# 3. Check and install prerequisite packages (curl, git, openssl)
log_info "Checking prerequisite tools..."
if ! command -v curl &>/dev/null || ! command -v git &>/dev/null || ! command -v openssl &>/dev/null; then
  log_info "Installing core utility packages (curl, git, openssl)..."
  if command -v apt-get &>/dev/null; then
    apt-get update -qq && apt-get install -y -qq curl git openssl ca-certificates >/dev/null
  elif command -v dnf &>/dev/null; then
    dnf install -y -q curl git openssl ca-certificates >/dev/null
  elif command -v yum &>/dev/null; then
    yum install -y -q curl git openssl ca-certificates >/dev/null
  elif command -v pacman &>/dev/null; then
    pacman -Sy --noconfirm curl git openssl ca-certificates >/dev/null
  fi
fi

# 4. Check and install Docker & Docker Compose
if ! command -v docker &>/dev/null || ! docker compose version &>/dev/null; then
  log_info "Docker or Docker Compose not found. Installing Docker engine automatically..."
  curl -fsSL https://get.docker.com | sh
  log_success "Docker installed successfully."
else
  log_info "Docker and Docker Compose plugin are already installed."
fi

# Enable and start Docker service
log_info "Enabling Docker service autostart..."
systemctl enable --now docker >/dev/null 2>&1 || true

# 5. Setup Installation Directory & Fetch Repository
INSTALL_DIR="/opt/custodian"

# If current working directory is already the custodian repository, use it directly
if [ -f "./deploy/docker-compose.yml" ] && [ -f "./go.mod" ]; then
  INSTALL_DIR="$(pwd)"
  log_info "Using existing local Custodian repository at ${BOLD}${INSTALL_DIR}${NC}"
else
  mkdir -p "${INSTALL_DIR}"
  if [ ! -f "${INSTALL_DIR}/deploy/docker-compose.yml" ]; then
    log_info "Fetching Custodian control plane repository into ${BOLD}${INSTALL_DIR}${NC}..."
    if command -v git &>/dev/null; then
      git clone --depth 1 https://github.com/CocoCopi/custodian.git "${INSTALL_DIR}" || {
        log_warn "Git clone failed, falling back to repository tarball download..."
        curl -fsSL https://github.com/CocoCopi/custodian/archive/refs/heads/main.tar.gz | tar -xz -C "${INSTALL_DIR}" --strip-components=1
      }
    else
      log_info "Downloading Custodian repository archive..."
      curl -fsSL https://github.com/CocoCopi/custodian/archive/refs/heads/main.tar.gz | tar -xz -C "${INSTALL_DIR}" --strip-components=1
    fi
    log_success "Repository successfully fetched into ${INSTALL_DIR}"
  else
    log_info "Repository already present at ${INSTALL_DIR}."
    if [ -d "${INSTALL_DIR}/.git" ] && command -v git &>/dev/null; then
      log_info "Updating Custodian codebase..."
      git -C "${INSTALL_DIR}" pull --rebase || true
    fi
  fi
fi

cd "${INSTALL_DIR}"

# Port Conflict Resolution Helpers
is_port_in_use() {
  local port=$1
  if command -v ss &>/dev/null; then
    ss -tuln | grep -q ":${port} "
  elif command -v netstat &>/dev/null; then
    netstat -tuln | grep -q ":${port} "
  elif command -v lsof &>/dev/null; then
    lsof -i :${port} &>/dev/null
  else
    (echo > /dev/tcp/127.0.0.1/${port}) &>/dev/null 2>&1
  fi
}

find_free_port() {
  local port=$1
  while is_port_in_use "${port}"; do
    log_warn "Port ${port} is in use, checking next available port..."
    port=$((port + 1))
  done
  echo "${port}"
}

# 6. Setup `.env` configuration file & resolve dynamic ports
ENV_FILE="${INSTALL_DIR}/deploy/.env"

if [ ! -f "${ENV_FILE}" ]; then
  log_info "Generating deployment .env configuration with security tokens..."

  JWT_SECRET=$(openssl rand -hex 32)
  DB_PASS=$(openssl rand -hex 16)
  MINIO_PASS=$(openssl rand -hex 16)
  GRAFANA_PASS=$(openssl rand -hex 16)

  # Conflict-free port resolution
  API_PORT=$(find_free_port 8080)
  HTTP_PORT=$(find_free_port 80)
  HTTPS_PORT=$(find_free_port 443)

  cat <<EOF > "${ENV_FILE}"
CUSTODIAN_DOMAIN=${SERVER_IP}

CUSTODIAN_PORT=${API_PORT}
CUSTODIAN_HTTP_PORT=${HTTP_PORT}
CUSTODIAN_HTTPS_PORT=${HTTPS_PORT}

CUSTODIAN_JWT_SECRET=${JWT_SECRET}

CUSTODIAN_DB_USER=custodian
CUSTODIAN_DB_PASSWORD=${DB_PASS}
CUSTODIAN_DB_NAME=custodian

CUSTODIAN_MINIO_USER=custodian
CUSTODIAN_MINIO_PASSWORD=${MINIO_PASS}

CUSTODIAN_GRAFANA_USER=admin
CUSTODIAN_GRAFANA_PASSWORD=${GRAFANA_PASS}

CUSTODIAN_LETSENCRYPT_EMAIL=admin@example.com
CUSTODIAN_ENGINE=compose
EOF

  log_success "Created deploy/.env (API Port: ${API_PORT}, Web Port: ${HTTP_PORT})."
else
  log_info "Existing deploy/.env found, preserving configuration."
fi

# Load active ports for display
UI_PORT=$(grep "^CUSTODIAN_HTTP_PORT=" "${ENV_FILE}" | cut -d'=' -f2 || echo "80")
API_PORT=$(grep "^CUSTODIAN_PORT=" "${ENV_FILE}" | cut -d'=' -f2 || echo "8080")

# 7. Configure Systemd Autostart Service
SERVICE_FILE="/etc/systemd/system/custodian.service"
log_info "Configuring systemd service for automatic server boot autostart..."

cat <<EOF > "${SERVICE_FILE}"
[Unit]
Description=Custodian PaaS Control Plane Service
Requires=docker.service
After=docker.service network.target

[Service]
Type=oneshot
RemainAfterExit=yes
WorkingDirectory=${INSTALL_DIR}
ExecStart=/usr/bin/docker compose -f ${INSTALL_DIR}/deploy/docker-compose.yml --env-file ${INSTALL_DIR}/deploy/.env up -d
ExecStop=/usr/bin/docker compose -f ${INSTALL_DIR}/deploy/docker-compose.yml --env-file ${INSTALL_DIR}/deploy/.env down
TimeoutStartSec=0

[Install]
WantedBy=multi-user.target
EOF

systemctl daemon-reload
systemctl enable custodian.service
log_success "Systemd service custodian.service enabled."

# 8. Start Control Plane Stack
log_info "Launching Custodian PaaS container stack..."
docker compose -f "${INSTALL_DIR}/deploy/docker-compose.yml" --env-file "${ENV_FILE}" up -d --build

# Form UI URL
if [ "${UI_PORT}" = "80" ]; then
  UI_URL="http://${SERVER_IP}"
else
  UI_URL="http://${SERVER_IP}:${UI_PORT}"
fi

# 9. Output Completion Summary
echo ""
echo -e "${GREEN}${BOLD}================================================================"
echo -e "  CUSTODIAN PAAS INSTALLED AND OPERATIONAL"
echo -e "================================================================${NC}"
echo ""
echo -e "  Web Console:        ${UI_URL}"
echo -e "  API Endpoint:       http://${SERVER_IP}:${API_PORT}"
echo -e "  Grafana Metrics:    http://${SERVER_IP}:3000"
echo -e "  MinIO S3 Console:   http://${SERVER_IP}:9001"
echo ""
echo -e "  Next Steps:"
echo -e "     1. Open ${BOLD}${UI_URL}${NC} in your browser to set up your admin account."
echo -e "     2. Generate an API token under the API Tokens tab for CLI access."
echo -e "     3. Build the CLI tool locally: ${CYAN}go build -o bin/custodian ./cmd/custodian-cli${NC}"
echo ""
echo -e "  Installed path: ${BOLD}${INSTALL_DIR}${NC}"
echo -e "  Systemd status: ${CYAN}systemctl status custodian.service${NC}"
echo -e "${GREEN}================================================================"
