# systemd lifecycle.

show_schedule() {
  source_config
  printf '\n────────────────────────────────────────────────────────\n'
  printf '  gc-rg schedule\n'
  printf '────────────────────────────────────────────────────────\n'
  printf '  %-24s : %s\n' "Timer" "$TIMER_NAME"
  printf '  %-24s : %s\n' "Service" "$SERVICE_NAME"
  printf '  %-24s : %s\n' "OnCalendar" "${GC_RG_SCHEDULE_ON_CALENDAR:-${SCHEDULE_ON_CALENDAR}}"
  printf '  %-24s : %s\n' "Persistent" "true"
  printf '  %-24s : %s\n' "RandomizedDelaySec" "5m"
  printf '────────────────────────────────────────────────────────\n'
  printf '  Edit schedule in: %s\n' "$CONFIG_FILE"
  printf '  Apply timer with: sudo gc-rg enable\n'
}

enable_timer() {
  need_root || return 1
  need_cmd systemctl || return 1
  systemctl daemon-reload
  systemctl enable --now "$TIMER_NAME"
  ok "timer enabled: $TIMER_NAME"
}

disable_timer() {
  need_root || return 1
  need_cmd systemctl || return 1
  systemctl disable --now "$TIMER_NAME" >/dev/null 2>&1 || true
  systemctl stop "$SERVICE_NAME" >/dev/null 2>&1 || true
  systemctl reset-failed "$SERVICE_NAME" "$TIMER_NAME" >/dev/null 2>&1 || true
  ok "timer disabled: $TIMER_NAME"
}

remove_self() {
  need_root || return 1
  need_cmd apt-get || return 1
  apt-get remove -y gc-rg
}
