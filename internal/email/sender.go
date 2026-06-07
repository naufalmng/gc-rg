package email

import (
	"crypto/tls"
	"fmt"
	"net"
	"net/smtp"
	"strings"
)

type SendRequest struct {
	Config  Config
	Message MessageRequest
}

func Send(request SendRequest) error {
	message, err := BuildMessage(request.Message)
	if err != nil {
		return err
	}
	address := net.JoinHostPort(request.Config.Host, fmt.Sprintf("%d", request.Config.Port))
	recipients := append([]string{}, request.Config.To...)
	recipients = append(recipients, request.Config.CC...)
	switch request.Config.TLSMode {
	case TLSModeSSL:
		return sendWithSSL(address, request.Config, recipients, message)
	case TLSModeStartTLS, TLSModeNone:
		return sendWithOptionalStartTLS(address, request.Config, recipients, message)
	default:
		return fmt.Errorf("unsupported SMTP TLS mode: %s", request.Config.TLSMode)
	}
}

func sendWithSSL(address string, config Config, recipients []string, message []byte) error {
	connection, err := tls.Dial("tcp", address, &tls.Config{ServerName: config.Host, MinVersion: tls.VersionTLS12})
	if err != nil {
		return fmt.Errorf("connect SMTP SSL: %w", err)
	}
	defer connection.Close()
	client, err := smtp.NewClient(connection, config.Host)
	if err != nil {
		return fmt.Errorf("create SMTP client: %w", err)
	}
	defer client.Close()
	if err := client.Hello(config.HeloName); err != nil {
		return fmt.Errorf("smtp hello: %w", err)
	}
	return sendWithClient(client, config, recipients, message)
}

func sendWithOptionalStartTLS(address string, config Config, recipients []string, message []byte) error {
	client, err := smtp.Dial(address)
	if err != nil {
		return fmt.Errorf("connect SMTP: %w", err)
	}
	defer client.Close()
	if err := client.Hello(config.HeloName); err != nil {
		return fmt.Errorf("smtp hello: %w", err)
	}
	if config.TLSMode == TLSModeStartTLS {
		if err := client.StartTLS(&tls.Config{ServerName: config.Host, MinVersion: tls.VersionTLS12}); err != nil {
			return fmt.Errorf("start SMTP TLS: %w", err)
		}
	}
	return sendWithClient(client, config, recipients, message)
}

func sendWithClient(client *smtp.Client, config Config, recipients []string, message []byte) error {
	if config.AuthEnabled {
		auth := smtp.PlainAuth("", config.Username, config.Password, config.Host)
		if err := client.Auth(auth); err != nil {
			return fmt.Errorf("smtp auth: %w", err)
		}
	}
	if err := client.Mail(config.From); err != nil {
		return fmt.Errorf("smtp from %s: %w", config.From, err)
	}
	for _, recipient := range recipients {
		trimmed := strings.TrimSpace(recipient)
		if trimmed == "" {
			continue
		}
		if err := client.Rcpt(trimmed); err != nil {
			return fmt.Errorf("smtp recipient %s: %w", trimmed, err)
		}
	}
	writer, err := client.Data()
	if err != nil {
		return fmt.Errorf("smtp data: %w", err)
	}
	if _, err := writer.Write(message); err != nil {
		return fmt.Errorf("write SMTP message: %w", err)
	}
	if err := writer.Close(); err != nil {
		return fmt.Errorf("close SMTP message: %w", err)
	}
	if err := client.Quit(); err != nil {
		return fmt.Errorf("smtp quit: %w", err)
	}
	return nil
}
