<!-- markdownlint-disable MD013 MD024 -->

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
9. [Release](#release)
10. [Troubleshooting](#troubleshooting)

### What it is

`gc-rg` is a Go-based daily report generator for Grafana Cloud validation
evidence. It reads local JSON evidence files, renders an operations-ready
Markdown report, converts that report to PDF, then sends the PDF through SMTP.

It is intentionally evidence-first:

- Prometheus long-range evidence becomes availability and resource sections.
- Latest validation evidence confirms which checks were present.
- Loki 24h scope evidence becomes logs and error summary rows.
- Markdown stays human-readable and versionable.
- PDF becomes the email attachment for daily handoff.
- SMTP delivery stays provider-agnostic and operator-owned.

`gc-rg` now exposes one user-facing command, like `gc-hc`: use `gc-rg` or the
short alias `gcrg`. Internal Go binaries remain implementation detail.

### Install

One-liner installer:

```bash
curl -fsSL https://github.com/naufalmng/gc-rg/releases/latest/download/gc-rg.sh | sudo bash
sudo gc-rg onboard
```

Non-interactive install:

```bash
curl -fsSL https://github.com/naufalmng/gc-rg/releases/latest/download/gc-rg.sh | sudo bash -s -- install --yes
sudo gc-rg onboard
```

Standalone install:

```bash
curl -fsSL https://github.com/naufalmng/gc-rg/releases/latest/download/gc-rg.sh | bash -s -- standalone
./gc-rg-standalone/gcrg generate --date today
./gc-rg-standalone/gcrg send --dry-run
```

The installer builds a local `.deb`, installs Linux release binaries into
`/opt/gc-rg/bin`, installs `/usr/bin/gc-rg`, adds `/usr/bin/gcrg`, creates
`/etc/gc-rg/gc-rg.env`, and installs `gc-rg.service` + `gc-rg.timer`.

### Commands

| Command | What it does |
| --- | --- |
| `gc-rg onboard` | Create config, enable timer, run first dry-run |
| `gc-rg config` | Create/update core config |
| `gc-rg config show` | Print sanitized config with secrets masked |
| `gc-rg config smtp` | Create/update SMTP config |
| `gc-rg generate --date today` | Generate Markdown + PDF |
| `gc-rg send --dry-run` | Validate report, SMTP config, MIME only |
| `gc-rg send --send` | Send PDF attachment through SMTP |
| `gc-rg run --date today` | Generate report, then send email |
| `gc-rg schedule` | Show systemd timer schedule |
| `gc-rg status` | Show timer, service, and latest report state |
| `gc-rg logs` | Show recent `gc-rg.service` journal logs |
| `gc-rg enable` | Enable/start the timer |
| `gc-rg disable` | Disable/stop the timer |
| `gc-rg remove` | Remove package with apt |
| `gcrg run` | Short alias for the daily flow |

### What you'll see

Generate output is path-based so schedulers can log it cleanly:

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

`gc-rg schedule` prints:

```text
gc-rg schedule
Timer                    : gc-rg.timer
Service                  : gc-rg.service
OnCalendar               : *-*-* 08:00:00
Persistent               : true
RandomizedDelaySec       : 5m
Edit schedule in: /etc/gc-rg/gc-rg.env
Apply timer with: sudo gc-rg enable
```

### Configuration

Config lives in `/etc/gc-rg/gc-rg.env` by default. Use:

```bash
sudo gc-rg config
sudo gc-rg config smtp
gcrg config show
```

| Variable | Required | What it is |
| --- | :---: | --- |
| `GC_RG_EMAIL_PROVIDER` | yes | `gmail`, `yahoo`, `outlook`, or `custom` |
| `GC_RG_EMAIL_FROM` | yes | Sender email address |
| `GC_RG_EMAIL_TO` | yes | Comma-separated recipients |
| `GC_RG_EMAIL_CC` | optional | Comma-separated CC recipients |
| `GC_RG_SMTP_USERNAME` | yes when auth on | SMTP username |
| `GC_RG_SMTP_PASSWORD` | yes when auth on | SMTP/app password |
| `GC_RG_SMTP_HOST` | custom only | SMTP host |
| `GC_RG_SMTP_PORT` | optional | SMTP port; default `587` |
| `GC_RG_SMTP_TLS` | optional | `starttls`, `ssl`, or `none` |
| `GC_RG_SMTP_AUTH` | optional | `on` or `off`; default `on` |
| `GC_RG_EMAIL_SUBJECT_PREFIX` | optional | Subject prefix; default `[GC-RG]` |
| `GC_RG_REPORT_DIR` | scheduler | Report directory |
| `GC_RG_WORKDIR` | scheduler | Work directory |
| `GC_RG_SCHEDULE_ON_CALENDAR` | scheduler | systemd `OnCalendar` value |

Provider presets:

| Provider | Host | Port | TLS | Notes |
| --- | --- | ---: | --- | --- |
| `gmail` | `smtp.gmail.com` | `587` | `starttls` | Use app password |
| `yahoo` | `smtp.mail.yahoo.com` | `587` | `starttls` | Use app password |
| `outlook` | `smtp.office365.com` | `587` | `starttls` | Depends on tenant policy |
| `custom` | operator-defined | operator-defined | any | Any SMTP provider |

### Scheduling

Linux scheduling is systemd-first. The timer exists in `assets/systemd/` and is
installed by the `.deb` package. The service runs the same command operators use
manually:

```ini
ExecStart=/usr/bin/gc-rg run --quiet
```

Default timer:

```ini
OnCalendar=*-*-* 08:00:00
Persistent=true
RandomizedDelaySec=5m
Unit=gc-rg.service
```

Manage it through the unified CLI:

```bash
gcrg schedule
sudo gc-rg enable
sudo gc-rg disable
gcrg status
gcrg logs
```

To change the schedule:

```bash
sudoedit /etc/gc-rg/gc-rg.env
sudo gc-rg enable
gcrg schedule
```

Set `GC_RG_SCHEDULE_ON_CALENDAR` to a valid systemd calendar expression. Keep
Linux production on `systemd.timer`, not crontab.

Windows Task Scheduler template remains in
`assets/windows/gc-rg-email-daily-task.xml`, with helper
`scripts/run-daily-email.cmd`.

### How it's wired

```text
evidence/*.json
  ↓
gc-rg generate
  ↓
reports/daily/YYYY-MM-DD-daily-monitoring-report.{md,pdf}
  ↓
gc-rg send --send
  ↓
operator inbox
```

For scheduled Linux runs, `gc-rg.timer` starts `gc-rg.service`, and the service
executes `gc-rg run --quiet`. If generation fails, the send step does not run.

### Build it yourself

```bash
git clone https://github.com/naufalmng/gc-rg
cd gc-rg
go test ./...
bash scripts/build.sh
bash tests/installer_smoke.sh
dist/gc-rg help
dist/gc-rg schedule
```

Source layout:

```text
gc-rg/
├── cmd/                      # internal Go CLIs
├── internal/                 # config, email, generator, report packages
├── src/tool/                 # unified gc-rg shell runtime modules
├── assets/systemd/           # Linux timer/service templates
├── assets/windows/           # Windows Task Scheduler template
├── scripts/                  # build and helper scripts
├── evidence/                 # local validation evidence inputs
└── reports/daily/            # generated report outputs
```

### Release

Release is tag-driven. GitHub Actions publishes `dist/*` when a `v*.*.*` tag is
pushed.

Local release flow:

```bash
go test ./...
bash scripts/build.sh
bash tests/installer_smoke.sh
git add .
git commit -m "fix: keep installer deb path clean"
git push origin main
git tag -a v$(cat VERSION) -m "Release v$(cat VERSION)"
git push origin v$(cat VERSION)
```

The release workflow builds Linux assets, Windows assets, verifies the
installer smoke test, then publishes GitHub release assets.

### Troubleshooting

**`PDF report not found`** — run `gcrg generate --date today`, then confirm
`wkhtmltopdf` exists in `PATH`. If your distro cannot install `wkhtmltopdf`, use
`gcrg generate --no-pdf` for Markdown-only output.

**`GC_RG_SMTP_PASSWORD is required`** — set provider app password in
`/etc/gc-rg/gc-rg.env`, then verify with `gcrg config show`.

**`GC_RG_SMTP_HOST is required`** — use a known provider or set custom host with
`GC_RG_SMTP_HOST`.

**Timer active but no email received** — run `gcrg logs`, then validate with
`gcrg send --dry-run`.

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
9. [Release](#release-1)
10. [Troubleshooting](#troubleshooting-1)

### Apa ini

`gc-rg` adalah generator report harian berbasis Go untuk evidence validasi
Grafana Cloud. Tool ini baca JSON evidence lokal, render report Markdown
siap-operasional, ubah report ke PDF, lalu kirim PDF lewat SMTP.

Prinsipnya evidence-first:

- Evidence Prometheus long-range jadi availability dan resource health.
- Evidence latest validation memastikan check apa saja yang tersedia.
- Evidence Loki 24h scope jadi ringkasan log dan error.
- Markdown tetap bisa dibaca manusia dan gampang di-versioning.
- PDF jadi attachment email untuk handoff harian.
- Delivery SMTP tetap milik operator dan bebas provider.

`gc-rg` sekarang punya satu command user-facing seperti `gc-hc`: pakai `gc-rg`
atau alias pendek `gcrg`. Binary Go internal tetap detail implementasi.

### Instalasi

One-liner installer:

```bash
curl -fsSL https://github.com/naufalmng/gc-rg/releases/latest/download/gc-rg.sh | sudo bash
sudo gc-rg onboard
```

Install non-interactive:

```bash
curl -fsSL https://github.com/naufalmng/gc-rg/releases/latest/download/gc-rg.sh | sudo bash -s -- install --yes
sudo gc-rg onboard
```

Standalone:

```bash
curl -fsSL https://github.com/naufalmng/gc-rg/releases/latest/download/gc-rg.sh | bash -s -- standalone
./gc-rg-standalone/gcrg generate --date today
./gc-rg-standalone/gcrg send --dry-run
```

Installer bikin `.deb` lokal, install binary release ke `/opt/gc-rg/bin`,
install `/usr/bin/gc-rg`, tambah `/usr/bin/gcrg`, buat
`/etc/gc-rg/gc-rg.env`, dan install `gc-rg.service` + `gc-rg.timer`.

### Command

| Command | Fungsi |
| --- | --- |
| `gc-rg onboard` | Buat config, enable timer, lalu first dry-run |
| `gc-rg config` | Buat/update config inti |
| `gc-rg config show` | Tampilkan config aman dengan secret masked |
| `gc-rg config smtp` | Buat/update config SMTP |
| `gc-rg generate --date today` | Generate Markdown + PDF |
| `gc-rg send --dry-run` | Validasi report, config SMTP, MIME saja |
| `gc-rg send --send` | Kirim PDF attachment lewat SMTP |
| `gc-rg run --date today` | Generate report, lalu kirim email |
| `gc-rg schedule` | Tampilkan jadwal systemd timer |
| `gc-rg status` | Tampilkan timer, service, dan report terbaru |
| `gc-rg logs` | Tampilkan journal log `gc-rg.service` |
| `gc-rg enable` | Enable/start timer |
| `gc-rg disable` | Disable/stop timer |
| `gc-rg remove` | Remove package lewat apt |
| `gcrg run` | Alias pendek untuk flow harian |

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

`gc-rg schedule` mencetak:

```text
gc-rg schedule
Timer                    : gc-rg.timer
Service                  : gc-rg.service
OnCalendar               : *-*-* 08:00:00
Persistent               : true
RandomizedDelaySec       : 5m
Edit schedule in: /etc/gc-rg/gc-rg.env
Apply timer with: sudo gc-rg enable
```

### Konfigurasi

Config default ada di `/etc/gc-rg/gc-rg.env`. Pakai:

```bash
sudo gc-rg config
sudo gc-rg config smtp
gcrg config show
```

| Variable | Wajib | Fungsi |
| --- | :---: | --- |
| `GC_RG_EMAIL_PROVIDER` | ya | `gmail`, `yahoo`, `outlook`, atau `custom` |
| `GC_RG_EMAIL_FROM` | ya | Alamat email sender |
| `GC_RG_EMAIL_TO` | ya | Recipient dipisah koma |
| `GC_RG_EMAIL_CC` | opsional | CC dipisah koma |
| `GC_RG_SMTP_USERNAME` | ya kalau auth on | Username SMTP |
| `GC_RG_SMTP_PASSWORD` | ya kalau auth on | Password/app password SMTP |
| `GC_RG_SMTP_HOST` | custom saja | Host SMTP |
| `GC_RG_SMTP_PORT` | opsional | Port SMTP; default `587` |
| `GC_RG_SMTP_TLS` | opsional | `starttls`, `ssl`, atau `none` |
| `GC_RG_SMTP_AUTH` | opsional | `on` atau `off`; default `on` |
| `GC_RG_EMAIL_SUBJECT_PREFIX` | opsional | Prefix subject; default `[GC-RG]` |
| `GC_RG_REPORT_DIR` | scheduler | Direktori report |
| `GC_RG_WORKDIR` | scheduler | Direktori kerja |
| `GC_RG_SCHEDULE_ON_CALENDAR` | scheduler | Nilai `OnCalendar` systemd |

Preset provider:

| Provider | Host | Port | TLS | Catatan |
| --- | --- | ---: | --- | --- |
| `gmail` | `smtp.gmail.com` | `587` | `starttls` | Pakai app password |
| `yahoo` | `smtp.mail.yahoo.com` | `587` | `starttls` | Pakai app password |
| `outlook` | `smtp.office365.com` | `587` | `starttls` | Tergantung policy tenant |
| `custom` | ditentukan operator | ditentukan operator | bebas | SMTP kompatibel |

### Scheduling

Scheduling Linux pakai systemd. Timer ada di `assets/systemd/` dan ikut
terinstall oleh package `.deb`. Service menjalankan command yang sama seperti
manual run:

```ini
ExecStart=/usr/bin/gc-rg run --quiet
```

Timer default:

```ini
OnCalendar=*-*-* 08:00:00
Persistent=true
RandomizedDelaySec=5m
Unit=gc-rg.service
```

Kelola lewat CLI unified:

```bash
gcrg schedule
sudo gc-rg enable
sudo gc-rg disable
gcrg status
gcrg logs
```

Ubah jadwal:

```bash
sudoedit /etc/gc-rg/gc-rg.env
sudo gc-rg enable
gcrg schedule
```

Set `GC_RG_SCHEDULE_ON_CALENDAR` ke ekspresi kalender systemd yang valid. Untuk
Linux production, tetap pakai `systemd.timer`, bukan crontab.

Template Windows Task Scheduler tetap ada di
`assets/windows/gc-rg-email-daily-task.xml`, dengan helper
`scripts/run-daily-email.cmd`.

### Alur kerja

```text
evidence/*.json
  ↓
gc-rg generate
  ↓
reports/daily/YYYY-MM-DD-daily-monitoring-report.{md,pdf}
  ↓
gc-rg send --send
  ↓
inbox operator
```

Untuk scheduled run Linux, `gc-rg.timer` start `gc-rg.service`, lalu service
menjalankan `gc-rg run --quiet`. Kalau generate gagal, send tidak jalan.

### Build sendiri

```bash
git clone https://github.com/naufalmng/gc-rg
cd gc-rg
go test ./...
bash scripts/build.sh
bash tests/installer_smoke.sh
dist/gc-rg help
dist/gc-rg schedule
```

Struktur source:

```text
gc-rg/
├── cmd/                      # internal Go CLI
├── internal/                 # config, email, generator, report packages
├── src/tool/                 # modul runtime shell gc-rg unified
├── assets/systemd/           # template timer/service Linux
├── assets/windows/           # template Task Scheduler Windows
├── scripts/                  # script build/helper
├── evidence/                 # input evidence validasi lokal
└── reports/daily/            # output report generated
```

### Release

Release berbasis tag. GitHub Actions publish `dist/*` saat tag `v*.*.*`
dipush.

Flow release lokal:

```bash
go test ./...
bash scripts/build.sh
bash tests/installer_smoke.sh
git add .
git commit -m "fix: keep installer deb path clean"
git push origin main
git tag -a v$(cat VERSION) -m "Release v$(cat VERSION)"
git push origin v$(cat VERSION)
```

Workflow release build asset Linux, asset Windows, verify installer smoke test,
lalu publish GitHub release assets.

### Troubleshooting

**`PDF report not found`** — jalankan `gcrg generate --date today`, lalu
pastikan `wkhtmltopdf` ada di `PATH`. Kalau distro tidak menyediakan
`wkhtmltopdf`, pakai `gcrg generate --no-pdf` untuk output Markdown saja.

**`GC_RG_SMTP_PASSWORD is required`** — set app password provider di
`/etc/gc-rg/gc-rg.env`, lalu cek dengan `gcrg config show`.

**`GC_RG_SMTP_HOST is required`** — pakai provider dikenal atau set custom host
lewat `GC_RG_SMTP_HOST`.

**Timer aktif tapi email tidak masuk** — jalankan `gcrg logs`, lalu validasi
manual dengan `gcrg send --dry-run`.
