#!/usr/bin/env bash
set -euo pipefail

ROOT="$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")/.." >/dev/null && pwd -P)"
DIST="${ROOT}/dist"
SRC_TOOL="${ROOT}/src/tool"
VERSION="$(tr -d '[:space:]' < "${ROOT}/VERSION")"
MAINTAINER="${PACKAGE_MAINTAINER:-Muhammad Naufal Hanif <naufalmng@gmail.com>}"
HOMEPAGE="${PACKAGE_HOMEPAGE:-https://github.com/naufalmng/gc-rg}"

step() { printf '==> %s\n' "$*"; }
done_() { printf ' ok %s\n' "$*"; }

concat_modules() {
  local dir="${1:?missing dir}"
  local first=1
  local file=""
  for file in "${dir}"/*.sh; do
    [[ -f "$file" ]] || continue
    if (( first == 1 )); then
      cat "$file"
      first=0
    else
      printf '\n# ===== %s =====\n' "$(basename "$file")"
      awk 'NR==1 && /^#!/ {next} {print}' "$file"
    fi
  done
}

write_installer() {
  local target="${1:?missing target}"
  cat > "$target" <<'INSTALLER'
#!/usr/bin/env bash
set -Eeuo pipefail

PACKAGE_NAME="gc-rg"
PACKAGE_VERSION="__PACKAGE_VERSION__"
PACKAGE_ARCH="amd64"
PACKAGE_MAINTAINER="__PACKAGE_MAINTAINER__"
PACKAGE_HOMEPAGE="__PACKAGE_HOMEPAGE__"
SYSTEMD_DIR="/lib/systemd/system"
CONFIG_DIR="/etc/gc-rg"
APP_DIR="/opt/gc-rg"
MAIN_GENERATE="/opt/gc-rg/bin/gc-rg-generate"
MAIN_EMAIL="/opt/gc-rg/bin/gc-rg-email"
MAIN_TOOL="/usr/bin/gc-rg"
SHORT_TOOL="/usr/bin/gcrg"
ACTION=""
YES="false"
FORCE="false"
KEEP_DEB="false"
NO_COLOR="${NO_COLOR:-}"
TMP_BUILD_DIR=""

if [[ -t 1 && -z "$NO_COLOR" ]]; then
  C_BOLD=$'\033[1m'; C_GREEN=$'\033[32m'; C_BLUE=$'\033[34m'; C_DIM=$'\033[2m'; C_RESET=$'\033[0m'
else
  C_BOLD=''; C_GREEN=''; C_BLUE=''; C_DIM=''; C_RESET=''
fi

info() { printf '%s\n' "$*"; }
ok() { printf '%s✓%s %s\n' "$C_GREEN" "$C_RESET" "$*"; }
step() { printf '%s==>%s %s\n' "$C_BLUE" "$C_RESET" "$*"; }
die() { printf 'ERROR: %s\n' "$*" >&2; exit 1; }

cleanup() {
  if [[ -n "$TMP_BUILD_DIR" && -d "$TMP_BUILD_DIR" ]]; then
    rm -rf "$TMP_BUILD_DIR"
  fi
}
trap cleanup EXIT

usage() {
  cat <<EOF
${C_BOLD}gc-rg installer${C_RESET}  ${C_DIM}v${PACKAGE_VERSION}${C_RESET}

${C_BOLD}Usage:${C_RESET}
  sudo bash ${0##*/} install
  sudo bash ${0##*/} uninstall
  bash ${0##*/} standalone

${C_BOLD}Options:${C_RESET}
  -y, --yes          assume yes
  -f, --force        overwrite existing standalone files
      --keep-deb     keep generated .deb in current directory
      --no-color     disable ANSI colors
  -h, --help         show help

${C_BOLD}Pipe examples:${C_RESET}
  curl -fsSL ${PACKAGE_HOMEPAGE}/releases/latest/download/gc-rg.sh | sudo bash
  curl -fsSL ${PACKAGE_HOMEPAGE}/releases/latest/download/gc-rg.sh | sudo bash -s -- install --yes
  curl -fsSL ${PACKAGE_HOMEPAGE}/releases/latest/download/gc-rg.sh | bash -s -- standalone

${C_BOLD}After install:${C_RESET}
  sudo gc-rg onboard
  gcrg generate
  gcrg send --dry-run
  gc-rg run
  gcrg status
  sudo apt-get remove gc-rg

${C_BOLD}Unified commands:${C_RESET}
  gc-rg generate
  gc-rg send
  gc-rg run
EOF
}

parse_args() {
  while (( $# > 0 )); do
    case "$1" in
      install|uninstall|remove|standalone|help) ACTION="$1"; shift ;;
      --install) ACTION="install"; shift ;;
      --uninstall|--remove) ACTION="uninstall"; shift ;;
      --standalone) ACTION="standalone"; shift ;;
      -y|--yes) YES="true"; shift ;;
      -f|--force) FORCE="true"; shift ;;
      --keep-deb) KEEP_DEB="true"; shift ;;
      --no-color) NO_COLOR="1"; shift ;;
      -h|--help) ACTION="help"; shift ;;
      *) die "unknown argument: $1" ;;
    esac
  done
  [[ -n "$ACTION" ]] || ACTION="install"
  if [[ "$ACTION" == "remove" ]]; then
    ACTION="uninstall"
  fi
}

