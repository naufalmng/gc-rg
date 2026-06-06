#!/usr/bin/env bash
set -euo pipefail

ROOT="$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")/.." >/dev/null && pwd -P)"
INSTALLER="${ROOT}/dist/gc-rg.sh"
TOOL="${ROOT}/dist/gc-rg"

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
[[ "$help_output" == *"sudo gc-rg onboard"* ]] || fail "help misses onboard next step"
[[ "$help_output" == *"gcrg"* ]] || fail "help misses short command alias"
[[ "$help_output" == *"gc-rg generate"* ]] || fail "help misses unified generate command"
[[ "$help_output" == *"gc-rg send"* ]] || fail "help misses unified send command"
[[ "$help_output" == *"gc-rg run"* ]] || fail "help misses unified run command"

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
if ! grep -q '/usr/bin/gc-rg' "$INSTALLER"; then
  fail "installer does not install unified command"
fi
if ! grep -q '/usr/bin/gcrg' "$INSTALLER"; then
  fail "installer does not install short alias"
fi
if ! grep -q 'ExecStart=/usr/bin/gc-rg run --quiet' "$INSTALLER"; then
  fail "systemd service does not use unified run command"
fi
if grep -q 'deb_path="$(build_deb)"' "$INSTALLER"; then
  fail "installer captures noisy build_deb stdout as apt package path"
fi
if ! grep -q 'build_deb "$deb_path"' "$INSTALLER"; then
  fail "installer does not pass explicit deb path to build_deb"
fi
if ! grep -q '^Depends: ca-certificates, systemd, curl$' "$INSTALLER"; then
  fail "package has unexpected hard dependencies"
fi
if grep -q '^Depends: .*wkhtmltopdf' "$INSTALLER"; then
  fail "package must not hard-depend on wkhtmltopdf"
fi
if ! grep -q '^Recommends: wkhtmltopdf$' "$INSTALLER"; then
  fail "package should recommend wkhtmltopdf for PDF support"
fi
if ! grep -q 'GC_RG_ASSET_BASE_URL' "$INSTALLER"; then
  fail "installer cannot override asset source for release verification"
fi

[[ "$(head -n 1 "${ROOT}/dist/gc-rg")" == '#!/usr/bin/env bash' ]] || fail "unified runtime misses shebang"
[[ -x "${ROOT}/dist/gc-rg" ]] || fail "unified runtime is not executable"
bash -n "${ROOT}/dist/gc-rg" || fail "unified runtime syntax check failed"
runtime_help="$("${ROOT}/dist/gc-rg" help)"
[[ "$runtime_help" == *"gc-rg onboard"* ]] || fail "runtime help misses onboard"
[[ "$runtime_help" == *"gcrg"* ]] || fail "runtime help misses gcrg"
[[ "$runtime_help" == *"gc-rg run"* ]] || fail "runtime help misses run"
if [[ "$runtime_help" == *"run_generate || true"* ]]; then
  fail "onboard must not auto-generate before evidence exists"
fi
if [[ "$runtime_help" == *"run_send || true"* ]]; then
  fail "onboard must not auto-send before report exists"
fi
if ! grep -q 'tty_read_secret' "$TOOL"; then
  fail "runtime tool missing secret-aware interactive input"
fi
if ! grep -q 'configure_smtp_interactive' "$TOOL"; then
  fail "runtime tool missing interactive SMTP config flow"
fi
if ! grep -q 'configure_core_interactive' "$TOOL"; then
  fail "runtime tool missing interactive core config flow"
fi
if ! grep -q 'smtp.gmail.com' "$TOOL"; then
  fail "runtime tool missing SMTP provider template defaults"
fi
if ! grep -q 'OnCalendar=${GC_RG_SCHEDULE_ON_CALENDAR:-${SCHEDULE_ON_CALENDAR}}' "$TOOL"; then
  fail "runtime tool does not apply configured schedule to systemd timer"
fi
if grep -q "GC_RG_EMAIL_PROVIDER='" "$TOOL"; then
  fail "runtime config writer uses shell-only single-quote env format"
fi
if ! grep -q 'GC_RG_EVIDENCE_DIR=/opt/gc-rg/evidence' "$INSTALLER"; then
  fail "installer config misses consistent evidence dir"
fi
if ! grep -q '"$MAIN_TOOL" evidence scaffold' "$INSTALLER"; then
  fail "installer must auto-create evidence scaffold after install"
fi
[[ "$runtime_help" == *"gc-rg schedule"* ]] || fail "runtime help misses schedule"
[[ "$runtime_help" == *"gc-rg evidence scaffold"* ]] || fail "runtime help misses evidence scaffold"
if ! grep -q 'local mode="--send"' "${ROOT}/dist/gc-rg"; then
  fail "send should default to real delivery"
fi
if ! grep -q 'deprecated; send is default' "${ROOT}/dist/gc-rg"; then
  fail "send help should mark --send deprecated"
fi

if ! grep -q 'OnCalendar=' "${ROOT}/assets/systemd/gc-rg.timer"; then
  fail "timer misses OnCalendar"
fi
if ! grep -q 'gc-rg schedule' "${ROOT}/documentation.md"; then
  fail "documentation misses schedule command"
fi
if ! grep -q 'gc-rg run' "${ROOT}/documentation.md"; then
  fail "documentation misses unified run command"
fi
if ! grep -q 'GC_RG_SCHEDULE_ON_CALENDAR' "${ROOT}/src/tool/05-config.sh"; then
  fail "config misses schedule environment support"
fi
if ! grep -q 'installer_live_ubuntu.sh' "${ROOT}/.github/workflows/ci.yml"; then
  fail "CI does not run live Ubuntu installer verification"
fi
if ! grep -q 'installer_live_ubuntu.sh' "${ROOT}/.github/workflows/release.yml"; then
  fail "release workflow does not run live Ubuntu installer verification"
fi
if ! grep -q 'GITHUB_REF_NAME' "${ROOT}/.github/workflows/release.yml"; then
  fail "release workflow does not guard tag against VERSION"
fi

printf 'installer_smoke=pass\n'
