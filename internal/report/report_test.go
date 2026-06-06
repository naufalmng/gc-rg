package report

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestResolve_ReturnsReportFiles(t *testing.T) {
	dir := t.TempDir()
	markdownPath := filepath.Join(dir, "2026-06-05-daily-monitoring-report.md")
	pdfPath := filepath.Join(dir, "2026-06-05-daily-monitoring-report.pdf")
	if err := os.WriteFile(markdownPath, []byte("# Report"), 0o600); err != nil {
		t.Fatalf("write markdown: %v", err)
	}
	if err := os.WriteFile(pdfPath, []byte("%PDF"), 0o600); err != nil {
		t.Fatalf("write pdf: %v", err)
	}

	files, err := Resolve(dir, "2026-06-05")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if files.PDFSize != 4 {
		t.Fatalf("PDFSize = %d, want 4", files.PDFSize)
	}
}

func TestBuildCaption_UsesOverallStatus(t *testing.T) {
	dir := t.TempDir()
	markdownPath := filepath.Join(dir, "report.md")
	content := "# Report\n**Overall Operational Status:** 🟢 Normal\n"
	if err := os.WriteFile(markdownPath, []byte(content), 0o600); err != nil {
		t.Fatalf("write markdown: %v", err)
	}

	caption, err := BuildCaption(markdownPath, "2026-06-05", "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(caption, "Overall Operational Status: 🟢 Normal") {
		t.Fatalf("caption missing status: %q", caption)
	}
}

func TestBuildCaption_UsesOverride(t *testing.T) {
	caption, err := BuildCaption("missing.md", "2026-06-05", "custom")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if caption != "custom" {
		t.Fatalf("caption = %q, want custom", caption)
	}
}
