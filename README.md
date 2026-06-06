<div align="center">

     ▄████   ▄████        ██████   ▄████
    ██       ██           ██   ██ ██
    ██  ▄▄   ██           ██████  ██  ▄▄
    ██  ██   ██           ██   ██ ██  ██
     ▀████    ▀████  ▄    ██   ██  ▀████

**Grafana Cloud Report Generator** — Go CLI for turning Grafana Cloud validation evidence into daily monitoring reports and sending the PDF by SMTP.

[![CI](https://github.com/naufalmng/gc-rg/actions/workflows/ci.yml/badge.svg)](https://github.com/naufalmng/gc-rg/actions/workflows/ci.yml)
[![Release](https://img.shields.io/github/v/release/naufalmng/gc-rg?display_name=tag&sort=semver)](https://github.com/naufalmng/gc-rg/releases)
[![License](https://img.shields.io/github/license/naufalmng/gc-rg)](LICENSE)
[![Go](https://img.shields.io/badge/go-1.25-00ADD8.svg)](https://go.dev/)
[![Platform](https://img.shields.io/badge/platform-linux%20%7C%20windows-d70a53)](#install)

</div>

---

## English

`gc-rg` takes validated Grafana Cloud evidence, renders a daily monitoring report in Markdown + PDF, then sends the PDF through operator-owned SMTP. It is built as the reporting pair for [`gc-hc`](https://github.com/naufalmng/gc-hc): `gc-hc` checks connectivity, `gc-rg` packages the validated evidence into an operator-friendly daily report.

### Install

```bash
curl -fsSL https://github.com/naufalmng/gc-rg/releases/latest/download/gc-rg.sh | sudo bash
```

The installer builds a real `.deb`, installs the Linux release binaries, creates `/etc/gc-rg/gc-rg.env`, and places the daily systemd timer on disk. PDF generation requires `wkhtmltopdf`; the package declares it as an apt dependency.

### Quick start

```bash
./bin/gc-rg-generate --date today        # generate Markdown + PDF
./bin/gc-rg-generate --date today --no-pdf
./bin/gc-rg-email --date today --dry-run # validate report + SMTP + MIME only
./bin/gc-rg-email --date today --send    # send PDF attachment
```

For daily Linux automation, use the systemd templates in `assets/systemd/` and the environment example in `assets/env/gc-rg.env.example`.

For full usage, configuration reference, architecture, scheduler setup, and design notes, see **[documentation.md](documentation.md)**.

---

## Bahasa Indonesia

`gc-rg` ambil evidence Grafana Cloud yang sudah tervalidasi, render daily monitoring report dalam Markdown + PDF, lalu kirim PDF lewat SMTP milik operator. Tool ini pasangan reporting untuk [`gc-hc`](https://github.com/naufalmng/gc-hc): `gc-hc` ngecek konektivitas, `gc-rg` bungkus evidence tervalidasi jadi report harian yang enak dibaca operator.

### Instalasi

```bash
curl -fsSL https://github.com/naufalmng/gc-rg/releases/latest/download/gc-rg.sh | sudo bash
```

Installer bikin `.deb`, install binary Linux release, buat `/etc/gc-rg/gc-rg.env`, dan taruh systemd timer harian di disk. Generate PDF butuh `wkhtmltopdf`; package mendeklarasikan itu sebagai dependency apt.

### Quick start

```bash
./bin/gc-rg-generate --date today        # generate Markdown + PDF
./bin/gc-rg-generate --date today --no-pdf
./bin/gc-rg-email --date today --dry-run # validasi report + SMTP + MIME saja
./bin/gc-rg-email --date today --send    # kirim PDF attachment
```

Untuk otomasi harian di Linux, pakai template systemd di `assets/systemd/` dan contoh env di `assets/env/gc-rg.env.example`.

Untuk panduan lengkap, referensi konfigurasi, arsitektur, setup scheduler, dan catatan desain, lihat **[documentation.md](documentation.md)**.

---

## License / Lisensi

Apache License 2.0. See [LICENSE](LICENSE).

## Author / Penulis

[Muhammad Naufal Hanif](https://github.com/naufalmng) — built this so daily Grafana Cloud evidence can become a repeatable report, not a manual copy-paste ritual.
