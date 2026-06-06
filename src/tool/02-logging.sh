# Logging helpers.

die() {
  printf '[ERROR] %s\n' "${1:-unknown error}" >&2
  return 1
}

info() {
  if [[ "$QUIET" != "true" ]]; then
    printf '[INFO] %s\n' "${1:-}"
  fi
}

ok() {
  if [[ "$QUIET" != "true" ]]; then
    printf '[OK] %s\n' "${1:-}"
  fi
}

warn() {
  printf '[WARN] %s\n' "${1:-}" >&2
}
