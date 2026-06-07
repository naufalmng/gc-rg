# Config UX.

provider_defaults() {
  case "${GC_RG_EMAIL_PROVIDER:-custom}" in
    gmail)
      GC_RG_SMTP_HOST="smtp.gmail.com"
      GC_RG_SMTP_PORT="587"
      GC_RG_SMTP_TLS="starttls"
      GC_RG_SMTP_AUTH="on"
      ;;
    outlook)
      GC_RG_SMTP_HOST="smtp.office365.com"
      GC_RG_SMTP_PORT="587"
      GC_RG_SMTP_TLS="starttls"
      GC_RG_SMTP_AUTH="on"
      ;;
    yahoo)
      GC_RG_SMTP_HOST="smtp.mail.yahoo.com"
      GC_RG_SMTP_PORT="587"
      GC_RG_SMTP_TLS="starttls"
      GC_RG_SMTP_AUTH="on"
      ;;
    custom) ;;
    *) die "email provider must be gmail, outlook, yahoo, or custom"; return 1 ;;
  esac
}

write_config() {
  need_root || return 1
  install -d -m 0750 "$CONFIG_DIR"
  if [[ -e "$CONFIG_FILE" && "$FORCE" != "true" ]]; then
    cp -a "$CONFIG_FILE" "${CONFIG_FILE}.bak.$(date +%Y%m%d%H%M%S)"
  fi
  {
    printf '# GC-RG runtime config\n'
    printf "GC_RG_EMAIL_PROVIDER=\"%s\"\n" "$(quote_env "${GC_RG_EMAIL_PROVIDER:-gmail}")"
    printf "GC_RG_SMTP_HOST=\"%s\"\n" "$(quote_env "${GC_RG_SMTP_HOST:-smtp.gmail.com}")"
    printf "GC_RG_SMTP_PORT=\"%s\"\n" "${GC_RG_SMTP_PORT:-587}"
    printf "GC_RG_SMTP_TLS=\"%s\"\n" "${GC_RG_SMTP_TLS:-starttls}"
    printf "GC_RG_SMTP_AUTH=\"%s\"\n" "${GC_RG_SMTP_AUTH:-on}"
    printf "GC_RG_SMTP_HELO_NAME=\"%s\"\n" "$(quote_env "${GC_RG_SMTP_HELO_NAME:-Superindo}")"
    printf "GC_RG_SMTP_USERNAME=\"%s\"\n" "$(quote_env "${GC_RG_SMTP_USERNAME:-}")"
    printf "GC_RG_SMTP_PASSWORD=\"%s\"\n" "$(quote_env "${GC_RG_SMTP_PASSWORD:-}")"
    printf "GC_RG_EMAIL_FROM=\"%s\"\n" "$(quote_env "${GC_RG_EMAIL_FROM:-}")"
    printf "GC_RG_EMAIL_TO=\"%s\"\n" "$(quote_env "${GC_RG_EMAIL_TO:-}")"
    printf "GC_RG_EMAIL_CC=\"%s\"\n" "$(quote_env "${GC_RG_EMAIL_CC:-}")"
    printf "GC_RG_EMAIL_SUBJECT_PREFIX=\"%s\"\n" "$(quote_env "${GC_RG_EMAIL_SUBJECT_PREFIX:-[GC-RG]}")"
    printf "GC_RG_WORKDIR=\"%s\"\n" "$(quote_env "${GC_RG_WORKDIR:-/opt/gc-rg}")"
    printf "GC_RG_REPORT_DIR=\"%s\"\n" "$(quote_env "${GC_RG_REPORT_DIR:-/opt/gc-rg/reports/daily}")"
    printf "GC_RG_EVIDENCE_DIR=\"%s\"\n" "$(quote_env "${GC_RG_EVIDENCE_DIR:-/opt/gc-rg/evidence}")"
    printf "GC_RG_SCHEDULE_ON_CALENDAR=\"%s\"\n" "$(quote_env "${GC_RG_SCHEDULE_ON_CALENDAR:-*-*-* 08:00:00}")"
  } > "$CONFIG_FILE"
  chmod 0600 "$CONFIG_FILE"
  ok "config saved: $CONFIG_FILE"
}

