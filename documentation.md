# Documentation

> 🇬🇧 [English](#english) · 🇮🇩 [Bahasa Indonesia](#bahasa-indonesia)

---

## English

### Table of contents

1. [What it is](#what-it-is)
2. [Install](#install)
3. [Commands](#commands)
4. [What you'll see](#what-youll-see)
5. [Configuration](#configuration)
6. [Scheduling](#scheduling)
7. [How it's wired](#how-its-wired)
8. [Build it yourself](#build-it-yourself)
9. [Design choices](#design-choices)
10. [Troubleshooting](#troubleshooting)

### What it is

`gc-rg` is a Go-based daily report generator for Grafana Cloud validation evidence. It reads local JSON evidence files, renders an operations-ready Markdown report, converts that report to PDF, then sends the PDF through SMTP when requested.

It is intentionally evidence-first:

- **Prometheus long-range summary** becomes availability and resource-health sections.
- **Latest validation summary** confirms which checks were present in the source bundle.
- **Loki 24h scope evidence** becomes logs and error summary rows.
- **Markdown report** remains human-readable and versionable.
- **PDF report** becomes the email attachment for daily handoff.
- **SMTP delivery** stays provider-agnostic and operator-owned.

`gc-rg` does not scrape Grafana directly during email delivery. Generate first, validate files, then send. That separation makes daily scheduling safer on Linux and Windows.

### Install

Clone and build:

```bash
git clone https://github.com/naufalmng/gc-rg
cd gc-rg
go test ./...
go build -o bin/gc-rg-generate ./cmd/generate-daily-report
go build -o bin/gc-rg-email ./cmd/send-email-report
```

Install `wkhtmltopdf` for PDF output:

```bash
sudo apt-get update
sudo apt-get install -y wkhtmltopdf ca-certificates
```

If `wkhtmltopdf` is not available, generate Markdown only:

```bash
./bin/gc-rg-generate --date today --no-pdf
```

### Commands

| Command | What it does |
| --- | --- |
| `gc-rg-generate --date today` | Generate Markdown + PDF from default evidence paths |
| `gc-rg-generate --date 2026-06-05 --no-pdf` | Generate Markdown only |
| `gc-rg-generate --output-dir reports/daily` | Override report output directory |
| `gc-rg-generate --long-range-json PATH` | Override long-range Prometheus evidence path |
| `gc-rg-generate --latest-json PATH` | Override latest validation evidence path |
| `gc-rg-generate --loki-scope-json PATH` | Override optional Loki scope evidence path |
| `gc-rg-email --date today --dry-run` | Validate report files, SMTP config, and MIME build without sending |
| `gc-rg-email --date today --send` | Send PDF attachment through SMTP |
| `gc-rg-email --report-dir PATH` | Read report files from a custom directory |

Email flags mirror environment variables: `--email-provider`, `--smtp-host`, `--smtp-port`, `--smtp-tls`, `--email-from`, `--email-to`, `--email-cc`, `--smtp-username`, `--smtp-password`, and `--email-subject-prefix`.

### What you'll see

Generate output is intentionally path-based so schedulers can log it cleanly:

```text
reports/daily/2026-06-05-daily-monitoring-report.md
reports/daily/2026-06-05-daily-monitoring-report.pdf
```

Email dry-run prints a delivery plan without exposing password values:

```text
mode=dry-run
date=2026-06-05
smtp_provider=gmail
smtp_host=smtp.gmail.com
smtp_port=587
email_from=your-email@gmail.com
email_to_count=2
attachment=reports/daily/2026-06-05-daily-monitoring-report.pdf
attachment_size=48213
dry_run_result=validated, not sent
```

Real send ends with:

```text
send_result=sent
```

### Configuration

Email configuration comes from environment variables or matching CLI flags.

| Variable | Required | What it is |
| --- | :---: | --- |
| `GC_RG_EMAIL_PROVIDER` | yes | `gmail`, `yahoo`, `outlook`, or `custom` |
| `GC_RG_EMAIL_FROM` | yes | Sender email address |
| `GC_RG_EMAIL_TO` | yes | Comma-separated recipients |
| `GC_RG_EMAIL_CC` | optional | Comma-separated CC recipients |
| `GC_RG_SMTP_USERNAME` | yes when auth on | SMTP username |
| `GC_RG_SMTP_PASSWORD` | yes when auth on | SMTP password or provider app password |
| `GC_RG_SMTP_HOST` | custom only | SMTP host; provider presets fill this automatically |
| `GC_RG_SMTP_PORT` | optional | SMTP port; default `587` |
| `GC_RG_SMTP_TLS` | optional | `starttls`, `ssl`, or `none`; default `starttls` |
| `GC_RG_SMTP_AUTH` | optional | `on` or `off`; default `on` |
| `GC_RG_EMAIL_SUBJECT_PREFIX` | optional | Subject prefix; default `[GC-RG]` |
| `GC_RG_REPORT_DIR` | scheduler | Report directory used by systemd template |
| `GC_RG_WORKDIR` | scheduler | Work directory used by scripts/templates |

Provider presets:

| Provider | Host | Port | TLS | Notes |
| --- | --- | ---: | --- | --- |
| `gmail` | `smtp.gmail.com` | `587` | `starttls` | Use Gmail app password |
| `yahoo` | `smtp.mail.yahoo.com` | `587` | `starttls` | Use Yahoo app password |
| `outlook` | `smtp.office365.com` | `587` | `starttls` | Depends on tenant SMTP auth policy |
| `custom` | operator-defined | operator-defined | `starttls`, `ssl`, or `none` | Any SMTP-compatible provider |

Example:

```bash
export GC_RG_EMAIL_PROVIDER=gmail
export GC_RG_EMAIL_FROM=your-email@gmail.com
export GC_RG_EMAIL_TO=ops@example.com,manager@example.com
export GC_RG_SMTP_USERNAME=your-email@gmail.com
export GC_RG_SMTP_PASSWORD=replace-with-app-password
export GC_RG_EMAIL_SUBJECT_PREFIX='[GC-RG]'
```

### Scheduling

Linux uses systemd templates in `assets/systemd/`.

Recommended layout:

```text
/opt/gc-rg/
├── bin/
│   ├── gc-rg-generate
│   └── gc-rg-email
├── evidence/
│   ├── grafana-longrange-validation/SUMMARY.json
│   ├── grafana-prometheus-validation/SUMMARY.json
│   └── grafana-live-loki-scope-24h.json
├── reports/daily/
└── tmp/
```

Setup example:

```bash
sudo useradd --system --home /opt/gc-rg --shell /usr/sbin/nologin gc-rg || true
sudo mkdir -p /opt/gc-rg/bin /opt/gc-rg/evidence /opt/gc-rg/reports/daily /opt/gc-rg/tmp /etc/gc-rg
sudo cp bin/gc-rg-generate bin/gc-rg-email /opt/gc-rg/bin/
sudo cp assets/env/gc-rg.env.example /etc/gc-rg/gc-rg.env
sudo chmod 600 /etc/gc-rg/gc-rg.env
sudo chown -R gc-rg:gc-rg /opt/gc-rg
sudo cp assets/systemd/gc-rg.service assets/systemd/gc-rg.timer /etc/systemd/system/
sudo systemctl daemon-reload
sudo systemctl enable --now gc-rg.timer
```

Edit secrets before enabling production sends:

```bash
sudoedit /etc/gc-rg/gc-rg.env
sudo systemctl start gc-rg.service
sudo journalctl -u gc-rg.service -n 100 --no-pager
```

Windows Task Scheduler template lives in `assets/windows/gc-rg-email-daily-task.xml`, with helper script `scripts/run-daily-email.cmd`.

### How it's wired

```text
          ┌──────────────────────────────┐
          │ evidence/*.json              │
          │ Prometheus + Loki summaries  │
          └──────────────┬───────────────┘
                         │
                         ▼
          ┌──────────────────────────────┐
          │ gc-rg-generate               │
          │ JSON → Markdown → PDF        │
          └──────────────┬───────────────┘
                         │
                         ▼
          reports/daily/YYYY-MM-DD-daily-monitoring-report.{md,pdf}
                         │
                         ▼
          ┌──────────────────────────────┐
          │ gc-rg-email                  │
          │ resolve → validate → MIME    │
          │ dry-run or SMTP send         │
          └──────────────┬───────────────┘
                         │
                         ▼
                  operator inbox
```

The scheduler runs generation first and email second. If generation fails, systemd stops the unit and the email step does not run.

### Build it yourself

```bash
git clone https://github.com/naufalmng/gc-rg
cd gc-rg
go test ./...
go build -o bin/gc-rg-generate ./cmd/generate-daily-report
go build -o bin/gc-rg-email ./cmd/send-email-report
```

Source layout:

```text
gc-rg/
├── cmd/
│   ├── generate-daily-report/ # report generation CLI
│   ├── send-email-report/     # SMTP email CLI
│   └── send-whatsapp-report/  # WhatsApp sender experiment
├── internal/
│   ├── config/                # shared config parsing
│   ├── email/                 # SMTP config, MIME, delivery
│   ├── generator/             # evidence parsing + report rendering
│   ├── report/                # report file resolution + status extraction
│   └── whatsapp/              # WhatsApp delivery support
├── assets/
│   ├── env/                   # env examples
│   ├── systemd/               # Linux timer/service templates
│   └── windows/               # Task Scheduler template
├── scripts/                   # scheduler helper scripts
├── evidence/                  # local validation evidence inputs
└── reports/daily/             # generated report outputs
```

### Design choices

- **Generate and send are separate commands.** This keeps scheduler failure modes obvious: no valid PDF, no email.
- **Dry-run is default when `--send` is absent.** Running the email command without a send flag validates only.
- **Operator-owned SMTP.** `gc-rg` does not hide delivery behind a SaaS API; it uses your configured SMTP provider.
- **Provider presets are conservative.** Gmail, Yahoo, and Outlook set host/port/TLS defaults, but credentials remain explicit.
- **Linux-first scheduling.** systemd service/timer templates support headless servers and persist missed runs via `Persistent=true`.
- **Evidence stays local.** Reports are generated from files, so runs are reproducible and easy to audit.

### Troubleshooting

**`PDF report not found`** — run `gc-rg-generate --date today` without `--no-pdf`, and confirm `wkhtmltopdf` exists in `PATH`.

**`GC_RG_SMTP_PASSWORD is required`** — set provider app password in `/etc/gc-rg/gc-rg.env` or pass `--smtp-password`.

**`GC_RG_SMTP_HOST is required`** — use a known provider or set custom host with `GC_RG_SMTP_HOST`.

**`x509: certificate signed by unknown authority`** — install `ca-certificates` on the host.

**Timer active but no email received** — check `journalctl -u gc-rg.service -n 100 --no-pager`, then run `gc-rg-email --date today --dry-run` manually.

---

## Bahasa Indonesia

### Daftar isi

1. [Apa ini](#apa-ini)
2. [Instalasi](#instalasi)
3. [Command](#command)
4. [Output yang terlihat](#output-yang-terlihat)
5. [Konfigurasi](#konfigurasi)
6. [Scheduling](#scheduling-1)
7. [Alur kerja](#alur-kerja)
8. [Build sendiri](#build-sendiri)
9. [Keputusan desain](#keputusan-desain)
10. [Troubleshooting](#troubleshooting-1)

### Apa ini

`gc-rg` adalah generator report harian berbasis Go untuk evidence validasi Grafana Cloud. Tool ini baca JSON evidence lokal, render report Markdown siap-operasional, ubah report ke PDF, lalu kirim PDF lewat SMTP saat diminta.

Prinsipnya evidence-first:

- **Prometheus long-range summary** jadi bagian availability dan resource health.
- **Latest validation summary** memastikan check apa saja yang ada di bundle sumber.
- **Loki 24h scope evidence** jadi ringkasan log dan error.
- **Markdown report** tetap bisa dibaca manusia dan gampang di-versioning.
- **PDF report** jadi attachment email untuk handoff harian.
- **SMTP delivery** tetap milik operator dan bebas provider.

`gc-rg` tidak scrape Grafana langsung saat kirim email. Generate dulu, validasi file, baru kirim. Pemisahan ini bikin scheduling harian lebih aman di Linux dan Windows.

### Instalasi

Clone dan build:

```bash
git clone https://github.com/naufalmng/gc-rg
cd gc-rg
go test ./...
go build -o bin/gc-rg-generate ./cmd/generate-daily-report
go build -o bin/gc-rg-email ./cmd/send-email-report
```

Install `wkhtmltopdf` untuk output PDF:

```bash
sudo apt-get update
sudo apt-get install -y wkhtmltopdf ca-certificates
```

Kalau `wkhtmltopdf` belum ada, generate Markdown saja:

```bash
./bin/gc-rg-generate --date today --no-pdf
```

### Command

| Command | Fungsi |
| --- | --- |
| `gc-rg-generate --date today` | Generate Markdown + PDF dari path evidence default |
| `gc-rg-generate --date 2026-06-05 --no-pdf` | Generate Markdown saja |
| `gc-rg-generate --output-dir reports/daily` | Override direktori output report |
| `gc-rg-generate --long-range-json PATH` | Override path evidence Prometheus long-range |
| `gc-rg-generate --latest-json PATH` | Override path evidence latest validation |
| `gc-rg-generate --loki-scope-json PATH` | Override path optional evidence Loki scope |
| `gc-rg-email --date today --dry-run` | Validasi file report, config SMTP, dan MIME tanpa kirim |
| `gc-rg-email --date today --send` | Kirim attachment PDF lewat SMTP |
| `gc-rg-email --report-dir PATH` | Baca report dari direktori custom |

Flag email sama dengan env var: `--email-provider`, `--smtp-host`, `--smtp-port`, `--smtp-tls`, `--email-from`, `--email-to`, `--email-cc`, `--smtp-username`, `--smtp-password`, dan `--email-subject-prefix`.

### Output yang terlihat

Output generate sengaja berupa path supaya log scheduler bersih:

```text
reports/daily/2026-06-05-daily-monitoring-report.md
reports/daily/2026-06-05-daily-monitoring-report.pdf
```

Email dry-run mencetak rencana delivery tanpa membocorkan password:

```text
mode=dry-run
date=2026-06-05
smtp_provider=gmail
smtp_host=smtp.gmail.com
smtp_port=587
email_from=your-email@gmail.com
email_to_count=2
attachment=reports/daily/2026-06-05-daily-monitoring-report.pdf
attachment_size=48213
dry_run_result=validated, not sent
```

Kirim sungguhan selesai dengan:

```text
send_result=sent
```

### Konfigurasi

Konfigurasi email berasal dari environment variable atau flag CLI yang setara.

| Variable | Wajib | Fungsi |
| --- | :---: | --- |
| `GC_RG_EMAIL_PROVIDER` | ya | `gmail`, `yahoo`, `outlook`, atau `custom` |
| `GC_RG_EMAIL_FROM` | ya | Alamat email sender |
| `GC_RG_EMAIL_TO` | ya | Recipient dipisah koma |
| `GC_RG_EMAIL_CC` | opsional | CC dipisah koma |
| `GC_RG_SMTP_USERNAME` | ya kalau auth on | Username SMTP |
| `GC_RG_SMTP_PASSWORD` | ya kalau auth on | Password SMTP atau app password provider |
| `GC_RG_SMTP_HOST` | custom saja | Host SMTP; preset provider isi otomatis |
| `GC_RG_SMTP_PORT` | opsional | Port SMTP; default `587` |
| `GC_RG_SMTP_TLS` | opsional | `starttls`, `ssl`, atau `none`; default `starttls` |
| `GC_RG_SMTP_AUTH` | opsional | `on` atau `off`; default `on` |
| `GC_RG_EMAIL_SUBJECT_PREFIX` | opsional | Prefix subject; default `[GC-RG]` |
| `GC_RG_REPORT_DIR` | scheduler | Direktori report yang dipakai template systemd |
| `GC_RG_WORKDIR` | scheduler | Working directory untuk script/template |

Preset provider:

| Provider | Host | Port | TLS | Catatan |
| --- | --- | ---: | --- | --- |
| `gmail` | `smtp.gmail.com` | `587` | `starttls` | Pakai Gmail app password |
| `yahoo` | `smtp.mail.yahoo.com` | `587` | `starttls` | Pakai Yahoo app password |
| `outlook` | `smtp.office365.com` | `587` | `starttls` | Tergantung policy SMTP auth tenant |
| `custom` | ditentukan operator | ditentukan operator | `starttls`, `ssl`, atau `none` | Provider SMTP apa pun |

Contoh:

```bash
export GC_RG_EMAIL_PROVIDER=gmail
export GC_RG_EMAIL_FROM=your-email@gmail.com
export GC_RG_EMAIL_TO=ops@example.com,manager@example.com
export GC_RG_SMTP_USERNAME=your-email@gmail.com
export GC_RG_SMTP_PASSWORD=replace-with-app-password
export GC_RG_EMAIL_SUBJECT_PREFIX='[GC-RG]'
```

### Scheduling

Linux pakai template systemd di `assets/systemd/`.

Layout yang disarankan:

```text
/opt/gc-rg/
├── bin/
│   ├── gc-rg-generate
│   └── gc-rg-email
├── evidence/
│   ├── grafana-longrange-validation/SUMMARY.json
│   ├── grafana-prometheus-validation/SUMMARY.json
│   └── grafana-live-loki-scope-24h.json
├── reports/daily/
└── tmp/
```

Contoh setup:

```bash
sudo useradd --system --home /opt/gc-rg --shell /usr/sbin/nologin gc-rg || true
sudo mkdir -p /opt/gc-rg/bin /opt/gc-rg/evidence /opt/gc-rg/reports/daily /opt/gc-rg/tmp /etc/gc-rg
sudo cp bin/gc-rg-generate bin/gc-rg-email /opt/gc-rg/bin/
sudo cp assets/env/gc-rg.env.example /etc/gc-rg/gc-rg.env
sudo chmod 600 /etc/gc-rg/gc-rg.env
sudo chown -R gc-rg:gc-rg /opt/gc-rg
sudo cp assets/systemd/gc-rg.service assets/systemd/gc-rg.timer /etc/systemd/system/
sudo systemctl daemon-reload
sudo systemctl enable --now gc-rg.timer
```

Edit secret sebelum aktif produksi:

```bash
sudoedit /etc/gc-rg/gc-rg.env
sudo systemctl start gc-rg.service
sudo journalctl -u gc-rg.service -n 100 --no-pager
```

Template Windows Task Scheduler ada di `assets/windows/gc-rg-email-daily-task.xml`, dengan helper `scripts/run-daily-email.cmd`.

### Alur kerja

```text
          ┌──────────────────────────────┐
          │ evidence/*.json              │
          │ Prometheus + Loki summaries  │
          └──────────────┬───────────────┘
                         │
                         ▼
          ┌──────────────────────────────┐
          │ gc-rg-generate               │
          │ JSON → Markdown → PDF        │
          └──────────────┬───────────────┘
                         │
                         ▼
          reports/daily/YYYY-MM-DD-daily-monitoring-report.{md,pdf}
                         │
                         ▼
          ┌──────────────────────────────┐
          │ gc-rg-email                  │
          │ resolve → validate → MIME    │
          │ dry-run or SMTP send         │
          └──────────────┬───────────────┘
                         │
                         ▼
                  inbox operator
```

Scheduler menjalankan generate dulu lalu email. Kalau generate gagal, systemd stop unit dan step email tidak jalan.

### Build sendiri

```bash
git clone https://github.com/naufalmng/gc-rg
cd gc-rg
go test ./...
go build -o bin/gc-rg-generate ./cmd/generate-daily-report
go build -o bin/gc-rg-email ./cmd/send-email-report
```

Struktur source:

```text
gc-rg/
├── cmd/
│   ├── generate-daily-report/ # CLI generate report
│   ├── send-email-report/     # CLI email SMTP
│   └── send-whatsapp-report/  # eksperimen sender WhatsApp
├── internal/
│   ├── config/                # parsing config shared
│   ├── email/                 # config SMTP, MIME, delivery
│   ├── generator/             # parsing evidence + render report
│   ├── report/                # resolve file report + status extraction
│   └── whatsapp/              # support delivery WhatsApp
├── assets/
│   ├── env/                   # contoh env
│   ├── systemd/               # template timer/service Linux
│   └── windows/               # template Task Scheduler
├── scripts/                   # helper script scheduler
├── evidence/                  # input evidence validasi lokal
└── reports/daily/             # output report generated
```

### Keputusan desain

- **Generate dan send dipisah.** Failure mode scheduler jelas: PDF tidak valid, email tidak jalan.
- **Dry-run default saat `--send` tidak ada.** Command email tanpa flag send hanya validasi.
- **SMTP milik operator.** `gc-rg` tidak sembunyikan delivery di balik API SaaS; dia pakai provider SMTP yang kamu set.
- **Preset provider konservatif.** Gmail, Yahoo, dan Outlook isi default host/port/TLS, tapi credential tetap eksplisit.
- **Scheduling Linux-first.** Template systemd service/timer support server headless dan missed run dengan `Persistent=true`.
- **Evidence tetap lokal.** Report dihasilkan dari file, jadi run bisa diaudit dan direproduksi.

### Troubleshooting

**`PDF report not found`** — jalankan `gc-rg-generate --date today` tanpa `--no-pdf`, dan pastikan `wkhtmltopdf` ada di `PATH`.

**`GC_RG_SMTP_PASSWORD is required`** — set app password provider di `/etc/gc-rg/gc-rg.env` atau lewat `--smtp-password`.

**`GC_RG_SMTP_HOST is required`** — pakai provider dikenal atau set custom host dengan `GC_RG_SMTP_HOST`.

**`x509: certificate signed by unknown authority`** — install `ca-certificates` di host.

**Timer aktif tapi email tidak masuk** — cek `journalctl -u gc-rg.service -n 100 --no-pager`, lalu jalankan `gc-rg-email --date today --dry-run` manual.
