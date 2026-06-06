# GC-RG — Grafana Cloud Report Generator

`gc-rg` turns Grafana Cloud validation evidence into daily monitoring reports and sends the generated PDF through provider-agnostic SMTP.

It pairs with `gc-hc`:

- `gc-hc` = Grafana Cloud Health Checker
- `gc-rg` = Grafana Cloud Report Generator

## Commands

Generate daily report:

```bash
go run ./cmd/generate-daily-report --date 2026-06-05
```

Generate Markdown only:

```bash
go run ./cmd/generate-daily-report --date 2026-06-05 --no-pdf
```

Validate email config without sending:

```bash
go run ./cmd/send-email-report --date 2026-06-05 --dry-run
```

Send email:

```bash
go run ./cmd/send-email-report --date 2026-06-05 --send
```

Build binaries on Windows:

```bash
go build -o bin/gc-rg-generate.exe ./cmd/generate-daily-report
go build -o bin/gc-rg-email.exe ./cmd/send-email-report
```

Build binaries on Linux:

```bash
go build -o bin/gc-rg-generate ./cmd/generate-daily-report
go build -o bin/gc-rg-email ./cmd/send-email-report
```

## SMTP config

SMTP is supplied by the operator. `gc-rg` does not provide SMTP infrastructure.

Supported provider presets:

| Provider | Host | Port | TLS | Notes |
|---|---:|---:|---|---|
| `gmail` | `smtp.gmail.com` | `587` | `starttls` | Use Gmail app password. |
| `yahoo` | `smtp.mail.yahoo.com` | `587` | `starttls` | Use Yahoo app password. |
| `outlook` | `smtp.office365.com` | `587` | `starttls` | Depends on tenant SMTP auth policy. |
| `custom` | operator-defined | operator-defined | `starttls`, `ssl`, or `none` | Any SMTP-compatible provider. |

Environment variables:

```bash
GC_RG_EMAIL_PROVIDER=gmail
GC_RG_EMAIL_FROM=your-email@gmail.com
GC_RG_EMAIL_TO=ops@example.com,manager@example.com
GC_RG_EMAIL_CC=
GC_RG_SMTP_USERNAME=your-email@gmail.com
GC_RG_SMTP_PASSWORD=replace-with-app-password
GC_RG_EMAIL_SUBJECT_PREFIX='[GC-RG]'
```

Custom SMTP also needs:

```bash
GC_RG_SMTP_HOST=mail.example.com
GC_RG_SMTP_PORT=587
GC_RG_SMTP_TLS=starttls
GC_RG_SMTP_AUTH=on
```

## Scheduling

Linux uses systemd timer templates in:

```text
deploy/systemd/gc-rg.service
deploy/systemd/gc-rg.timer
deploy/env/gc-rg.env.example
scripts/run-daily-email.sh
```

Windows uses Task Scheduler template in:

```text
deploy/windows/gc-rg-email-daily-task.xml
scripts/run-daily-email.cmd
```

## Evidence inputs

Default evidence paths:

```text
evidence/grafana-longrange-validation/SUMMARY.json
evidence/grafana-prometheus-validation/SUMMARY.json
evidence/grafana-live-loki-scope-24h.json
```

Default report output:

```text
reports/daily/{YYYY-MM-DD}-daily-monitoring-report.md
reports/daily/{YYYY-MM-DD}-daily-monitoring-report.pdf
```

## Notes

- PDF conversion uses Go `goldmark` + `go-wkhtmltopdf` and requires `wkhtmltopdf` in PATH.
- `--dry-run` validates report files, SMTP config, and MIME generation without sending.
- `--send` sends the PDF attachment through configured SMTP.
