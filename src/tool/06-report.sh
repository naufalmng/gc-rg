# Report actions.

run_generate() {
  local bin=""
  source_config
  bin="$(runtime_bin gc-rg-generate "$GENERATE_BIN")" || return 1
  "$bin" --date "$DATE_VALUE"
}

run_send() {
  local bin=""
  local mode="--dry-run"
  source_config
  bin="$(runtime_bin gc-rg-email "$EMAIL_BIN")" || return 1
  if [[ "$SEND_MODE" == "true" ]]; then
    mode="--send"
  fi
  "$bin" --date "$DATE_VALUE" --report-dir "${GC_RG_REPORT_DIR:-$REPORT_DIR}" "$mode"
}

run_all() {
  run_generate || return 1
  SEND_MODE="true"
  run_send
}

onboard() {
  cat <<EOF
Onboard plan:
  1. Create/update config:
     ${CONFIG_FILE}

  2. Enable systemd timer:
     ${TIMER_NAME}

  3. Generate first report and validate email dry-run.
EOF
  if ! confirm "Continue onboard?"; then
    info "cancelled"
    return 0
  fi
  write_default_config
  enable_timer
  run_generate || true
  DRY_RUN="true"
  SEND_MODE="false"
  run_send || true
  show_status
}
