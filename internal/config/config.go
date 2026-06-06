package config

import (
	"errors"
	"flag"
	"fmt"
	"os"
	"regexp"
	"strings"
	"time"
)

const (
	DefaultReportDir = "reports/daily"
	DefaultSessionDB = ".wa-go/whatsmeow.db"
)

var (
	datePattern       = regexp.MustCompile(`^\d{4}-\d{2}-\d{2}$`)
	groupJIDPattern   = regexp.MustCompile(`^\d+@g\.us$`)
	numericJIDPattern = regexp.MustCompile(`^\d+$`)
)

type Options struct {
	Date      string
	ReportDir string
	SessionDB string
	JID       string
	Caption   string
	Send      bool
	FreshAuth bool
	LoginOnly bool
}

func ParseArgs(args []string, env map[string]string) (Options, error) {
	options := Options{
		Date:      time.Now().Format(time.DateOnly),
		ReportDir: DefaultReportDir,
		SessionDB: DefaultSessionDB,
		JID:       strings.TrimSpace(env["REPORT_WHATSAPP_GROUP_JID"]),
	}

	flags := flag.NewFlagSet("send-whatsapp-report", flag.ContinueOnError)
	flags.StringVar(&options.Date, "date", options.Date, "report date in YYYY-MM-DD format")
	flags.StringVar(&options.ReportDir, "report-dir", options.ReportDir, "daily report directory")
	flags.StringVar(&options.SessionDB, "session-db", options.SessionDB, "WhatsApp session SQLite database")
	flags.StringVar(&options.JID, "jid", options.JID, "WhatsApp group JID")
	flags.StringVar(&options.Caption, "caption", "", "override document caption")
	flags.BoolVar(&options.Send, "send", false, "send the WhatsApp document")
	flags.BoolVar(&options.FreshAuth, "fresh-auth", false, "remove existing session DB before connecting")
	flags.BoolVar(&options.LoginOnly, "login-only", false, "login and save session without sending")

	if err := flags.Parse(args); err != nil {
		return Options{}, fmt.Errorf("parse args: %w", err)
	}
	if positional := flags.Args(); len(positional) > 0 {
		if len(positional) > 1 || !datePattern.MatchString(positional[0]) {
			return Options{}, fmt.Errorf("unknown positional argument: %s", strings.Join(positional, " "))
		}
		options.Date = positional[0]
	}
	if numericJIDPattern.MatchString(options.JID) {
		options.JID = options.JID + "@g.us"
	}
	if err := Validate(options); err != nil {
		return Options{}, err
	}
	return options, nil
}

func Validate(options Options) error {
	if !datePattern.MatchString(options.Date) {
		return errors.New("--date must use YYYY-MM-DD format")
	}
	if options.JID == "" {
		return errors.New("REPORT_WHATSAPP_GROUP_JID or --jid is required")
	}
	if !groupJIDPattern.MatchString(options.JID) {
		return errors.New("WhatsApp group JID must look like 120363000000000000@g.us")
	}
	if strings.TrimSpace(options.ReportDir) == "" {
		return errors.New("--report-dir is required")
	}
	if strings.TrimSpace(options.SessionDB) == "" {
		return errors.New("--session-db is required")
	}
	return nil
}

func EnvMap() map[string]string {
	return map[string]string{
		"REPORT_WHATSAPP_GROUP_JID": os.Getenv("REPORT_WHATSAPP_GROUP_JID"),
	}
}
