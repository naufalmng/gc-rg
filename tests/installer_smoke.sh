#!/usr/bin/env bash
set -euo pipefail

ROOT="$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")/.." >/dev/null && pwd -P)"
INSTALLER="${ROOT}/dist/gc-rg.sh"

fail() {
  printf 'FAIL: %s\n' "$*" >&2
  exit 1
}

[[ -x "$INSTALLER" ]] || fail "installer is not executable: $INSTALLER"
bash -n "$INSTALLER" || fail "installer syntax check failed"

help_output="$(bash "$INSTALLER" help)"
[[ "$help_output" == *"gc-rg installer"* ]] || fail "help misses installer title"
[[ "$help_output" == *"sudo bash"*"install"* ]] || fail "help misses install usage"
[[ "$help_output" == *"standalone"* ]] || fail "help misses standalone usage"
[[ "$help_output" == *"gc-rg-generate"* ]] || fail "help misses generate command"
[[ "$help_output" == *"gc-rg-email"* ]] || fail "help misses email command"

if ! grep -q 'PACKAGE_NAME="gc-rg"' "$INSTALLER"; then
  fail "installer package name mismatch"
fi
if ! grep -q '/etc/gc-rg/gc-rg.env' "$INSTALLER"; then
  fail "installer does not manage /etc/gc-rg/gc-rg.env"
fi
if ! grep -q 'gc-rg.timer' "$INSTALLER"; then
  fail "installer does not install timer"
fi
if ! grep -q 'gc-rg-generate-linux-amd64' "$INSTALLER"; then
  fail "installer does not fetch release generate binary"
fi
if ! grep -q 'gc-rg-email-linux-amd64' "$INSTALLER"; then
  fail "installer does not fetch release email binary"
fi

printf 'installer_smoke=pass\n'
