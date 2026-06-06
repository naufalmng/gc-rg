# EOS Lifecycle Coverage Draft Implementation Plan

> **For Hermes:** Use subagent-driven-development skill to implement this plan task-by-task after user approves the draft.

**Goal:** Add optional EOS/EOL lifecycle coverage to `gc-rg` so daily reports can show obsolete and approaching-obsolete assets using observed versions plus automatically fetched lifecycle calendars.

**Architecture:** Keep current `gc-rg` evidence-first design. A future collector can query Grafana API/Prometheus for current product versions, while report generation consumes a lifecycle evidence JSON file. EOS calendars come from `endoflife.date` with local cache fallback, not from Grafana itself.

**Tech Stack:** Go 1.25, existing `gc-rg` CLI layout, Grafana API/Prometheus query API for current observed versions, `endoflife.date` API for EOS calendars, Markdown/PDF report renderer, systemd scheduler.

---

## Draft status

This is **draft only**. Do not implement until approved.

## Problem

Grafana can observe **current state**: OS label, exporter version, agent version, database version, certificate expiry, etc.

Grafana usually cannot answer **latest EOS/EOL policy** by itself. EOS needs another source:

- `endoflife.date` for OS/database/runtime products.
- Local override rules for products not covered there.
- Hardcoded migration advisories for special cases like Grafana Agent → Alloy if external source is incomplete.

So correct model:

```text
Grafana API / Prometheus metrics = current observed product/version
endoflife.date / local rules     = lifecycle calendar
GC-RG comparison engine          = normal / approaching obsolete / obsolete
GC-RG report renderer            = lifecycle section in report
```

## Non-goals

- Do not make daily report fail just because EOS API is down.
- Do not require user to manually search EOS dates.
- Do not scrape random websites.
- Do not mix capacity risk with lifecycle risk.
- Do not send live SMTP during lifecycle tests.

## Proposed report section

Place after `## Database Health` and before `## Logs and Error Summary`.

```md
## Obsolescence / EOS Risk

| Asset | Product | Current | Latest Supported | EOS Date | Days Left | Status |
|---|---|---:|---:|---:|---:|---:|
| `linux-node-01` | Ubuntu | 22.04 | 24.04 LTS | 2027-04-01 | 300 | ✅ Normal |
| `db-node-01` | MySQL | 8.0 | 8.4 LTS | 2026-04-30 | -37 | ⛔ EOS |
| `agent-01` | Grafana Agent | 0.40 | Alloy migration recommended | 2025-11-01 | -219 | ⛔ EOS |
```

Status mapping:

| Status | Meaning |
| --- | --- |
| `normal` | Supported and outside warning window |
| `approaching_obsolete` | EOS within warning window |
| `obsolete` | EOS passed or version below minimum supported rule |
| `unknown` | Missing current version, unsupported product, API/cache unavailable |

Default thresholds:

```bash
GC_RG_LIFECYCLE_WARN_DAYS=90
GC_RG_LIFECYCLE_CRITICAL_DAYS=0
GC_RG_LIFECYCLE_CACHE_TTL=24h
```

## Evidence format draft

Input path:

```text
evidence/grafana-lifecycle-validation/SUMMARY.json
```

Minimal manual evidence:

```json
{
  "generated_at": "2026-06-06T10:00:00Z",
  "source": "manual",
  "items": [
    {
      "asset": "linux-node-01",
      "product": "ubuntu",
      "cycle": "22.04",
      "current": "22.04.4"
    },
    {
      "asset": "db-node-01",
      "product": "mysql",
      "cycle": "8.0",
      "current": "8.0.36"
    }
  ]
}
```

Rendered/enriched evidence candidate:

```json
{
  "generated_at": "2026-06-06T10:00:00Z",
  "source": "endoflife.date",
  "source_mode": "live",
  "items": [
    {
      "asset": "db-node-01",
      "product": "mysql",
      "cycle": "8.0",
      "current": "8.0.36",
      "latest_supported": "8.4",
      "eol_date": "2026-04-30",
      "days_left": -37,
      "status": "obsolete",
      "evidence": "manual:mysql:8.0"
    }
  ]
}
```

## Grafana API / metric detection draft

Phase 1 should not require Grafana API. Use manual lifecycle JSON + automatic EOS lookup.

Phase 2 collector can query Grafana API/Prometheus.

Useful metric candidates:

```promql
node_os_info
node_uname_info
node_exporter_build_info
mysqld_exporter_build_info
alloy_build_info
grafana_agent_build_info
prometheus_build_info
grafana_build_info
probe_ssl_earliest_cert_expiry
```

If distro/version is missing, recommend node textfile collector metric:

```text
node_os_info{name="ubuntu",version_id="22.04",pretty_name="Ubuntu 22.04.4 LTS"} 1
```

## File-by-file draft plan

### Task 1: Add lifecycle domain types

**Objective:** Create typed lifecycle model without touching renderer yet.

**Files:**

- Create: `internal/lifecycle/types.go`
- Test: `internal/lifecycle/types_test.go`

Types:

```go
type Status string

const (
    StatusNormal              Status = "normal"
    StatusApproachingObsolete Status = "approaching_obsolete"
    StatusObsolete            Status = "obsolete"
    StatusUnknown             Status = "unknown"
)

type InputItem struct {
    Asset   string `json:"asset"`
    Product string `json:"product"`
    Cycle   string `json:"cycle"`
    Current string `json:"current"`
}

type ReportItem struct {
    Asset           string `json:"asset"`
    Product         string `json:"product"`
    Cycle           string `json:"cycle"`
    Current         string `json:"current"`
    LatestSupported string `json:"latest_supported"`
    EOLDate         string `json:"eol_date"`
    DaysLeft        *int   `json:"days_left"`
    Status          Status `json:"status"`
    Evidence        string `json:"evidence"`
}
```

