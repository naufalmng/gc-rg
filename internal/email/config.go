package email

import (
	"errors"
	"flag"
	"fmt"
	"os"
	"strconv"
	"strings"
)

type TLSMode string

const (
	TLSModeStartTLS TLSMode = "starttls"
	TLSModeSSL      TLSMode = "ssl"
	TLSModeNone     TLSMode = "none"
)

type Config struct {
	Provider      string
	Host          string
	Port          int
	TLSMode       TLSMode
	From          string
	To            []string
	CC            []string
	Username      string
	Password      string
	AuthEnabled   bool
	SubjectPrefix string
}

func ParseConfig(args []string, env map[string]string) (Config, error) {
	config := Config{
		Provider:      valueFromEnv(env, "GC_RG_EMAIL_PROVIDER", "custom"),
		Host:          valueFromEnv(env, "GC_RG_SMTP_HOST", ""),
		Port:          intFromEnv(env, "GC_RG_SMTP_PORT", 587),
		TLSMode:       TLSMode(valueFromEnv(env, "GC_RG_SMTP_TLS", string(TLSModeStartTLS))),
		From:          valueFromEnv(env, "GC_RG_EMAIL_FROM", ""),
		To:            splitRecipients(valueFromEnv(env, "GC_RG_EMAIL_TO", "")),
		CC:            splitRecipients(valueFromEnv(env, "GC_RG_EMAIL_CC", "")),
		Username:      valueFromEnv(env, "GC_RG_SMTP_USERNAME", ""),
		Password:      valueFromEnv(env, "GC_RG_SMTP_PASSWORD", ""),
		AuthEnabled:   valueFromEnv(env, "GC_RG_SMTP_AUTH", "on") != "off",
		SubjectPrefix: valueFromEnv(env, "GC_RG_EMAIL_SUBJECT_PREFIX", "[GC-RG]"),
	}
	flags := flag.NewFlagSet("email-config", flag.ContinueOnError)
	flags.StringVar(&config.Provider, "email-provider", config.Provider, "SMTP provider: gmail, yahoo, outlook, or custom")
	flags.StringVar(&config.Host, "smtp-host", config.Host, "SMTP host")
	flags.IntVar(&config.Port, "smtp-port", config.Port, "SMTP port")
	flags.StringVar((*string)(&config.TLSMode), "smtp-tls", string(config.TLSMode), "SMTP TLS: starttls, ssl, or none")
	flags.StringVar(&config.From, "email-from", config.From, "sender email address")
	toRaw := strings.Join(config.To, ",")
	ccRaw := strings.Join(config.CC, ",")
	flags.StringVar(&toRaw, "email-to", toRaw, "comma-separated email recipients")
	flags.StringVar(&ccRaw, "email-cc", ccRaw, "comma-separated email CC recipients")
	flags.StringVar(&config.Username, "smtp-username", config.Username, "SMTP username")
	flags.StringVar(&config.Password, "smtp-password", config.Password, "SMTP password or app password")
	flags.StringVar(&config.SubjectPrefix, "email-subject-prefix", config.SubjectPrefix, "email subject prefix")
	if err := flags.Parse(args); err != nil {
		return Config{}, fmt.Errorf("parse email config args: %w", err)
	}
	config.To = splitRecipients(toRaw)
	config.CC = splitRecipients(ccRaw)
	applyProviderDefaults(&config)
	if err := ValidateConfig(config); err != nil {
		return Config{}, err
	}
	return config, nil
}

func ValidateConfig(config Config) error {
	if !isKnownProvider(config.Provider) {
		return errors.New("GC_RG_EMAIL_PROVIDER must be gmail, yahoo, outlook, or custom")
	}
	if strings.TrimSpace(config.Host) == "" {
		return errors.New("GC_RG_SMTP_HOST is required")
	}
	if config.Port <= 0 || config.Port > 65535 {
		return errors.New("GC_RG_SMTP_PORT must be between 1 and 65535")
	}
	if config.TLSMode != TLSModeStartTLS && config.TLSMode != TLSModeSSL && config.TLSMode != TLSModeNone {
		return errors.New("GC_RG_SMTP_TLS must be starttls, ssl, or none")
	}
	if strings.TrimSpace(config.From) == "" {
		return errors.New("GC_RG_EMAIL_FROM is required")
	}
	if len(config.To) == 0 {
		return errors.New("GC_RG_EMAIL_TO is required")
	}
	if config.AuthEnabled && strings.TrimSpace(config.Username) == "" {
		return errors.New("GC_RG_SMTP_USERNAME is required")
	}
	if config.AuthEnabled && strings.TrimSpace(config.Password) == "" {
		return errors.New("GC_RG_SMTP_PASSWORD is required")
	}
	return nil
}

func EnvMap() map[string]string {
	keys := []string{
		"GC_RG_EMAIL_PROVIDER", "GC_RG_SMTP_HOST", "GC_RG_SMTP_PORT", "GC_RG_SMTP_TLS",
		"GC_RG_EMAIL_FROM", "GC_RG_EMAIL_TO", "GC_RG_EMAIL_CC", "GC_RG_SMTP_USERNAME",
		"GC_RG_SMTP_PASSWORD", "GC_RG_SMTP_AUTH", "GC_RG_EMAIL_SUBJECT_PREFIX",
	}
	values := make(map[string]string, len(keys))
	for _, key := range keys {
		values[key] = getenv(key)
	}
	return values
}

var getenv = os.Getenv

func applyProviderDefaults(config *Config) {
	switch strings.ToLower(config.Provider) {
	case "gmail":
		config.Host = defaultString(config.Host, "smtp.gmail.com")
		config.Port = defaultPort(config.Port, 587)
		config.TLSMode = defaultTLS(config.TLSMode, TLSModeStartTLS)
	case "yahoo":
		config.Host = defaultString(config.Host, "smtp.mail.yahoo.com")
		config.Port = defaultPort(config.Port, 587)
		config.TLSMode = defaultTLS(config.TLSMode, TLSModeStartTLS)
	case "outlook":
		config.Host = defaultString(config.Host, "smtp.office365.com")
		config.Port = defaultPort(config.Port, 587)
		config.TLSMode = defaultTLS(config.TLSMode, TLSModeStartTLS)
	}
}

func isKnownProvider(provider string) bool {
	switch strings.ToLower(provider) {
	case "gmail", "yahoo", "outlook", "custom":
		return true
	default:
		return false
	}
}

func valueFromEnv(env map[string]string, key, fallback string) string {
	if env == nil {
		return fallback
	}
	value := strings.TrimSpace(env[key])
	if value == "" {
		return fallback
	}
	return value
}

func intFromEnv(env map[string]string, key string, fallback int) int {
	value := valueFromEnv(env, key, "")
	if value == "" {
		return fallback
	}
	parsed, err := strconv.Atoi(value)
	if err != nil {
		return fallback
	}
	return parsed
}

func splitRecipients(raw string) []string {
	parts := strings.Split(raw, ",")
	recipients := make([]string, 0, len(parts))
	for _, part := range parts {
		trimmed := strings.TrimSpace(part)
		if trimmed != "" {
			recipients = append(recipients, trimmed)
		}
	}
	return recipients
}

func defaultString(value, fallback string) string {
	if strings.TrimSpace(value) == "" {
		return fallback
	}
	return value
}

func defaultPort(value, fallback int) int {
	if value == 0 {
		return fallback
	}
	return value
}

func defaultTLS(value, fallback TLSMode) TLSMode {
	if value == "" {
		return fallback
	}
	return value
}
