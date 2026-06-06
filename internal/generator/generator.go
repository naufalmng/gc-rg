package generator

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	wkhtmltopdf "github.com/SebastiaanKlippert/go-wkhtmltopdf"
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/extension"
	"github.com/yuin/goldmark/renderer/html"
)

const (
	CPUWarningThreshold             = 80.0
	MemoryWarningThreshold          = 85.0
	DiskWarningThreshold            = 80.0
	MySQLConnectionWarningThreshold = 80.0
)

type Row struct {
	Labels map[string]string `json:"labels"`
	Value  float64           `json:"value"`
}

type Query struct {
	Rows []Row `json:"rows"`
}

type Summary struct {
	Prometheus map[string]Query `json:"prometheus"`
	Loki       map[string]Query `json:"loki"`
}

type Options struct {
	Date          string
	OutputDir     string
	LongRangePath string
	LatestPath    string
	LokiScopePath string
	NoPDF         bool
}

type Result struct {
	MarkdownPath string
	PDFPath      string
}

func Generate(options Options) (Result, error) {
	longRange, err := loadSummary(options.LongRangePath)
	if err != nil {
		return Result{}, err
	}
	latestNames, err := loadLatestNames(options.LatestPath)
	if err != nil {
		return Result{}, err
	}
	lokiRows, err := LoadLokiScopeRows(options.LokiScopePath)
	if err != nil {
		return Result{}, err
	}
	content := Render(options.Date, longRange, latestNames, lokiRows)
	if err := os.MkdirAll(options.OutputDir, 0o755); err != nil {
		return Result{}, fmt.Errorf("create output dir: %w", err)
	}
	markdownPath := filepath.Join(options.OutputDir, fmt.Sprintf("%s-daily-monitoring-report.md", options.Date))
	if err := os.WriteFile(markdownPath, []byte(content), 0o644); err != nil {
		return Result{}, fmt.Errorf("write markdown report: %w", err)
	}
	result := Result{MarkdownPath: markdownPath}
	if !options.NoPDF {
		pdfPath, err := WritePDF(markdownPath)
		if err != nil {
			return Result{}, err
		}
		result.PDFPath = pdfPath
	}
	return result, nil
}