need_root() {
  [[ "${EUID:-$(id -u)}" -eq 0 ]] || die "run as root: sudo bash ${0##*/} install"
}

need_cmd() {
  local cmd=""
  for cmd in "$@"; do
    command -v "$cmd" >/dev/null 2>&1 || die "missing command: $cmd"
  done
}

confirm() {
  local question="${1:?missing question}"
  local answer=""
  [[ "$YES" == "true" ]] && return 0
  printf '%s [Y/n] ' "$question" > /dev/tty
  read -r answer < /dev/tty || return 1
  [[ -z "$answer" || "$answer" =~ ^[Yy]$ ]]
}

asset_url() {
  printf '%s/releases/latest/download/%s\n' "$PACKAGE_HOMEPAGE" "$1"
}

download_asset() {
  local name="${1:?missing name}"
  local target="${2:?missing target}"
  local url="$(asset_url "$name")"
  step "Downloading $name"
  curl -fsSL "$url" -o "$target" || die "download failed: $url"
  chmod 0755 "$target"
}

write_config_example() {
  local target="${1:?missing target}"
  cat > "$target" <<'EOF'
# GC-RG runtime config
GC_RG_EMAIL_PROVIDER=gmail
GC_RG_SMTP_HOST=
GC_RG_SMTP_PORT=587
GC_RG_SMTP_TLS=starttls
GC_RG_SMTP_AUTH=on
GC_RG_SMTP_USERNAME=your-email@gmail.com
GC_RG_SMTP_PASSWORD=replace-with-app-password
GC_RG_EMAIL_FROM=your-email@gmail.com
GC_RG_EMAIL_TO=ops@example.com,manager@example.com
GC_RG_EMAIL_CC=
GC_RG_EMAIL_SUBJECT_PREFIX=[GC-RG]
GC_RG_REPORT_DIR=/opt/gc-rg/reports/daily
GC_RG_WORKDIR=/opt/gc-rg
EOF
}

write_service_file() {
  local target="${1:?missing target}"
  cat > "$target" <<'EOF'
[Unit]
Description=Grafana Cloud Report Generator daily email
Wants=network-online.target
After=network-online.target

[Service]
Type=oneshot
WorkingDirectory=/opt/gc-rg
EnvironmentFile=/etc/gc-rg/gc-rg.env
ExecStart=/usr/bin/gc-rg run --quiet
User=gc-rg
Group=gc-rg
NoNewPrivileges=true
PrivateTmp=true
ProtectSystem=full
ProtectHome=true
ReadWritePaths=/opt/gc-rg/reports /opt/gc-rg/tmp
EOF
}

write_timer_file() {
  local target="${1:?missing target}"
  cat > "$target" <<'EOF'
[Unit]
Description=Run Grafana Cloud Report Generator daily

[Timer]
OnCalendar=*-*-* 08:00:00
Persistent=true
RandomizedDelaySec=5m
Unit=gc-rg.service

[Install]
WantedBy=timers.target
EOF
}

