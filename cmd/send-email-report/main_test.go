package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestRun_DryRunValidatesWithoutSending(t *testing.T) {
	dir := t.TempDir()
	reportDir := filepath.Join(dir, "reports")
	if err := os.MkdirAll(reportDir, 0o700); err != nil {
		t.Fatalf("mkdir report dir: %v", err)
	}
	date := "2026-06-05"
	markdownPath := filepath.Join(reportDir, date+"-daily-monitoring-report.md")
	pdfPath := filepath.Join(reportDir, date+"-daily-monitoring-report.pdf")
	if err := os.WriteFile(markdownPath, []byte("# Report\n**Overall Operational Status:** 🟢 Normal\n"), 0o600); err != nil {
		t.Fatalf("write markdown: %v", err)
	}
	if err := os.WriteFile(pdfPath, []byte("%PDF"), 0o600); err != nil {
		t.Fatalf("write pdf: %v", err)
	}
	env := map[string]string{
		"GC_RG_EMAIL_PROVIDER": "gmail",
		"GC_RG_EMAIL_FROM":     "sender@gmail.com",
		"GC_RG_EMAIL_TO":       "ops@example.com",
		"GC_RG_SMTP_USERNAME":  "sender@gmail.com",
		"GC_RG_SMTP_PASSWORD":  "app-password",
	}

	err := run([]string{"--date", date, "--report-dir", reportDir, "--dry-run"}, env)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}


func TestRun_SendAndDryRunConflict(t *testing.T) {
	dir := t.TempDir()
	reportDir := filepath.Join(dir, "reports")
	if err := os.MkdirAll(reportDir, 0o700); err != nil {
		t.Fatalf("mkdir report dir: %v", err)
	}
	env := map[string]string{
		"GC_RG_EMAIL_PROVIDER": "gmail",
		"GC_RG_EMAIL_FROM":     "sender@gmail.com",
		"GC_RG_EMAIL_TO":       "ops@example.com",
		"GC_RG_SMTP_USERNAME":  "sender@gmail.com",
		"GC_RG_SMTP_PASSWORD":  "app-password",
	}

	err := run([]string{"--report-dir", reportDir, "--send", "--dry-run"}, env)
	if err == nil {
		t.Fatal("expected conflict error")
	}
}
