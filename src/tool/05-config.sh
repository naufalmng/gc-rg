# Config UX.

write_default_config() {
  need_root || return 1
  install -d -m 0750 "$CONFIG_DIR"
  if [[ -e "$CONFIG_FILE" && "$FORCE" != "true" ]]; then
    cp -a "$CONFIG_FILE" "${CONFIG_FILE}.bak.$(date +%Y%m%d%H%M%S)"
  fi
  cat > "$CONFIG_FILE" <<'EOF'
# GC-RG runtime config
GC_RG_EMAIL_PROVIDER=gmail
GC_RG_SMTP_HOST=
GC_RG_SMTP_PORT=587
GC_RG_SMTP_TLS=starttls
GC_RG_SMTP_AUTH=on
GC_RG_SMTP_USERNAME=your-email@gmail.com
GC_RG_SMTP_PASSWORD=replace-with-app-password
GC_RG_EMAIL_FROM=your-email@gmail.com
GC_RG_EMAIL_TO=ops@example.com,manager@example.com
GC_RG_EMAIL_CC=
GC_RG_EMAIL_SUBJECT_PREFIX=[GC-RG]
GC_RG_REPORT_DIR=/opt/gc-rg/reports/daily
GC_RG_WORKDIR=/opt/gc-rg
GC_RG_SCHEDULE_ON_CALENDAR="*-*-* 08:00:00"
EOF
  chmod 0600 "$CONFIG_FILE"
  ok "config saved: $CONFIG_FILE"
}

show_config() {
  source_config
  printf '\n────────────────────────────────────────────────────────\n'
  printf '  gc-rg config\n'
  printf '────────────────────────────────────────────────────────\n'
  printf '  %-30s : %s\n' "GC_RG_CONFIG_FILE" "$CONFIG_FILE"
  printf '  %-30s : %s\n' "GC_RG_REPORT_DIR" "${GC_RG_REPORT_DIR:-${REPORT_DIR}}"
  printf '  %-30s : %s\n' "GC_RG_WORKDIR" "${GC_RG_WORKDIR:-${APP_DIR}}"
  printf '  %-30s : %s\n' "GC_RG_EMAIL_PROVIDER" "${GC_RG_EMAIL_PROVIDER:-<unset>}"
  printf '  %-30s : %s\n' "GC_RG_SMTP_HOST" "${GC_RG_SMTP_HOST:-<unset>}"
  printf '  %-30s : %s\n' "GC_RG_SMTP_USERNAME" "${GC_RG_SMTP_USERNAME:-<unset>}"
  printf '  %-30s : %s\n' "GC_RG_SMTP_PASSWORD" "$(mask_value "${GC_RG_SMTP_PASSWORD:-}")"
  printf '  %-30s : %s\n' "GC_RG_EMAIL_TO" "${GC_RG_EMAIL_TO:-<unset>}"
  printf '  %-30s : %s\n' "GC_RG_SCHEDULE_ON_CALENDAR" "${GC_RG_SCHEDULE_ON_CALENDAR:-${SCHEDULE_ON_CALENDAR}}"
  printf '────────────────────────────────────────────────────────\n'
}

config_smtp() {
  info "SMTP config uses same file for now: $CONFIG_FILE"
  write_default_config
}
