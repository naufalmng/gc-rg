# Runtime entrypoint.

main() {
  parse_args "$@"

  case "$ACTION" in
    onboard) onboard ;;
    config) write_default_config ;;
    config-show) show_config ;;
    config-smtp) config_smtp ;;
    generate) run_generate ;;
    send) run_send ;;
    run) run_all ;;
    status) show_status ;;
    logs) show_logs ;;
    schedule) show_schedule ;;
    enable) enable_timer ;;
    disable) disable_timer ;;
    remove) remove_self ;;
    help) usage ;;
    version) printf '%s %s\n' "$APP" "$VERSION" ;;
    *) die "unknown command: $ACTION"; usage; return 1 ;;
  esac
}

main "$@"
