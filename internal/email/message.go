package email

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"mime"
	"net/textproto"
	"os"
	"path/filepath"
	"strings"
)

const mimeLineLength = 76

type MessageRequest struct {
	From        string
	To          []string
	CC          []string
	Subject     string
	Body        string
	Attachment  string
	GeneratedAt string
}

func BuildMessage(request MessageRequest) ([]byte, error) {
	attachmentBytes, err := os.ReadFile(request.Attachment)
	if err != nil {
		return nil, fmt.Errorf("read email attachment: %w", err)
	}
	boundary := "gc-rg-report-boundary"
	var buffer bytes.Buffer
	writeHeaders(&buffer, request, boundary)
	writeTextPart(&buffer, boundary, request.Body)
	writeAttachmentPart(&buffer, boundary, request.Attachment, attachmentBytes)
	fmt.Fprintf(&buffer, "--%s--\r\n", boundary)
	return buffer.Bytes(), nil
}

func writeHeaders(buffer *bytes.Buffer, request MessageRequest, boundary string) {
	headers := textproto.MIMEHeader{}
	headers.Set("From", request.From)
	headers.Set("To", strings.Join(request.To, ", "))
	if len(request.CC) > 0 {
		headers.Set("Cc", strings.Join(request.CC, ", "))
	}
	headers.Set("Subject", request.Subject)
	headers.Set("MIME-Version", "1.0")
	headers.Set("Content-Type", fmt.Sprintf("multipart/mixed; boundary=%q", boundary))
	if request.GeneratedAt != "" {
		headers.Set("X-GC-RG-Generated-At", request.GeneratedAt)
	}
	for key, values := range headers {
		for _, value := range values {
			fmt.Fprintf(buffer, "%s: %s\r\n", key, value)
		}
	}
	buffer.WriteString("\r\n")
}

func writeTextPart(buffer *bytes.Buffer, boundary, body string) {
	fmt.Fprintf(buffer, "--%s\r\n", boundary)
	buffer.WriteString("Content-Type: text/plain; charset=utf-8\r\n")
	buffer.WriteString("Content-Transfer-Encoding: 8bit\r\n\r\n")
	buffer.WriteString(body)
	buffer.WriteString("\r\n")
}

func writeAttachmentPart(buffer *bytes.Buffer, boundary, path string, content []byte) {
	fileName := filepath.Base(path)
	encodedName := mime.QEncoding.Encode("utf-8", fileName)
	fmt.Fprintf(buffer, "--%s\r\n", boundary)
	buffer.WriteString("Content-Type: application/pdf\r\n")
	buffer.WriteString("Content-Transfer-Encoding: base64\r\n")
	fmt.Fprintf(buffer, "Content-Disposition: attachment; filename=\"%s\"\r\n", encodedName)
	buffer.WriteString("\r\n")
	writeBase64Lines(buffer, content)
	buffer.WriteString("\r\n")
}

func writeBase64Lines(buffer *bytes.Buffer, content []byte) {
	encoded := base64.StdEncoding.EncodeToString(content)
	for len(encoded) > mimeLineLength {
		buffer.WriteString(encoded[:mimeLineLength])
		buffer.WriteString("\r\n")
		encoded = encoded[mimeLineLength:]
	}
	buffer.WriteString(encoded)
	buffer.WriteString("\r\n")
}