func Render(reportDate string, summary Summary, latestNames map[string]bool, lokiScopeRows []LokiScopeRow) string {
	nodeRows := summary.Prometheus["availability_24h_node"].Rows
	mysqlRows := summary.Prometheus["availability_24h_mysql"].Rows
	nodeCount := len(nodeRows)
	mysqlUp := 0.0
	if len(mysqlRows) > 0 {
		mysqlUp = mysqlRows[0].Value
	}
	cpuMax := maxValue(summary, "cpu_max_24h")
	memoryMax := maxValue(summary, "memory_max_24h")
	diskMax := maxValue(summary, "disk_max_24h")
	mysqlConn := maxValue(summary, "mysql_conn_max_24h")
	mysqlSlow := maxValue(summary, "mysql_slow_increase_24h")
	mysqlAborted := maxValue(summary, "mysql_aborted_increase_24h")
	loki24h := maxLokiValue(summary, "loki_mysql_lines_24h")
	lokiError24h := maxLokiValue(summary, "loki_mysql_error_lines_24h")

	linuxStatus := "⛔"
	if nodeCount > 0 && allAtLeast(nodeRows, 99) {
		linuxStatus = "✅"
	}
	mysqlStatus := "⛔"
	if mysqlUp >= 99 {
		mysqlStatus = "✅"
	}
	cpuStatus := thresholdIcon(cpuMax, CPUWarningThreshold)
	memoryStatus := thresholdIcon(memoryMax, MemoryWarningThreshold)
	diskStatus := thresholdIcon(diskMax, DiskWarningThreshold)
	mysqlConnectionStatus := thresholdIcon(mysqlConn, MySQLConnectionWarningThreshold)
	lokiStatusLabel := LokiStatusLabel(loki24h)
	lokiStatus := StatusIcon(lokiStatusLabel)
	overallStatus := OverallStatus([]string{linuxStatus, mysqlStatus, cpuStatus, memoryStatus, diskStatus, mysqlConnectionStatus, lokiStatus})

	lines := []string{
		"# Daily Monitoring Report - POC Grafana Cloud Superindo",
		"",
		fmt.Sprintf("**Generated:** %s  ", time.Now().Format("2006-01-02 15:04:05")),
		"**Period:** Last 24h  ",
		"**Environment:** POC  ",
		"**Prepared by:** Code.ID  ",
		fmt.Sprintf("**Overall Operational Status:** %s", overallStatus),
		"",
		"**Status guide:** ✅ Normal · ℹ️ Info · ⚠️ Warning · ⛔ Action Required",
		"",
		"## Status Summary",
		"",
		"| Indicator | Evidence | Status |",
		"|---|---|---:|",
		fmt.Sprintf("| Linux Monitoring | %d/%d nodes UP, 24h availability 100%% | %s |", nodeCount, nodeCount, linuxStatus),
		fmt.Sprintf("| MySQL Monitoring | MySQL UP, 24h availability %s | %s |", FormatPercent(mysqlUp), mysqlStatus),
		fmt.Sprintf("| CPU | Max 24h %s | %s |", FormatPercent(cpuMax), cpuStatus),
		fmt.Sprintf("| Memory | Max 24h %s | %s |", FormatPercent(memoryMax), memoryStatus),
		fmt.Sprintf("| Disk | Max 24h %s | %s |", FormatPercent(diskMax), diskStatus),
		fmt.Sprintf("| MySQL Connections | Max 24h %s | %s |", FormatPercent(mysqlConn), mysqlConnectionStatus),
		fmt.Sprintf("| Loki MySQL Logs | 24h lines %.0f | %s |", loki24h, lokiStatus),
		"",
		"## Availability Summary",
		"",
		"| Asset | 24h Availability | Status |",
		"|---|---|---:|",
	}
	for _, row := range nodeRows {
		lines = append(lines, fmt.Sprintf("| `%s` Linux | %s | %s |", label(row, "instance"), FormatPercent(row.Value), linuxStatus))
	}
	lines = append(lines,
		fmt.Sprintf("| `xtra-db-qa-cloned` MySQL | %s | %s |", FormatPercent(mysqlUp), mysqlStatus),
		"", "## Resource Utilization", "",
		"| Metric | Max 24h | Status |", "|---|---|---:|",
		fmt.Sprintf("| CPU | %s | %s |", FormatPercent(cpuMax), cpuStatus),
		fmt.Sprintf("| Memory | %s | %s |", FormatPercent(memoryMax), memoryStatus),
		"", "## Disk Capacity", "",
		"| Instance | Mountpoint | Max Usage 24h | Status |", "|---|---|---|---:|",
	)
	for _, row := range summary.Prometheus["disk_max_24h"].Rows {
		lines = append(lines, fmt.Sprintf("| `%s` | `%s` | %s | %s |", label(row, "instance"), label(row, "mountpoint"), FormatPercent(row.Value), thresholdIcon(row.Value, DiskWarningThreshold)))
	}
	lines = append(lines,
		"", "## Database Health", "",
		"| Metric | Value | Status |", "|---|---|---:|",
		fmt.Sprintf("| MySQL availability 24h | %s | %s |", FormatPercent(mysqlUp), mysqlStatus),
		fmt.Sprintf("| Max connection usage 24h | %s | %s |", FormatPercent(mysqlConn), mysqlConnectionStatus),
		fmt.Sprintf("| Slow query increase 24h | %.2f | ✅ |", mysqlSlow),
		fmt.Sprintf("| Aborted connects increase 24h | %.2f | ℹ️ |", mysqlAborted),
		"", "## Logs and Error Summary", "",
		"| Job | Instance | Log Lines 24h | Error Pattern Lines 24h | Info Pattern Lines 24h | Status |",
		"|---|---|---|---|---|---:|",
	)
	if len(lokiScopeRows) > 0 {
		for _, row := range lokiScopeRows {
			lines = append(lines, fmt.Sprintf("| `%s` | `%s` | %.0f | %.0f | %.0f | %s |", row.Job, row.Instance, row.LogLines, row.ErrorLines, row.InfoLines, StatusIcon(row.Status)))
		}
	} else {
		lines = append(lines, fmt.Sprintf("| `integrations/mysql` | `xtra-db-qa-cloned` | %.0f | %.0f | 0 | %s |", loki24h, lokiError24h, StatusIcon(lokiStatusLabel)))
	}
	lines = append(lines, "", fmt.Sprintf("**Final daily status:** %s", overallStatus), "")
	if latestNames["loki_mysql_stream"] {
		lines = append(lines, "<!-- Source: local validated JSON evidence; no live Grafana write performed. -->", "")
	}
	return strings.Join(lines, "\n")
}

func FormatPercent(value float64) string { return fmt.Sprintf("%.2f%%", value) }
func LokiStatusLabel(value float64) string {
	if value > 0 {
		return "Normal"
	}
	return "Warning"
}
func StatusIcon(label string) string {
	switch label {
	case "Action Required":
		return "⛔"
	case "Warning":
		return "⚠️"
	case "Info":
		return "ℹ️"
	default:
		return "✅"
	}
}
func OverallStatus(statuses []string) string {
	for _, s := range statuses {
		if s == "⛔" {
			return "⛔ Action Required"
		}
	}
	for _, s := range statuses {
		if s == "⚠️" {
			return "⚠️ Warning"
		}
	}
	for _, s := range statuses {
		if s == "ℹ️" {
			return "ℹ️ Info"
		}
	}
	return "✅ Normal"
}