write_default_config() {
  source_config || true
  GC_RG_EMAIL_PROVIDER="${GC_RG_EMAIL_PROVIDER:-gmail}"
  provider_defaults
  GC_RG_SMTP_USERNAME="${GC_RG_SMTP_USERNAME:-your-email@gmail.com}"
  GC_RG_SMTP_PASSWORD="${GC_RG_SMTP_PASSWORD:-replace-with-app-password}"
  GC_RG_SMTP_HELO_NAME="${GC_RG_SMTP_HELO_NAME:-Superindo}"
  GC_RG_EMAIL_FROM="${GC_RG_EMAIL_FROM:-your-email@gmail.com}"
  GC_RG_EMAIL_TO="${GC_RG_EMAIL_TO:-ops@example.com,manager@example.com}"
  GC_RG_EMAIL_SUBJECT_PREFIX="${GC_RG_EMAIL_SUBJECT_PREFIX:-[GC-RG]}"
  GC_RG_WORKDIR="${GC_RG_WORKDIR:-/opt/gc-rg}"
  GC_RG_REPORT_DIR="${GC_RG_REPORT_DIR:-/opt/gc-rg/reports/daily}"
  GC_RG_EVIDENCE_DIR="${GC_RG_EVIDENCE_DIR:-/opt/gc-rg/evidence}"
  GC_RG_SCHEDULE_ON_CALENDAR="${GC_RG_SCHEDULE_ON_CALENDAR:-*-*-* 08:00:00}"
  write_config
}

prompt_value() {
  local var="${1:?missing var}"
  local label="${2:?missing label}"
  local required="${3:-true}"
  local current="${!var:-}"
  local input=""
  while true; do
    if [[ -n "$current" ]]; then
      if [[ "$var" == *PASSWORD* ]]; then
        input="$(tty_read_secret "${label} [$(mask_value "$current")]: ")" || return 1
      else
        input="$(tty_read "${label} [${current}]: ")" || return 1
      fi
    else
      if [[ "$var" == *PASSWORD* ]]; then
        input="$(tty_read_secret "${label}: ")" || return 1
      else
        input="$(tty_read "${label}: ")" || return 1
      fi
    fi
    [[ -n "$input" || -z "$current" ]] || input="$current"
    if [[ -n "$input" || "$required" != "true" ]]; then
      printf -v "$var" '%s' "$input"
      export "$var"
      return 0
    fi
    warn "value cannot be empty"
  done
}

prompt_choice() {
  local var="${1:?missing var}"
  local label="${2:?missing label}"
  local allowed="${3:?missing allowed}"
  local current="${!var:-}"
  local input=""
  while true; do
    input="$(tty_read "${label} [${current}]: ")" || return 1
    [[ -n "$input" ]] || input="$current"
    if [[ " $allowed " == *" $input "* ]]; then
      printf -v "$var" '%s' "$input"
      export "$var"
      return 0
    fi
    warn "allowed values: $allowed"
  done
}

configure_core_interactive() {
  source_config || true
  GC_RG_WORKDIR="${GC_RG_WORKDIR:-/opt/gc-rg}"
  GC_RG_REPORT_DIR="${GC_RG_REPORT_DIR:-/opt/gc-rg/reports/daily}"
  GC_RG_EVIDENCE_DIR="${GC_RG_EVIDENCE_DIR:-/opt/gc-rg/evidence}"
  GC_RG_SCHEDULE_ON_CALENDAR="${GC_RG_SCHEDULE_ON_CALENDAR:-*-*-* 08:00:00}"
  printf '\nCore config target: %s\n' "$CONFIG_FILE"
  prompt_value "GC_RG_WORKDIR" "Workdir"
  prompt_value "GC_RG_REPORT_DIR" "Report dir"
  prompt_value "GC_RG_EVIDENCE_DIR" "Evidence dir"
  prompt_value "GC_RG_SCHEDULE_ON_CALENDAR" "Systemd schedule"
}

