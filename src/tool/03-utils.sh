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

tty_read() {
  local prompt="${1:?missing prompt}"
  local answer=""
  if [[ ! -r /dev/tty ]]; then
    return 2
  fi
  printf '%s' "$prompt" > /dev/tty
  IFS= read -r answer < /dev/tty || return 1
  answer="$(printf '%s' "$answer" | sed 's/^[[:space:]]*//;s/[[:space:]]*$//')"
  printf '%s' "$answer"
}

tty_read_secret() {
  local prompt="${1:?missing prompt}"
  local answer=""
  local stty_state=""
  if [[ ! -r /dev/tty ]]; then
    return 2
  fi
  stty_state="$(stty -g < /dev/tty)"
  printf '%s' "$prompt" > /dev/tty
  stty -echo < /dev/tty
  if ! IFS= read -r answer < /dev/tty; then
    stty "$stty_state" < /dev/tty
    return 1
  fi
  stty "$stty_state" < /dev/tty
  printf '\n' > /dev/tty
  answer="$(printf '%s' "$answer" | sed 's/^[[:space:]]*//;s/[[:space:]]*$//')"
  printf '%s' "$answer"
}

confirm() {
  local question="${1:?missing question}"
  local default="${2:-y}"
  local suffix="[Y/n]"
  local answer=""
  [[ "$YES" == "true" ]] && return 0
  [[ "$default" =~ ^[Nn]$ ]] && suffix="[y/N]"
  answer="$(tty_read "$question $suffix ")" || return 1
  [[ -n "$answer" ]] || answer="$default"
  [[ "$answer" =~ ^[Yy]([Ee][Ss])?$ ]]
}

quote_env() {
  local value="${1:-}"
  printf '%s' "$value" | sed 's/\\/\\\\/g; s/"/\\"/g'
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