### Task 2: Parse optional lifecycle JSON

**Objective:** Load lifecycle evidence if file exists; skip if missing.

**Files:**

- Create: `internal/lifecycle/load.go`
- Test: `internal/lifecycle/load_test.go`
- Modify: `cmd/generate-daily-report/main.go`

Behavior:

- Missing lifecycle file = no error, empty lifecycle section.
- Invalid lifecycle JSON = error with path context.
- Empty items = no lifecycle section.

### Task 3: Add endoflife.date client with cache

**Objective:** Fetch product lifecycle calendar and cache response locally.

**Files:**

- Create: `internal/lifecycle/eol_client.go`
- Create: `internal/lifecycle/cache.go`
- Test: `internal/lifecycle/eol_client_test.go`
- Test: `internal/lifecycle/cache_test.go`

Endpoint pattern:

```text
https://endoflife.date/api/{product}.json
```

Cache path default:

```text
.cache/gc-rg/eol/{product}.json
```

Rules:

- Live fetch success: use live, update cache.
- Live fetch fails and cache exists: use cache, mark `source_mode=cached`.
- Live fetch fails and no cache: mark affected items `unknown`; report still generated.

### Task 4: Compare current cycle to EOS calendar

**Objective:** Compute days left and status.

**Files:**

- Create: `internal/lifecycle/evaluate.go`
- Test: `internal/lifecycle/evaluate_test.go`

Rules:

```text
if no matching cycle -> unknown
if eol date missing/false -> normal
if days_left <= critical_days -> obsolete
if days_left <= warn_days -> approaching_obsolete
else normal
```

Default:

```text
warn_days=90
critical_days=0
```

### Task 5: Render lifecycle section in report

**Objective:** Add optional section after Database Health.

**Files:**

- Modify: `internal/generator/generator.go`
- Test: `internal/generator/generator_test.go`

Behavior:

- If lifecycle item count is zero, section omitted.
- If present, render table with status icon.
- Overall operational status should optionally include lifecycle status. Draft decision: include obsolete as ⛔ and approaching obsolete as ⚠️.

### Task 6: Add CLI flags and env defaults

**Objective:** Add lifecycle config knobs without breaking existing commands.

**Files:**

- Modify: `cmd/generate-daily-report/main.go`
- Modify: `assets/env/gc-rg.env.example`
- Modify: `scripts/run-daily-email.sh`
- Modify: `scripts/build.sh` embedded env sample

Flags:

```bash
--lifecycle-json evidence/grafana-lifecycle-validation/SUMMARY.json
--lifecycle-warn-days 90
--lifecycle-critical-days 0
--lifecycle-cache-dir .cache/gc-rg/eol
--no-lifecycle
```

Env:

```bash
GC_RG_LIFECYCLE_ENABLED=true
GC_RG_LIFECYCLE_WARN_DAYS=90
GC_RG_LIFECYCLE_CRITICAL_DAYS=0
GC_RG_LIFECYCLE_CACHE_DIR=/opt/gc-rg/.cache/eol
```

### Task 7: Draft collector command for later phase

**Objective:** Add only command skeleton or separate future plan, not full Grafana API collector yet.

**Files:**

- Create later: `cmd/collect-lifecycle/main.go`
- Create later: `internal/grafana/client.go`

Draft command:

```bash
gc-rg-lifecycle collect \
  --grafana-url https://example.grafana.net \
  --token-env GC_RG_GRAFANA_TOKEN \
  --datasource-uid prometheus_uid \
  --output evidence/grafana-lifecycle-validation/SUMMARY.json
```

### Task 8: Installer/release integration

**Objective:** If collector is added, installer and release publish it too.

**Files:**

- Modify: `scripts/build.sh`
- Modify: `.github/workflows/release.yml`
- Modify: `.github/workflows/ci.yml`
- Modify: `tests/installer_smoke.sh`

Artifacts candidate:

```text
gc-rg-lifecycle-linux-amd64
gc-rg-lifecycle-windows-amd64.exe
```

## Acceptance criteria

- Existing daily report generation still passes unchanged when no lifecycle file exists.
- Lifecycle section renders when lifecycle evidence exists.
- EOS date lookup is automatic through `endoflife.date` and cache fallback.
- Report does not fail when external EOS API is down and cache exists.
- Missing cache + API down marks lifecycle status `unknown`, not fatal.
- CI covers load, evaluate, render, cache fallback, and no-lifecycle default behavior.

## Open questions

1. Which products must be supported first?
   - Candidate: Ubuntu, Debian, MySQL, Grafana, Go, Node.js, Python.
2. Does lifecycle status affect overall report status?
   - Suggested: obsolete = ⛔, approaching = ⚠️, unknown = ℹ️.
3. Should Grafana Agent be treated as fixed local advisory or fetched source?
   - Suggested: local advisory first, because Agent → Alloy lifecycle nuance may not map cleanly.
4. Should certificate expiry live in lifecycle section or separate security/certificate section?
   - Suggested: separate later, unless user wants one combined risk section.

## Recommended implementation order

1. Manual lifecycle JSON + renderer section.
2. endoflife.date lookup + cache.
3. Optional overall status integration.
4. Docs + template examples.
5. Grafana API collector as separate approved phase.
