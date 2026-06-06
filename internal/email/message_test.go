package email

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestBuildMessage_AttachesPDFAndMarkdownBody(t *testing.T) {
	dir := t.TempDir()
	pdfPath := filepath.Join(dir, "report.pdf")
	if err := os.WriteFile(pdfPath, []byte("%PDF test"), 0o600); err != nil {
		t.Fatalf("write pdf: %v", err)
	}
	request := MessageRequest{
		From:        "sender@example.com",
		To:          []string{"ops@example.com"},
		Subject:     "Daily Report",
		Body:        "Report attached.",
		Attachment:  pdfPath,
		GeneratedAt: "2026-06-06T00:00:00Z",
	}

	message, err := BuildMessage(request)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	text := string(message)
	assertContains(t, text, "From: sender@example.com")
	assertContains(t, text, "To: ops@example.com")
	assertContains(t, text, "Subject: Daily Report")
	assertContains(t, text, "Content-Type: application/pdf")
	assertContains(t, text, "filename=\"report.pdf\"")
	assertContains(t, text, "JVBERiB0ZXN0")
}

func assertContains(t *testing.T, text, want string) {
	t.Helper()
	if !strings.Contains(text, want) {
		t.Fatalf("message missing %q\n%s", want, text)
	}
}