write_maintainer_scripts() {
  local debian_dir="${1:?missing debian dir}"
  cat > "${debian_dir}/postinst" <<'EOF'
#!/usr/bin/env bash
set -e
getent group gc-rg >/dev/null || groupadd --system gc-rg
getent passwd gc-rg >/dev/null || useradd --system --gid gc-rg --home-dir /opt/gc-rg --shell /usr/sbin/nologin gc-rg
mkdir -p /opt/gc-rg/evidence /opt/gc-rg/reports/daily /opt/gc-rg/tmp /etc/gc-rg
if [ ! -f /etc/gc-rg/gc-rg.env ]; then
  cp /usr/share/gc-rg/gc-rg.env.example /etc/gc-rg/gc-rg.env
  chmod 0600 /etc/gc-rg/gc-rg.env
fi
chown -R gc-rg:gc-rg /opt/gc-rg
systemctl daemon-reload >/dev/null 2>&1 || true
EOF
  cat > "${debian_dir}/prerm" <<'EOF'
#!/usr/bin/env bash
set -e
if [ "$1" = "remove" ] || [ "$1" = "deconfigure" ]; then
  systemctl disable --now gc-rg.timer >/dev/null 2>&1 || true
fi
EOF
  cat > "${debian_dir}/postrm" <<'EOF'
#!/usr/bin/env bash
set -e
systemctl daemon-reload >/dev/null 2>&1 || true
EOF
  chmod 0755 "${debian_dir}/postinst" "${debian_dir}/prerm" "${debian_dir}/postrm"
}

build_deb() {
  local pkg_dir="" debian_dir="" deb_path=""
  need_cmd curl dpkg-deb install chmod chown mktemp
  TMP_BUILD_DIR="$(mktemp -d -p /var/tmp "${PACKAGE_NAME}.XXXXXX")"
  chmod 0755 "$TMP_BUILD_DIR"
  pkg_dir="${TMP_BUILD_DIR}/${PACKAGE_NAME}_${PACKAGE_VERSION}_${PACKAGE_ARCH}"
  debian_dir="${pkg_dir}/DEBIAN"
  deb_path="${TMP_BUILD_DIR}/${PACKAGE_NAME}_${PACKAGE_VERSION}_${PACKAGE_ARCH}.deb"

  install -d -m 0755 "$debian_dir" "${pkg_dir}/opt/gc-rg/bin" "${pkg_dir}/usr/bin" "${pkg_dir}/usr/share/gc-rg" "${pkg_dir}${SYSTEMD_DIR}"
  install -d -m 0750 "${pkg_dir}/etc/gc-rg" "${pkg_dir}/opt/gc-rg/evidence" "${pkg_dir}/opt/gc-rg/reports/daily" "${pkg_dir}/opt/gc-rg/tmp"

  download_asset "gc-rg-generate-linux-amd64" "${pkg_dir}${MAIN_GENERATE}"
  download_asset "gc-rg-email-linux-amd64" "${pkg_dir}${MAIN_EMAIL}"
  download_asset "gc-rg" "${pkg_dir}${MAIN_TOOL}"
  ln -s gc-rg "${pkg_dir}${SHORT_TOOL}"
  write_config_example "${pkg_dir}/usr/share/gc-rg/gc-rg.env.example"
  write_service_file "${pkg_dir}${SYSTEMD_DIR}/gc-rg.service"
  write_timer_file "${pkg_dir}${SYSTEMD_DIR}/gc-rg.timer"
  write_maintainer_scripts "$debian_dir"

  cat > "${debian_dir}/control" <<EOF
Package: ${PACKAGE_NAME}
Version: ${PACKAGE_VERSION}
Section: admin
Priority: optional
Architecture: ${PACKAGE_ARCH}
Depends: ca-certificates, systemd, curl, wkhtmltopdf
Maintainer: ${PACKAGE_MAINTAINER}
Homepage: ${PACKAGE_HOMEPAGE}
Description: Grafana Cloud daily report generator with SMTP delivery
 A Go-based report generator that renders Grafana Cloud validation evidence
 into daily Markdown/PDF reports and sends them through operator-owned SMTP.
EOF
  dpkg-deb --build "$pkg_dir" "$deb_path" >/dev/null || die "failed to build deb package"
  chmod 0644 "$deb_path"
  printf '%s\n' "$deb_path"
}

