# Runtime defaults.

ACTION="help"
QUIET="false"
YES="false"
FORCE="false"
DATE_VALUE="today"
SEND_MODE="false"
DRY_RUN="false"

CONFIG_DIR="${GC_RG_CONFIG_DIR:-/etc/gc-rg}"
CONFIG_FILE="${GC_RG_CONFIG_FILE:-${CONFIG_DIR}/gc-rg.env}"
APP_DIR="${GC_RG_APP_DIR:-/opt/gc-rg}"
BIN_DIR="${GC_RG_BIN_DIR:-${APP_DIR}/bin}"
REPORT_DIR="${GC_RG_REPORT_DIR:-${APP_DIR}/reports/daily}"
EVIDENCE_DIR="${GC_RG_EVIDENCE_DIR:-${APP_DIR}/evidence}"
TMP_DIR="${GC_RG_TMP_DIR:-${APP_DIR}/tmp}"
SERVICE_NAME="gc-rg.service"
TIMER_NAME="gc-rg.timer"

GENERATE_BIN="${GC_RG_GENERATE_BIN:-${BIN_DIR}/gc-rg-generate}"
EMAIL_BIN="${GC_RG_EMAIL_BIN:-${BIN_DIR}/gc-rg-email}"
SCHEDULE_ON_CALENDAR="${GC_RG_SCHEDULE_ON_CALENDAR:-*-*-* 08:00:00}"
