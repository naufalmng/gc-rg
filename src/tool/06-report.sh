# Report actions.

run_generate() {
  local bin=""
  source_config
  bin="$(runtime_bin gc-rg-generate "$GENERATE_BIN")" || return 1
  "$bin" \
    --date "$DATE_VALUE" \
    --output-dir "${GC_RG_REPORT_DIR:-$REPORT_DIR}" \
    --long-range-json "${GC_RG_EVIDENCE_DIR:-$EVIDENCE_DIR}/grafana-longrange-validation/SUMMARY.json" \
    --latest-json "${GC_RG_EVIDENCE_DIR:-$EVIDENCE_DIR}/grafana-prometheus-validation/SUMMARY.json" \
    --loki-scope-json "${GC_RG_EVIDENCE_DIR:-$EVIDENCE_DIR}/grafana-live-loki-scope-24h.json"
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

  3. Configure core paths and SMTP interactively unless --yes is passed.

Fresh install creates bundled sample evidence automatically, so first generate can run.
Production reports need real validated evidence replacing those sample files.

Expected evidence layout:
  ${EVIDENCE_DIR}/grafana-longrange-validation/SUMMARY.json
  ${EVIDENCE_DIR}/grafana-prometheus-validation/SUMMARY.json
  ${EVIDENCE_DIR}/grafana-live-loki-scope-24h.json
EOF
  if ! confirm "Continue onboard?"; then
    info "cancelled"
    return 0
  fi
  if [[ "$YES" == "true" ]]; then
    write_default_config
  else
    configure_all_interactive
  fi
  enable_timer
  cat <<EOF

Next steps:
  1. Replace sample evidence with validated Grafana evidence under:
     ${EVIDENCE_DIR}

  2. Generate report:
     sudo gc-rg generate

  3. Validate email config without sending:
     sudo gc-rg send --dry-run

  4. Send for real:
     sudo gc-rg send --send

Timer is enabled. First scheduled run uses sample evidence until real evidence replaces it.
EOF
  show_status
}
