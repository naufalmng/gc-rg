# CLI surface.

usage() {
  cat <<'EOF'
gc-rg - Grafana Cloud daily report generator

Usage:
  gc-rg onboard              interactive config + enable timer + next steps
  gc-rg config               create/update core config
  gc-rg config show          print sanitized config
  gc-rg config smtp          interactive SMTP config with provider templates

  gc-rg generate             generate report once
  gc-rg send                 send existing report
  gc-rg evidence scaffold    create sample evidence JSON templates
  gc-rg run                  generate report then send it
  gc-rg status               show timer and latest report state
  gc-rg logs                 show recent service logs
  gc-rg schedule             show configured systemd schedule
  gc-rg enable               enable/start timer
  gc-rg disable              disable/stop timer
  gc-rg remove               remove package
  gc-rg help                 show help

Short command:
  gcrg onboard
  gcrg generate
  gcrg send --dry-run
  gcrg run
  gcrg status
  gcrg config show
  gcrg logs
  gcrg --help

Options:
  -d, --date today            report date
  -q, --quiet                 less output
  -y, --yes                   assume yes
  -f, --force                 overwrite where relevant
      --send                  deprecated; send is default
      --dry-run               validate without SMTP delivery
  -h, --help                  show help

Environment:
  GC_RG_CONFIG_FILE           default: /etc/gc-rg/gc-rg.env
  GC_RG_REPORT_DIR            default: /opt/gc-rg/reports/daily
  GC_RG_EVIDENCE_DIR          default: /opt/gc-rg/evidence
  GC_RG_GENERATE_BIN          default: /opt/gc-rg/bin/gc-rg-generate
  GC_RG_EMAIL_BIN             default: /opt/gc-rg/bin/gc-rg-email
  GC_RG_SCHEDULE_ON_CALENDAR  default: *-*-* 08:00:00
EOF
}

parse_args() {
  local arg=""
  if (( $# == 0 )); then
    ACTION="help"
    return 0
  fi

  arg="$1"
  shift
  case "$arg" in
    --*) ACTION="${arg#--}" ;;
    *) ACTION="$arg" ;;
  esac

  case "$ACTION" in
    init|setup) ACTION="onboard" ;;
    gen) ACTION="generate" ;;
    email) ACTION="send" ;;
    log) ACTION="logs" ;;
    rm|uninstall) ACTION="remove" ;;
    h) ACTION="help" ;;
    v) ACTION="version" ;;
  esac

  while (( $# > 0 )); do
    arg="$1"
    case "$arg" in
      show|smtp)
        [[ "$ACTION" == "config" ]] || die "$arg subcommand is only valid after config" || return 1
        ACTION="config-$arg"
        shift
        ;;
      scaffold|init|template|templates)
        [[ "$ACTION" == "evidence" || "$ACTION" == "evidence-scaffold" ]] || die "$arg subcommand is only valid after evidence" || return 1
        ACTION="evidence-scaffold"
        shift
        ;;
      -d|--date)
        [[ $# -ge 2 ]] || die "--date needs value" || return 1
        DATE_VALUE="$2"
        shift 2
        ;;
      -q|--quiet) QUIET="true"; shift ;;
      -y|--yes) YES="true"; shift ;;
      -f|--force) FORCE="true"; shift ;;
      --send) SEND_MODE="true"; DRY_RUN="false"; shift ;;
      --dry-run) DRY_RUN="true"; SEND_MODE="false"; shift ;;
      -h|--help) ACTION="help"; shift ;;
      *) die "unknown option: $arg" || return 1 ;;
    esac
  done
}
