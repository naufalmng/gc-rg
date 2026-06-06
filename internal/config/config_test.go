package config

import "testing"

func TestParseArgs_NormalizesNumericJID(t *testing.T) {
	options, err := ParseArgs([]string{"--date", "2026-06-05", "--jid", "120363026366173025"}, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if options.JID != "120363026366173025@g.us" {
		t.Fatalf("JID = %q, want normalized group JID", options.JID)
	}
}

func TestParseArgs_RejectsInvalidDate(t *testing.T) {
	_, err := ParseArgs([]string{"--date", "2026/06/05", "--jid", "120363026366173025@g.us"}, nil)
	if err == nil {
		t.Fatal("expected invalid date error")
	}
}

func TestParseArgs_ReadsJIDFromEnv(t *testing.T) {
	env := map[string]string{"REPORT_WHATSAPP_GROUP_JID": "120363026366173025@g.us"}
	options, err := ParseArgs([]string{"2026-06-05"}, env)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if options.Date != "2026-06-05" {
		t.Fatalf("Date = %q, want positional date", options.Date)
	}
	if options.JID != env["REPORT_WHATSAPP_GROUP_JID"] {
		t.Fatalf("JID = %q, want env JID", options.JID)
	}
}