configure_smtp_interactive() {
  source_config || true
  GC_RG_EMAIL_PROVIDER="${GC_RG_EMAIL_PROVIDER:-gmail}"
  GC_RG_EMAIL_SUBJECT_PREFIX="${GC_RG_EMAIL_SUBJECT_PREFIX:-[GC-RG]}"
  GC_RG_SMTP_HELO_NAME="${GC_RG_SMTP_HELO_NAME:-Superindo}"
  printf '\nSMTP config target: %s\n' "$CONFIG_FILE"
  prompt_choice "GC_RG_EMAIL_PROVIDER" "Provider (gmail / outlook / yahoo / custom)" "gmail outlook yahoo custom"
  provider_defaults
  prompt_value "GC_RG_EMAIL_TO" "Email to"
  prompt_value "GC_RG_EMAIL_FROM" "Email from"
  prompt_value "GC_RG_SMTP_USERNAME" "SMTP username"
  prompt_value "GC_RG_SMTP_PASSWORD" "SMTP app password"
  prompt_value "GC_RG_SMTP_HELO_NAME" "SMTP HELO/EHLO name"
  if [[ "$GC_RG_EMAIL_PROVIDER" == "custom" ]]; then
    prompt_value "GC_RG_SMTP_HOST" "SMTP host"
    prompt_value "GC_RG_SMTP_PORT" "SMTP port"
    prompt_choice "GC_RG_SMTP_TLS" "SMTP TLS (starttls / ssl / none)" "starttls ssl none"
    prompt_choice "GC_RG_SMTP_AUTH" "SMTP auth (on / off)" "on off"
  fi
  prompt_value "GC_RG_EMAIL_CC" "Email CC (optional)" false
  prompt_value "GC_RG_EMAIL_SUBJECT_PREFIX" "Subject prefix"
  write_config
}

configure_all_interactive() {
  configure_core_interactive
  configure_smtp_interactive
}

show_config() {
  source_config
  printf '\n────────────────────────────────────────────────────────\n'
  printf '  gc-rg config\n'
  printf '────────────────────────────────────────────────────────\n'
  printf '  %-30s : %s\n' "GC_RG_CONFIG_FILE" "$CONFIG_FILE"
  printf '  %-30s : %s\n' "GC_RG_WORKDIR" "${GC_RG_WORKDIR:-${APP_DIR}}"
  printf '  %-30s : %s\n' "GC_RG_REPORT_DIR" "${GC_RG_REPORT_DIR:-${REPORT_DIR}}"
  printf '  %-30s : %s\n' "GC_RG_EVIDENCE_DIR" "${GC_RG_EVIDENCE_DIR:-${EVIDENCE_DIR}}"
  printf '  %-30s : %s\n' "GC_RG_EMAIL_PROVIDER" "${GC_RG_EMAIL_PROVIDER:-<unset>}"
  printf '  %-30s : %s\n' "GC_RG_SMTP_HOST" "${GC_RG_SMTP_HOST:-<unset>}"
  printf '  %-30s : %s\n' "GC_RG_SMTP_HELO_NAME" "${GC_RG_SMTP_HELO_NAME:-Superindo}"
  printf '  %-30s : %s\n' "GC_RG_SMTP_USERNAME" "${GC_RG_SMTP_USERNAME:-<unset>}"
  printf '  %-30s : %s\n' "GC_RG_SMTP_PASSWORD" "$(mask_value "${GC_RG_SMTP_PASSWORD:-}")"
  printf '  %-30s : %s\n' "GC_RG_EMAIL_TO" "${GC_RG_EMAIL_TO:-<unset>}"
  printf '  %-30s : %s\n' "GC_RG_SCHEDULE_ON_CALENDAR" "${GC_RG_SCHEDULE_ON_CALENDAR:-${SCHEDULE_ON_CALENDAR}}"
  printf '────────────────────────────────────────────────────────\n'
}

config_smtp() {
  if [[ "$YES" == "true" ]]; then
    write_default_config
    return 0
  fi
  configure_smtp_interactive
}
