# Shared utilities.

need_cmd() {
  local cmd=""
  for cmd in "$@"; do
    command -v "$cmd" >/dev/null 2>&1 || die "missing command: $cmd" || return 1
  done
}

need_root() {
  [[ "${EUID:-$(id -u)}" -eq 0 ]] || die "run as root"
}

source_config() {
  if [[ -s "$CONFIG_FILE" ]]; then
    set -a
    # shellcheck disable=SC1090
    . "$CONFIG_FILE"
    set +a
  fi
}

mask_value() {
  local value="${1:-}"
  if [[ -z "$value" ]]; then
    printf '<unset>\n'
  elif (( ${#value} <= 6 )); then
    printf '******\n'
  else
    printf '%s***%s\n' "${value:0:3}" "${value: -3}"
  fi
}

confirm() {
  local question="${1:?missing question}"
  local answer=""
  [[ "$YES" == "true" ]] && return 0
  if [[ ! -t 0 ]]; then
    return 1
  fi
  printf '%s [Y/n] ' "$question" > /dev/tty
  read -r answer < /dev/tty || return 1
  [[ -z "$answer" || "$answer" =~ ^[Yy]$ ]]
}

runtime_bin() {
  local name="${1:?missing binary name}"
  local configured="${2:?missing configured path}"
  if [[ -x "$configured" ]]; then
    printf '%s\n' "$configured"
    return 0
  fi
  if command -v "$name" >/dev/null 2>&1; then
    command -v "$name"
    return 0
  fi
  if [[ -x "./dist/$name" ]]; then
    printf './dist/%s\n' "$name"
    return 0
  fi
  if [[ -x "./bin/$name" ]]; then
    printf './bin/%s\n' "$name"
    return 0
  fi
  die "binary not found: $name"
}
