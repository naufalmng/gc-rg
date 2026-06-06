package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"gc-rg/internal/config"
	"gc-rg/internal/report"
	"gc-rg/internal/whatsapp"
)

const appTimeout = 10 * time.Minute

func main() {
	if err := run(os.Args[1:]); err != nil {
		fmt.Fprintf(os.Stderr, "ERROR: %v\n", err)
		os.Exit(1)
	}
}

func run(args []string) error {
	options, err := config.ParseArgs(args, config.EnvMap())
	if err != nil {
		return err
	}
	root := appRoot()
	options.ReportDir = resolvePath(root, options.ReportDir)
	options.SessionDB = resolvePath(root, options.SessionDB)
	files, err := report.Resolve(options.ReportDir, options.Date)
	if err != nil {
		return err
	}
	caption, err := report.BuildCaption(files.MarkdownPath, options.Date, options.Caption)
	if err != nil {
		return err
	}
	printPlan(options, files, caption)
	if !options.Send && !options.LoginOnly {
		fmt.Println("dry_run_result=validated, not sent")
		return nil
	}
	ctx, cancel := context.WithTimeout(context.Background(), appTimeout)
	defer cancel()
	return whatsapp.SendReport(ctx, whatsapp.SendRequest{
		JID:       options.JID,
		PDFPath:   files.PDFPath,
		Caption:   caption,
		SessionDB: options.SessionDB,
		FreshAuth: options.FreshAuth,
		LoginOnly: options.LoginOnly,
	})
}

func appRoot() string {
	executablePath, err := os.Executable()
	if err == nil {
		executableDir := filepath.Dir(executablePath)
		if hasReports(executableDir) {
			return executableDir
		}
		parentDir := filepath.Dir(executableDir)
		if hasReports(parentDir) {
			return parentDir
		}
	}
	workingDir, err := os.Getwd()
	if err == nil && hasReports(workingDir) {
		return workingDir
	}
	if workingDir != "" {
		return workingDir
	}
	return "."
}

func hasReports(root string) bool {
	_, err := os.Stat(filepath.Join(root, "reports", "daily"))
	return err == nil
}

func resolvePath(root, value string) string {
	if filepath.IsAbs(value) {
		return value
	}
	return filepath.Join(root, value)
}

func printPlan(options config.Options, files report.Files, caption string) {
	mode := "dry-run"
	if options.Send {
		mode = "send"
	}
	if options.LoginOnly {
		mode = "login-only"
	}
	fmt.Printf("mode=%s\n", mode)
	fmt.Printf("jid=%s\n", options.JID)
	fmt.Printf("date=%s\n", options.Date)
	fmt.Printf("session_db=%s\n", options.SessionDB)
	fmt.Printf("attachment=%s\n", files.PDFPath)
	fmt.Printf("attachment_size=%d\n", files.PDFSize)
	fmt.Printf("caption_lines=%d\n", len(strings.Split(caption, "\n")))
}
