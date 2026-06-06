package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"gc-rg/internal/generator"
)

func main() {
	if err := run(os.Args[1:]); err != nil {
		fmt.Fprintf(os.Stderr, "ERROR: %v\n", err)
		os.Exit(1)
	}
}

func run(args []string) error {
	root := appRoot()
	options := generator.Options{
		Date:          time.Now().Format(time.DateOnly),
		OutputDir:     filepath.Join(root, "reports", "daily"),
		LongRangePath: filepath.Join(root, "evidence", "grafana-longrange-validation", "SUMMARY.json"),
		LatestPath:    filepath.Join(root, "evidence", "grafana-prometheus-validation", "SUMMARY.json"),
		LokiScopePath: filepath.Join(root, "evidence", "grafana-live-loki-scope-24h.json"),
	}
	flags := flag.NewFlagSet("generate-daily-report", flag.ContinueOnError)
	flags.StringVar(&options.Date, "date", options.Date, "report date in YYYY-MM-DD format")
	flags.StringVar(&options.OutputDir, "output-dir", options.OutputDir, "daily report output directory")
	flags.StringVar(&options.LongRangePath, "long-range-json", options.LongRangePath, "long-range evidence JSON path")
	flags.StringVar(&options.LatestPath, "latest-json", options.LatestPath, "latest evidence JSON path")
	flags.StringVar(&options.LokiScopePath, "loki-scope-json", options.LokiScopePath, "optional Loki scope JSON path")
	flags.BoolVar(&options.NoPDF, "no-pdf", false, "skip PDF conversion")
	if err := flags.Parse(args); err != nil {
		return fmt.Errorf("parse args: %w", err)
	}
	if options.Date == "today" {
		options.Date = time.Now().Format(time.DateOnly)
	}
	result, err := generator.Generate(options)
	if err != nil {
		return err
	}
	fmt.Println(result.MarkdownPath)
	if result.PDFPath != "" {
		fmt.Println(result.PDFPath)
	}
	return nil
}

func appRoot() string {
	workingDir, err := os.Getwd()
	if err == nil && hasEvidence(workingDir) {
		return workingDir
	}
	executablePath, err := os.Executable()
	if err == nil {
		executableDir := filepath.Dir(executablePath)
		if hasEvidence(executableDir) {
			return executableDir
		}
		parentDir := filepath.Dir(executableDir)
		if hasEvidence(parentDir) {
			return parentDir
		}
	}
	if workingDir != "" {
		return workingDir
	}
	return "."
}

func hasEvidence(root string) bool {
	path := filepath.Join(root, "evidence", "grafana-longrange-validation", "SUMMARY.json")
	_, err := os.Stat(path)
	return err == nil
}