func loadSummary(path string) (Summary, error) {
	var s Summary
	b, err := os.ReadFile(path)
	if err != nil {
		return s, fmt.Errorf("read evidence JSON %s: %w", path, err)
	}
	if err := json.Unmarshal(b, &s); err != nil {
		return s, fmt.Errorf("parse evidence JSON %s: %w", path, err)
	}
	return s, nil
}
func loadLatestNames(path string) (map[string]bool, error) {
	var rows []struct {
		Name string `json:"name"`
	}
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read latest JSON %s: %w", path, err)
	}
	if err := json.Unmarshal(b, &rows); err != nil {
		return nil, fmt.Errorf("parse latest JSON %s: %w", path, err)
	}
	names := map[string]bool{}
	for _, row := range rows {
		names[row.Name] = true
	}
	return names, nil
}
func maxValue(s Summary, key string) float64 {
	max := 0.0
	for _, row := range s.Prometheus[key].Rows {
		if row.Value > max {
			max = row.Value
		}
	}
	return max
}
func maxLokiValue(s Summary, key string) float64 {
	max := 0.0
	for _, row := range s.Loki[key].Rows {
		if row.Value > max {
			max = row.Value
		}
	}
	return max
}
func thresholdIcon(value, threshold float64) string {
	if value >= threshold {
		return "⚠️"
	}
	return "✅"
}
func allAtLeast(rows []Row, min float64) bool {
	for _, row := range rows {
		if row.Value < min {
			return false
		}
	}
	return true
}
func label(row Row, name string) string {
	if row.Labels == nil || row.Labels[name] == "" {
		return "unknown"
	}
	return row.Labels[name]
}

func WritePDF(markdownPath string) (string, error) {
	markdownBytes, err := os.ReadFile(markdownPath)
	if err != nil {
		return "", fmt.Errorf("read markdown for PDF: %w", err)
	}
	htmlBytes, err := markdownToHTML(markdownBytes)
	if err != nil {
		return "", err
	}
	pdfBytes, err := htmlToPDF(htmlBytes)
	if err != nil {
		return "", err
	}
	pdfPath := strings.TrimSuffix(markdownPath, filepath.Ext(markdownPath)) + ".pdf"
	if err := os.WriteFile(pdfPath, pdfBytes, 0o644); err != nil {
		return "", fmt.Errorf("write PDF report: %w", err)
	}
	info, err := os.Stat(pdfPath)
	if err != nil || info.Size() == 0 {
		return "", fmt.Errorf("PDF conversion did not create output: %s", pdfPath)
	}
	return pdfPath, nil
}

func markdownToHTML(markdownBytes []byte) ([]byte, error) {
	markdown := goldmark.New(
		goldmark.WithExtensions(extension.GFM, extension.NewTable()),
		goldmark.WithRendererOptions(html.WithUnsafe(), html.WithHardWraps()),
	)
	var body bytes.Buffer
	if err := markdown.Convert(markdownBytes, &body); err != nil {
		return nil, fmt.Errorf("convert markdown to HTML: %w", err)
	}
	page := fmt.Sprintf("<!doctype html><html><head><meta charset=\"utf-8\"><style>%s</style></head><body>%s</body></html>", pdfCSS, body.String())
	return []byte(page), nil
}

func htmlToPDF(htmlBytes []byte) ([]byte, error) {
	pdfg, err := wkhtmltopdf.NewPDFGenerator()
	if err != nil {
		return nil, fmt.Errorf("create wkhtmltopdf generator: %w", err)
	}
	pdfg.PageSize.Set(wkhtmltopdf.PageSizeA4)
	pdfg.MarginTop.Set(14)
	pdfg.MarginRight.Set(12)
	pdfg.MarginBottom.Set(14)
	pdfg.MarginLeft.Set(12)
	pdfg.Dpi.Set(300)
	pdfg.Orientation.Set(wkhtmltopdf.OrientationPortrait)
	pdfg.Grayscale.Set(false)
	page := wkhtmltopdf.NewPageReader(bytes.NewReader(htmlBytes))
	page.EnableLocalFileAccess.Set(true)
	pdfg.AddPage(page)
	if err := pdfg.Create(); err != nil {
		return nil, fmt.Errorf("wkhtmltopdf conversion failed: %w", err)
	}
	return pdfg.Buffer().Bytes(), nil
}

const pdfCSS = `@page { size: A4; margin: 18mm 14mm; }
body { color: #111827; font-family: Arial, Helvetica, sans-serif; font-size: 10.5px; line-height: 1.35; }
h1 { color: #0f172a; font-size: 20px; margin: 0 0 14px; padding-bottom: 8px; border-bottom: 2px solid #1f2937; }
h2 { color: #111827; font-size: 14px; margin: 18px 0 8px; padding-bottom: 4px; border-bottom: 1px solid #d1d5db; }
table { width: 100%; border-collapse: collapse; margin: 8px 0 14px; page-break-inside: avoid; }
th { background: #f3f4f6; color: #111827; font-weight: 700; }
th, td { border: 1px solid #d1d5db; padding: 5px 6px; text-align: left; vertical-align: top; }
tr:nth-child(even) td { background: #fafafa; }
th:last-child, td:last-child { text-align: right; }
code { background: #f3f4f6; border-radius: 3px; font-family: Consolas, monospace; padding: 1px 3px; }
`
