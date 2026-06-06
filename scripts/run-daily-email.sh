#!/usr/bin/env bash
set -euo pipefail

APP_DIR="${GC_RG_WORKDIR:-/opt/gc-rg}"
REPORT_DIR="${GC_RG_REPORT_DIR:-${APP_DIR}/reports/daily}"
DATE="${GC_RG_DATE:-$(date +%F)}"

cd "$APP_DIR"
./bin/gc-rg-generate --date "$DATE"
./bin/gc-rg-email --date "$DATE" --report-dir "$REPORT_DIR" --send
