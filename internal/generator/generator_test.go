package generator

import (
	"strings"
	"testing"
)

func TestOverallStatus_PrioritizesWarning(t *testing.T) {
	got := OverallStatus([]string{"✅", "⚠️", "ℹ️"})
	if got != "⚠️ Warning" {
		t.Fatalf("OverallStatus = %q", got)
	}
}

func TestRender_IncludesDataOnlySections(t *testing.T) {
	summary := Summary{Prometheus: map[string]Query{
		"availability_24h_node":  {Rows: []Row{{Labels: map[string]string{"instance": "node-a"}, Value: 100}}},
		"availability_24h_mysql": {Rows: []Row{{Value: 100}}},
		"cpu_max_24h":            {Rows: []Row{{Value: 10}}},
		"memory_max_24h":         {Rows: []Row{{Value: 20}}},
		"disk_max_24h":           {Rows: []Row{{Labels: map[string]string{"instance": "node-a", "mountpoint": "/"}, Value: 30}}},
		"mysql_conn_max_24h":     {Rows: []Row{{Value: 1}}},
	}, Loki: map[string]Query{"loki_mysql_lines_24h": {Rows: []Row{{Value: 0}}}}}

	content := Render("2026-06-05", summary, map[string]bool{"loki_mysql_stream": true}, nil)
	for _, want := range []string{"## Status Summary", "## Availability Summary", "## Logs and Error Summary", "Loki MySQL Logs | 24h lines 0 | ⚠️"} {
		if !strings.Contains(content, want) {
			t.Fatalf("rendered content missing %q", want)
		}
	}
}
