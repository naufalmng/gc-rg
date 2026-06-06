package email

import (
	"strings"
	"testing"
)

func TestParseConfig_AppliesGmailDefaults(t *testing.T) {
	env := map[string]string{
		"GC_RG_EMAIL_PROVIDER":      "gmail",
		"GC_RG_EMAIL_FROM":          "sender@gmail.com",
		"GC_RG_EMAIL_TO":            "ops@example.com, manager@example.com",
		"GC_RG_SMTP_USERNAME":       "sender@gmail.com",
		"GC_RG_SMTP_PASSWORD":       "app-password",
		"GC_RG_EMAIL_SUBJECT_PREFIX": "[GC-RG]",
	}

	config, err := ParseConfig(nil, env)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if config.Host != "smtp.gmail.com" {
		t.Fatalf("Host = %q, want smtp.gmail.com", config.Host)
	}
	if config.Port != 587 {
		t.Fatalf("Port = %d, want 587", config.Port)
	}
	if config.TLSMode != TLSModeStartTLS {
		t.Fatalf("TLSMode = %q, want starttls", config.TLSMode)
	}
	if len(config.To) != 2 {
		t.Fatalf("To count = %d, want 2", len(config.To))
	}
}

func TestParseConfig_RejectsMissingPasswordWhenAuthEnabled(t *testing.T) {
	env := map[string]string{
		"GC_RG_EMAIL_PROVIDER": "gmail",
		"GC_RG_EMAIL_FROM":     "sender@gmail.com",
		"GC_RG_EMAIL_TO":       "ops@example.com",
		"GC_RG_SMTP_USERNAME":  "sender@gmail.com",
	}

	_, err := ParseConfig(nil, env)
	if err == nil {
		t.Fatal("expected missing password error")
	}
	if !strings.Contains(err.Error(), "GC_RG_SMTP_PASSWORD") {
		t.Fatalf("error = %q, want password context", err.Error())
	}
}

func TestParseConfig_AllowsCustomProvider(t *testing.T) {
	env := map[string]string{
		"GC_RG_EMAIL_PROVIDER": "custom",
		"GC_RG_SMTP_HOST":      "mail.example.com",
		"GC_RG_SMTP_PORT":      "465",
		"GC_RG_SMTP_TLS":       "ssl",
		"GC_RG_EMAIL_FROM":     "sender@example.com",
		"GC_RG_EMAIL_TO":       "ops@example.com",
		"GC_RG_SMTP_USERNAME":  "sender@example.com",
		"GC_RG_SMTP_PASSWORD":  "secret",
	}

	config, err := ParseConfig(nil, env)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if config.Host != "mail.example.com" {
		t.Fatalf("Host = %q, want mail.example.com", config.Host)
	}
	if config.Port != 465 {
		t.Fatalf("Port = %d, want 465", config.Port)
	}
	if config.TLSMode != TLSModeSSL {
		t.Fatalf("TLSMode = %q, want ssl", config.TLSMode)
	}
}
