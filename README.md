<!-- markdownlint-disable MD013 MD024 MD033 MD041 MD046 -->

<div align="center">

     ▄████   ▄████        ██████   ▄████
    ██       ██           ██   ██ ██
    ██  ▄▄   ██           ██████  ██  ▄▄
    ██  ██   ██           ██   ██ ██  ██
     ▀████    ▀████  ▄    ██   ██  ▀████

**Grafana Cloud Report Generator** — Go CLI for Grafana Cloud daily reports,
PDF attachments, SMTP delivery, and Linux-first automatic email scheduling.

[![CI](https://github.com/naufalmng/gc-rg/actions/workflows/ci.yml/badge.svg)](https://github.com/naufalmng/gc-rg/actions/workflows/ci.yml)
[![Release](https://img.shields.io/github/v/release/naufalmng/gc-rg?display_name=tag&sort=semver)](https://github.com/naufalmng/gc-rg/releases)
[![License](https://img.shields.io/github/license/naufalmng/gc-rg)](LICENSE)
[![Go](https://img.shields.io/badge/go-1.25-00ADD8.svg)](https://go.dev/)
[![Platform](https://img.shields.io/badge/platform-linux%20%7C%20windows-d70a53)](#install)

</div>

---

## English

`gc-rg` takes validated Grafana Cloud evidence, renders a daily monitoring
report in Markdown + PDF, then sends the PDF through operator-owned SMTP. It is
the reporting pair for [`gc-hc`](https://github.com/naufalmng/gc-hc): `gc-hc`
checks connectivity, `gc-rg` packages validated evidence into an
operator-friendly daily report.

### What's new

- Unified command: `gc-rg` with short alias `gcrg`.
- User flow now mirrors `gc-hc`: install, onboard, run, status, logs, remove.
- Linux scheduling uses packaged `gc-rg.timer` + `gc-rg.service`.
- Daily scheduled service now calls `gc-rg run --quiet`.
- `gc-rg schedule` shows timer settings and where to edit them.
- Config display masks SMTP password values.

### Install

```bash
curl -fsSL https://github.com/naufalmng/gc-rg/releases/latest/download/gc-rg.sh | sudo bash
sudo gc-rg onboard
```

The installer builds a real `.deb`, installs release binaries, installs
`/usr/bin/gc-rg`, adds `/usr/bin/gcrg`, creates `/etc/gc-rg/gc-rg.env`, and
places daily systemd units on disk. The runtime tree is consistent:
`GC_RG_WORKDIR=/opt/gc-rg`, `GC_RG_EVIDENCE_DIR=/opt/gc-rg/evidence`, and
`GC_RG_REPORT_DIR=/opt/gc-rg/reports/daily`. Onboard only configures and enables
the timer; it does not generate/send until evidence exists. PDF generation uses
`wkhtmltopdf` when available; if your distro does not ship it, install still
works and you can run Markdown-only reports with `gcrg generate --no-pdf`.

### Quick start

```bash
sudo gc-rg onboard       # configure + enable timer + first dry-run
gcrg generate            # generate Markdown + PDF on demand
gcrg send --dry-run      # validate report + SMTP + MIME only
gcrg run                 # generate + send PDF attachment
gcrg schedule            # show systemd timer schedule
gcrg status              # see timer and latest report state
gcrg logs                # show recent service logs
gcrg config show         # show config with secrets masked
sudo gc-rg config smtp   # configure SMTP delivery
sudo gc-rg enable        # enable/start timer
sudo gc-rg disable       # disable/stop timer
sudo apt-get remove gc-rg
```

`gcrg` is a short alias for `gc-rg` — same command, fewer keystrokes. Linux
automation uses `gc-rg.timer`, but operators manage it through `gc-rg enable`,
`gc-rg disable`, and `gc-rg schedule`.

For full usage, configuration reference, architecture, scheduler setup, and
design notes, see **[documentation.md](documentation.md)**.

---

## Bahasa Indonesia

`gc-rg` ambil evidence Grafana Cloud yang sudah tervalidasi, render daily
monitoring report dalam Markdown + PDF, lalu kirim PDF lewat SMTP milik
operator. Tool ini pasangan reporting untuk
[`gc-hc`](https://github.com/naufalmng/gc-hc): `gc-hc` ngecek konektivitas,
`gc-rg` bungkus evidence tervalidasi jadi report harian yang enak dibaca
operator.

### Yang baru

- Command user-facing jadi satu: `gc-rg`, dengan alias pendek `gcrg`.
- Flow mirip `gc-hc`: install, onboard, run, status, logs, remove.
- Scheduling Linux pakai paket `gc-rg.timer` + `gc-rg.service`.
- Service harian sekarang memanggil `gc-rg run --quiet`.
- `gc-rg schedule` menampilkan setting timer dan lokasi edit.
- Tampilan config melakukan masking untuk password SMTP.

### Instalasi

```bash
curl -fsSL https://github.com/naufalmng/gc-rg/releases/latest/download/gc-rg.sh | sudo bash
sudo gc-rg onboard
```

Installer bikin `.deb`, install binary release, install `/usr/bin/gc-rg`, tambah
alias `/usr/bin/gcrg`, buat `/etc/gc-rg/gc-rg.env`, dan taruh unit systemd
harian di disk. Runtime tree konsisten: `GC_RG_WORKDIR=/opt/gc-rg`,
`GC_RG_EVIDENCE_DIR=/opt/gc-rg/evidence`, dan
`GC_RG_REPORT_DIR=/opt/gc-rg/reports/daily`. Onboard cuma config + enable timer;
belum generate/send sampai evidence tersedia. Generate PDF memakai `wkhtmltopdf`
kalau tersedia; kalau distro lo tidak menyediakan package itu, install tetap
jalan dan report Markdown bisa dibuat dengan `gcrg generate --no-pdf`.

### Quick start

```bash
sudo gc-rg onboard       # konfigurasi + aktifkan timer + dry-run pertama
gcrg generate            # generate Markdown + PDF kapan saja
gcrg send --dry-run      # validasi report + SMTP + MIME saja
gcrg run                 # generate + kirim PDF attachment
gcrg schedule            # lihat jadwal systemd timer
gcrg status              # lihat timer dan report terakhir
gcrg logs                # lihat log service terbaru
gcrg config show         # lihat config dengan secret di-mask
sudo gc-rg config smtp   # konfigurasi SMTP delivery
sudo gc-rg enable        # aktifkan/start timer
sudo gc-rg disable       # disable/stop timer
sudo apt-get remove gc-rg
```

`gcrg` adalah alias pendek untuk `gc-rg` — perintah sama, lebih ringkas.
Otomasi Linux memakai `gc-rg.timer`, tapi operator mengelolanya lewat
`gc-rg enable`, `gc-rg disable`, dan `gc-rg schedule`.

Untuk panduan lengkap, referensi konfigurasi, arsitektur, setup scheduler, dan
catatan desain, lihat **[documentation.md](documentation.md)**.

---

## License / Lisensi

Apache License 2.0. See [LICENSE](LICENSE).

## Author / Penulis

[Muhammad Naufal Hanif](https://github.com/naufalmng) — built this so daily
Grafana Cloud evidence can become a repeatable report, not a manual copy-paste
ritual.
