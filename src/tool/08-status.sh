# Status and logs.

show_status() {
  local latest_md=""
  local latest_pdf=""
  printf '\n────────────────────────────────────────────────────────\n'
  printf '  gc-rg status\n'
  printf '────────────────────────────────────────────────────────\n'
  if command -v systemctl >/dev/null 2>&1; then
    printf '  %-18s : %s\n' "timer" "$(systemctl is-active "$TIMER_NAME" 2>/dev/null || true)"
    printf '  %-18s : %s\n' "service" "$(systemctl is-active "$SERVICE_NAME" 2>/dev/null || true)"
  else
    printf '  %-18s : %s\n' "systemd" "unavailable"
  fi
  latest_md="$(ls -1t "${REPORT_DIR}"/*.md 2>/dev/null | head -n 1 || true)"
  latest_pdf="$(ls -1t "${REPORT_DIR}"/*.pdf 2>/dev/null | head -n 1 || true)"
  printf '  %-18s : %s\n' "latest md" "${latest_md:-<none>}"
  printf '  %-18s : %s\n' "latest pdf" "${latest_pdf:-<none>}"
  printf '────────────────────────────────────────────────────────\n'
}

show_logs() {
  need_cmd journalctl || return 1
  journalctl -u "$SERVICE_NAME" -n 100 --no-pager
}
