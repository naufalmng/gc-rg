# GC-RG GC-HC Parity Alignment Implementation Plan

> **For Hermes:** Implement task-by-task with TDD. Reference repo: `gc-hc` from `https://github.com/naufalmng/gc-hc`.

**Goal:** Make `gc-rg` feel like the same product family as `gc-hc`, so users familiar with `gc-hc` can operate `gc-rg` with the same command rhythm.

**Architecture:** Keep existing Go binaries for report generation and SMTP sending. Add a shell runtime wrapper `gc-rg` plus short alias `gcrg` that provides `onboard`, `config`, `generate`, `send`, `run`, `status`, `logs`, `enable`, `disable`, and `remove`. Make installer and systemd call the unified command.

**Tech Stack:** Bash wrapper, existing Go binaries, systemd service/timer, Debian package installer, existing Go tests + shell smoke tests.

---

## Reference behavior from gc-hc

`gc-hc` user flow:

```bash
sudo gc-hc onboard
gchc check
gchc status
gchc config show
sudo gc-hc config smtp
gchc logs
sudo gc-hc enable
sudo gc-hc disable
```

Target `gc-rg` user flow:

```bash
sudo gc-rg onboard
gcrg generate
gcrg send --dry-run
gcrg run
gcrg status
gcrg config show
sudo gc-rg config smtp
gcrg logs
sudo gc-rg enable
sudo gc-rg disable
```

## Task 1: Add failing installer smoke expectations

**Objective:** Prove the current installer is not yet gc-hc-style.

**Files:**

- Modify: `tests/installer_smoke.sh`

**Assertions to add first:**

- Installer help mentions `sudo gc-rg onboard`.
- Installer help mentions short command `gcrg`.
- Installer embeds/installs unified command `gc-rg`.
- Installer embeds/installs short alias `gcrg`.
- systemd service calls `gc-rg run --quiet`, not direct generate/email binaries.

**RED command:**

```bash
bash scripts/build.sh && bash tests/installer_smoke.sh
```

Expected: FAIL before wrapper/installer update.

## Task 2: Add unified shell wrapper runtime

**Objective:** Create a `gc-rg` shell runtime similar in spirit to `gc-hc`.

**Files:**

- Create: `src/tool/00-header.sh`
- Create: `src/tool/01-globals.sh`
- Create: `src/tool/02-logging.sh`
- Create: `src/tool/03-utils.sh`
- Create: `src/tool/04-cli.sh`
- Create: `src/tool/05-config.sh`
- Create: `src/tool/06-report.sh`
- Create: `src/tool/07-systemd.sh`
- Create: `src/tool/08-status.sh`
- Create: `src/tool/99-main.sh`

Minimum commands:

```text
gc-rg onboard
gc-rg config
gc-rg config show
gc-rg config smtp
gc-rg generate
gc-rg send
gc-rg run
gc-rg status
gc-rg logs
gc-rg enable
gc-rg disable
gc-rg remove
gc-rg help
gc-rg version
```

Aliases:

```text
gcrg = gc-rg
```

**GREEN command:**

```bash
bash scripts/build.sh && bash -n dist/gc-rg && dist/gc-rg help
```

## Task 3: Update installer to build/package wrapper

**Objective:** Install `/usr/bin/gc-rg` and `/usr/bin/gcrg`, while keeping Go binaries in `/opt/gc-rg/bin`.

**Files:**

- Modify: `scripts/build.sh`
- Modify: `assets/systemd/gc-rg.service`
- Modify: `assets/systemd/gc-rg.timer` only if needed

Package layout:

```text
/usr/bin/gc-rg
/usr/bin/gcrg -> /usr/bin/gc-rg
/opt/gc-rg/bin/gc-rg-generate
/opt/gc-rg/bin/gc-rg-email
/etc/gc-rg/gc-rg.env
```

Systemd target:

```ini
ExecStart=/usr/bin/gc-rg run --quiet
```

## Task 4: Add config UX parity

**Objective:** Replace manual-only env editing with guided config commands.

**Files:**

- Modify: `src/tool/05-config.sh`

Behaviors:

- `gc-rg config` prompts core report config.
- `gc-rg config smtp` prompts SMTP settings.
- `gc-rg config show` prints masked config.
- Existing `/etc/gc-rg/gc-rg.env` is backed up before overwrite.

## Task 5: Add status/logs/enable/disable parity

**Objective:** Provide the same operational muscle memory as `gc-hc`.

**Files:**

- Modify: `src/tool/07-systemd.sh`
- Modify: `src/tool/08-status.sh`

Behaviors:

- `status` shows timer/service state and latest report files.
- `logs` runs `journalctl -u gc-rg.service -n 100 --no-pager` or follows with flag later.
- `enable` runs `systemctl daemon-reload` + `systemctl enable --now gc-rg.timer`.
- `disable` disables/stops timer and resets failed unit state.

## Task 6: Update docs and README vibe

**Objective:** Make README quick start match gc-hc style.

**Files:**

- Modify: `README.md`
- Modify: `documentation.md`

Target quick start:

```bash
sudo gc-rg onboard       # configure + enable timer + first report dry-run
gcrg generate            # generate Markdown + PDF on demand
gcrg send --dry-run      # validate report + SMTP + MIME
gcrg run                 # generate + send
gcrg status              # see timer and latest report state
gcrg config show         # show config with secrets masked
sudo gc-rg config smtp   # configure SMTP delivery
gcrg --help              # full command reference
sudo apt-get remove gc-rg
```

## Task 7: Full verification

Run:

```bash
go test ./...
bash scripts/build.sh
bash tests/installer_smoke.sh
bash -n dist/gc-rg.sh
bash -n dist/gc-rg
dist/gc-rg help
dist/gc-rg version
```

Pass criteria:

- Existing Go tests still pass.
- Installer smoke passes.
- Help output exposes gc-hc-like commands.
- `gc-rg` wrapper can run in standalone/dev mode without root for help/version/generate/send dry-run.
- Systemd service calls unified wrapper.
