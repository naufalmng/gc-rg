package report

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

const defaultCaptionPrefix = "Daily Monitoring Report - POC Grafana Cloud Superindo (GO Edition)"

type Files struct {
	MarkdownPath string
	PDFPath      string
	PDFSize      int64
}

func Resolve(reportDir, date string) (Files, error) {
	markdownPath := filepath.Join(reportDir, fmt.Sprintf("%s-daily-monitoring-report.md", date))
	pdfPath := filepath.Join(reportDir, fmt.Sprintf("%s-daily-monitoring-report.pdf", date))
	if _, err := os.Stat(markdownPath); err != nil {
		return Files{}, fmt.Errorf("markdown report not found: %s", markdownPath)
	}
	info, err := os.Stat(pdfPath)
	if err != nil {
		return Files{}, fmt.Errorf("PDF report not found: %s", pdfPath)
	}
	return Files{MarkdownPath: markdownPath, PDFPath: pdfPath, PDFSize: info.Size()}, nil
}

func BuildCaption(markdownPath, date, override string) (string, error) {
	if strings.TrimSpace(override) != "" {
		return override, nil
	}
	status, err := ReadOverallStatus(markdownPath)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf(
		"%s\nPeriod: Last 24h\nDate: %s\nOverall Operational Status: %s",
		defaultCaptionPrefix,
		date,
		status,
	), nil
}

func ReadOverallStatus(markdownPath string) (string, error) {
	content, err := os.ReadFile(markdownPath)
	if err != nil {
		return "", fmt.Errorf("read markdown report: %w", err)
	}
	for _, line := range strings.Split(string(content), "\n") {
		if strings.HasPrefix(line, "**Overall Operational Status:**") {
			return strings.TrimSpace(strings.TrimPrefix(line, "**Overall Operational Status:**")), nil
		}
	}
	return "Unknown", nil
}
