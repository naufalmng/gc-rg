# Daily Monitoring Report - POC Grafana Cloud Superindo

**Generated:** 2026-06-05 13:56:56  
**Period:** Last 24h  
**Environment:** POC  
**Prepared by:** Code.ID  
**Overall Operational Status:** ⚠️ Warning

**Status guide:** ✅ Normal · ℹ️ Info · ⚠️ Warning · ⛔ Action Required

## Status Summary

| Indicator | Evidence | Status |
|---|---|---:|
| Linux Monitoring | 2/2 nodes UP, 24h availability 100% | ✅ |
| MySQL Monitoring | MySQL UP, 24h availability 100.00% | ✅ |
| CPU | Max 24h 10.07% | ✅ |
| Memory | Max 24h 44.85% | ✅ |
| Disk | Max 24h 50.72% | ✅ |
| MySQL Connections | Max 24h 1.32% | ✅ |
| Loki MySQL Logs | 24h lines 0 | ⚠️ |

## Availability Summary

| Asset | 24h Availability | Status |
|---|---|---:|
| `xtra-db-qa-cloned` Linux | 100.00% | ✅ |
| `xtra-qa-newtech` Linux | 100.00% | ✅ |
| `xtra-db-qa-cloned` MySQL | 100.00% | ✅ |

## Resource Utilization

| Metric | Max 24h | Status |
|---|---|---:|
| CPU | 10.07% | ✅ |
| Memory | 44.85% | ✅ |

## Disk Capacity

| Instance | Mountpoint | Max Usage 24h | Status |
|---|---|---|---:|
| `xtra-qa-newtech` | `/` | 50.72% | ✅ |
| `xtra-db-qa-cloned` | `/` | 19.09% | ✅ |
| `xtra-db-qa-cloned` | `/boot/efi` | 5.00% | ✅ |
| `xtra-qa-newtech` | `/boot/efi` | 5.85% | ✅ |
| `xtra-qa-newtech` | `/boot` | 20.71% | ✅ |

## Database Health

| Metric | Value | Status |
|---|---|---:|
| MySQL availability 24h | 100.00% | ✅ |
| Max connection usage 24h | 1.32% | ✅ |
| Slow query increase 24h | 0.00 | ✅ |
| Aborted connects increase 24h | 1.00 | ℹ️ |

## Logs and Error Summary

| Job | Instance | Log Lines 24h | Error Pattern Lines 24h | Info Pattern Lines 24h | Status |
|---|---|---|---|---|---:|
| `gc-hc` | `-` | 576 | 0 | 0 | ℹ️ |
| `integrations/mysql` | `xtra-db-qa-cloned` | 15 | 1 | 0 | ⚠️ |
| `integrations/node_exporter` | `xtra-db-qa-cloned` | 3870 | 276 | 1648 | ⚠️ |
| `integrations/node_exporter` | `xtra-qa-newtech` | 37494 | 388 | 3688 | ⚠️ |

**Final daily status:** ⚠️ Warning

<!-- Source: local validated JSON evidence; no live Grafana write performed. -->
