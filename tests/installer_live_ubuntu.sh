#!/usr/bin/env bash
set -euo pipefail

fail() {
  printf 'FAIL: %s\n' "$*" >&2
  exit 1
}

need_cmd() {
  local command_name=""
  for command_name in "$@"; do
    command -v "$command_name" >/dev/null 2>&1 || fail "missing command: $command_name"
  done
}

need_cmd bash sudo apt-get dpkg systemctl

ROOT="$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")/.." >/dev/null && pwd -P)"
INSTALLER="${ROOT}/dist/gc-rg.sh"

[[ -x "$INSTALLER" ]] || fail "installer is not executable: $INSTALLER"

if dpkg -s gc-rg >/dev/null 2>&1; then
  sudo apt-get remove -y gc-rg >/dev/null
fi

sudo env GC_RG_ASSET_BASE_URL="file://${ROOT}/dist" bash "$INSTALLER" install --yes --keep-deb

dpkg -s gc-rg >/dev/null 2>&1 || fail "gc-rg package is not installed"
command -v gc-rg >/dev/null 2>&1 || fail "gc-rg command missing"
command -v gcrg >/dev/null 2>&1 || fail "gcrg alias missing"

gc-rg help >/tmp/gc-rg-live-help.out
gc-rg schedule >/tmp/gc-rg-live-schedule.out
grep -q '08:00:00' /tmp/gc-rg-live-schedule.out || fail "schedule output misses default time"
sudo bash -c '. /etc/gc-rg/gc-rg.env; test "$GC_RG_SCHEDULE_ON_CALENDAR" = "*-*-* 08:00:00"' || fail "config schedule value cannot be sourced"

grep -q 'gc-rg run' /lib/systemd/system/gc-rg.service || fail "service misses unified run command"
grep -q 'OnCalendar=' /lib/systemd/system/gc-rg.timer || fail "timer misses OnCalendar"
sudo grep -q 'GC_RG_REPORT_DIR=' /etc/gc-rg/gc-rg.env || fail "config missing report dir"
sudo grep -q 'GC_RG_SCHEDULE_ON_CALENDAR=' /etc/gc-rg/gc-rg.env || fail "config missing schedule"

test -x /opt/gc-rg/bin/gc-rg-generate || fail "generate binary missing"
test -x /opt/gc-rg/bin/gc-rg-email || fail "email binary missing"
test -x /usr/bin/gc-rg || fail "runtime command missing"
test -L /usr/bin/gcrg || fail "short alias symlink missing"

sudo apt-get remove -y gc-rg >/dev/null
if dpkg -s gc-rg >/dev/null 2>&1; then
  fail "gc-rg package still installed after remove"
fi

printf 'installer_live_ubuntu=pass\n'