install_package() {
  local deb_path="" keep_path=""
  need_root
  need_cmd apt-get dpkg-deb systemctl curl mktemp
  usage
  confirm "Install gc-rg package?" || { info "Abort."; return 0; }
  deb_path="$(build_deb)"
  apt-get install -y "$deb_path" || die "apt-get install failed"
  if [[ "$KEEP_DEB" == "true" ]]; then
    keep_path="${PWD}/${PACKAGE_NAME}_${PACKAGE_VERSION}_${PACKAGE_ARCH}.deb"
    cp -f "$deb_path" "$keep_path"
    chmod 0644 "$keep_path"
    ok "Debian package kept: $keep_path"
  fi
  ok "gc-rg installed successfully"
  cat <<EOF

  Main command:
    gc-rg help

  Short command:
    gcrg help

  Next step:
    sudo gc-rg onboard

  Remove:
    sudo apt-get remove gc-rg
EOF
}

uninstall_package() {
  need_root
  need_cmd apt-get
  confirm "Remove gc-rg package?" || { info "Abort."; return 0; }
  apt-get remove -y "$PACKAGE_NAME"
}

standalone() {
  local dir="${PWD}/gc-rg-standalone"
  need_cmd curl chmod mkdir
  if [[ -e "$dir" && "$FORCE" != "true" ]]; then
    die "$dir already exists; use standalone --force"
  fi
  confirm "Create standalone gc-rg in $dir?" || { info "Abort."; return 0; }
  mkdir -p "$dir/bin" "$dir/reports/daily" "$dir/evidence"
  download_asset "gc-rg-generate-linux-amd64" "$dir/bin/gc-rg-generate"
  download_asset "gc-rg-email-linux-amd64" "$dir/bin/gc-rg-email"
  download_asset "gc-rg" "$dir/gc-rg"
  ln -sf gc-rg "$dir/gcrg"
  write_config_example "$dir/gc-rg.env"
  ok "Standalone gc-rg created successfully: $dir"
}

main() {
  parse_args "$@"
  case "$ACTION" in
    install) install_package ;;
    uninstall) uninstall_package ;;
    standalone) standalone ;;
    help) usage ;;
    *) die "unknown action: $ACTION" ;;
  esac
}
main "$@"
INSTALLER
}

main() {
  step "preparing dist/"
  rm -rf "$DIST"
  mkdir -p "$DIST"

  step "building Go binaries for host"
  go build -o "${DIST}/gc-rg-generate" "${ROOT}/cmd/generate-daily-report"
  go build -o "${DIST}/gc-rg-email" "${ROOT}/cmd/send-email-report"

  step "building Linux release binaries"
  GOOS=linux GOARCH=amd64 go build -o "${DIST}/gc-rg-generate-linux-amd64" "${ROOT}/cmd/generate-daily-report"
  GOOS=linux GOARCH=amd64 go build -o "${DIST}/gc-rg-email-linux-amd64" "${ROOT}/cmd/send-email-report"

  step "assembling unified runtime"
  concat_modules "$SRC_TOOL" > "${DIST}/gc-rg"
  sed -i "s|__PACKAGE_VERSION__|${VERSION}|g" "${DIST}/gc-rg"
  chmod 0755 "${DIST}/gc-rg"

  step "building installer"
  write_installer "${DIST}/gc-rg.sh"
  sed -i "s|__PACKAGE_VERSION__|${VERSION}|g" "${DIST}/gc-rg.sh"
  sed -i "s|__PACKAGE_MAINTAINER__|${MAINTAINER}|g" "${DIST}/gc-rg.sh"
  sed -i "s|__PACKAGE_HOMEPAGE__|${HOMEPAGE}|g" "${DIST}/gc-rg.sh"
  chmod 0755 "${DIST}/gc-rg.sh"

  step "syntax-checking artifacts"
  bash -n "${DIST}/gc-rg.sh"
  bash -n "${DIST}/gc-rg"

  done_ "artifacts: dist/gc-rg.sh dist/gc-rg dist/gc-rg-generate dist/gc-rg-email"
}

main "$@"
