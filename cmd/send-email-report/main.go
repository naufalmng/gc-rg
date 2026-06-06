package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"gc-rg/internal/email"
	"gc-rg/internal/report"
)

const defaultEmailBody = "Daily monitoring report attached."

func main() {
	if err := run(os.Args[1:], email.EnvMap()); err != nil {
		fmt.Fprintf(os.Stderr, "ERROR: %v\n", err)
		os.Exit(1)
	}
}

func run(args []string, env map[string]string) error {
	root := appRoot()
	date := time.Now().Format(time.DateOnly)
	reportDir := filepath.Join(root, "reports", "daily")
	dryRun := false
	sendNow := false
	emailArgs := make([]string, 0, len(args))
	flags := flag.NewFlagSet("send-email-report", flag.ContinueOnError)
	flags.StringVar(&date, "date", date, "report date in YYYY-MM-DD format")
	flags.StringVar(&reportDir, "report-dir", reportDir, "daily report directory")
	flags.BoolVar(&dryRun, "dry-run", false, "validate email without sending")
	flags.BoolVar(&sendNow, "send", false, "send email report")
	flags.Func("email-provider", "SMTP provider", collectValue(&emailArgs, "email-provider"))
	flags.Func("smtp-host", "SMTP host", collectValue(&emailArgs, "smtp-host"))
	flags.Func("smtp-port", "SMTP port", collectValue(&emailArgs, "smtp-port"))
	flags.Func("smtp-tls", "SMTP TLS", collectValue(&emailArgs, "smtp-tls"))
	flags.Func("email-from", "email from", collectValue(&emailArgs, "email-from"))
	flags.Func("email-to", "email recipients", collectValue(&emailArgs, "email-to"))
	flags.Func("email-cc", "email CC", collectValue(&emailArgs, "email-cc"))
	flags.Func("smtp-username", "SMTP username", collectValue(&emailArgs, "smtp-username"))
	flags.Func("smtp-password", "SMTP password", collectValue(&emailArgs, "smtp-password"))
	flags.Func("email-subject-prefix", "email subject prefix", collectValue(&emailArgs, "email-subject-prefix"))
	if err := flags.Parse(args); err != nil {
		return fmt.Errorf("parse args: %w", err)
	}
	if date == "today" {
		date = time.Now().Format(time.DateOnly)
	}
	if !dryRun && !sendNow {
		dryRun = true
	}
	config, err := email.ParseConfig(emailArgs, env)
	if err != nil {
		return err
	}
	files, err := report.Resolve(reportDir, date)
	if err != nil {
		return err
	}
	status, err := report.ReadOverallStatus(files.MarkdownPath)
	if err != nil {
		return err
	}
	request := email.MessageRequest{
		From:        config.From,
		To:          config.To,
		CC:          config.CC,
		Subject:     fmt.Sprintf("%s Daily Monitoring Report %s", config.SubjectPrefix, date),
		Body:        fmt.Sprintf("%s\n\nDate: %s\nOverall Operational Status: %s\n", defaultEmailBody, date, status),
		Attachment:  files.PDFPath,
		GeneratedAt: time.Now().UTC().Format(time.RFC3339),
	}
	if _, err := email.BuildMessage(request); err != nil {
		return err
	}
	printPlan(config, files, date, dryRun)
	if dryRun {
		fmt.Println("dry_run_result=validated, not sent")
		return nil
	}
	if err := email.Send(email.SendRequest{Config: config, Message: request}); err != nil {
		return err
	}
	fmt.Println("send_result=sent")
	return nil
}

func collectValue(args *[]string, name string) func(string) error {
	return func(value string) error {
		*args = append(*args, "--"+name, value)
		return nil
	}
}

func appRoot() string {
	workingDir, err := os.Getwd()
	if err == nil && hasReports(workingDir) {
		return workingDir
	}
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
	if workingDir != "" {
		return workingDir
	}
	return "."
}

func hasReports(root string) bool {
	_, err := os.Stat(filepath.Join(root, "reports", "daily"))
	return err == nil
}

func printPlan(config email.Config, files report.Files, date string, dryRun bool) {
	mode := "send"
	if dryRun {
		mode = "dry-run"
	}
	fmt.Printf("mode=%s\n", mode)
	fmt.Printf("date=%s\n", date)
	fmt.Printf("smtp_provider=%s\n", config.Provider)
	fmt.Printf("smtp_host=%s\n", config.Host)
	fmt.Printf("smtp_port=%d\n", config.Port)
	fmt.Printf("email_from=%s\n", config.From)
	fmt.Printf("email_to_count=%d\n", len(config.To))
	fmt.Printf("attachment=%s\n", files.PDFPath)
	fmt.Printf("attachment_size=%d\n", files.PDFSize)
}
