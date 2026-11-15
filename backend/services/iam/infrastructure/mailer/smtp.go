package mailer

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"errors"
	"fmt"
	"net"
	"net/smtp"
	"strconv"
	"strings"
	"time"

	domainservice "github.com/integrione/erp-retail/backend/services/iam/domain/service"
)

// SMTPSender mengirim email melalui protokol SMTP dengan dukungan TLS implisit (465), STARTTLS (587), maupun relay lokal dev.
type SMTPSender struct {
	addr           string
	host           string
	serverName     string
	username       string
	password       string
	fromHeader     string
	fromEnvelope   string
	replyToHeader  string
	tlsMode        SMTPTLSMode
	allowPlaintext bool
	timeout        time.Duration
	tlsRootCAs     *x509.CertPool
}

// NewSMTPSender memvalidasi konfigurasi SMTP dan membentuk instance SMTPSender.
func NewSMTPSender(cfg Config) (*SMTPSender, error) {
	norm := cfg.Normalize()
	if norm.SMTPHost == "" {
		return nil, errors.New("smtp host is required")
	}
	fromEnv, err := extractEnvelopeAddress(norm.From)
	if err != nil {
		return nil, fmt.Errorf("invalid SMTP from address: %w", err)
	}
	serverName := norm.SMTPServerName
	if serverName == "" {
		serverName = norm.SMTPHost
	}

	return &SMTPSender{
		addr:           net.JoinHostPort(norm.SMTPHost, strconv.Itoa(norm.SMTPPort)),
		host:           norm.SMTPHost,
		serverName:     serverName,
		username:       norm.SMTPUsername,
		password:       norm.SMTPPassword,
		fromHeader:     norm.From,
		fromEnvelope:   fromEnv,
		replyToHeader:  norm.ReplyTo,
		tlsMode:        norm.SMTPTLSMode,
		allowPlaintext: norm.AllowPlaintext,
		timeout:        norm.Timeout,
	}, nil
}

// ProviderName mengembalikan pengenal provider SMTP.
func (s *SMTPSender) ProviderName() string {
	return string(ProviderSMTP)
}

// Send membangun muatan MIME yang aman dan mengirimkannya via SMTP dengan penghormatan deadline konteks.
func (s *SMTPSender) Send(ctx context.Context, msg domainservice.EmailMessage) (*domainservice.EmailSendResult, error) {
	toEnv, err := extractEnvelopeAddress(msg.To)
	if err != nil {
		return nil, &PermanentError{Err: err}
	}

	rawMIME, msgID, err := buildMIMEMessage(s.fromHeader, s.replyToHeader, msg, msg.CreatedAt)
	if err != nil {
		return nil, err
	}

	sendCtx := ctx
	if sendCtx == nil {
		sendCtx = context.Background()
	}
	if _, hasDeadline := sendCtx.Deadline(); !hasDeadline && s.timeout > 0 {
		var cancel context.CancelFunc
		sendCtx, cancel = context.WithTimeout(sendCtx, s.timeout)
		defer cancel()
	}

	client, err := s.dialClient(sendCtx)
	if err != nil {
		return nil, err
	}
	defer func() { _ = client.Close() }()

	if s.username != "" {
		if ok, _ := client.Extension("AUTH"); !ok {
			return nil, errors.New("smtp server does not advertise AUTH extension")
		}
		auth := smtp.PlainAuth("", s.username, s.password, s.host)
		if err := client.Auth(auth); err != nil {
			return nil, &PermanentError{Err: fmt.Errorf("smtp authentication failed: %w", err)}
		}
	}

	if err := client.Mail(s.fromEnvelope); err != nil {
		return nil, classifySMTPError("MAIL FROM", err)
	}
	if err := client.Rcpt(toEnv); err != nil {
		return nil, classifySMTPError("RCPT TO", err)
	}

	wc, err := client.Data()
	if err != nil {
		return nil, classifySMTPError("DATA", err)
	}
	if _, err := wc.Write(rawMIME); err != nil {
		_ = wc.Close()
		return nil, fmt.Errorf("smtp write failed: %w", err)
	}
	if err := wc.Close(); err != nil {
		return nil, classifySMTPError("DATA CLOSE", err)
	}
	if err := client.Quit(); err != nil {
		// Beberapa relay memutus TCP segera setelah '.' tanpa merespons QUIT secara eksplisit.
		_ = client.Close()
	}

	return &domainservice.EmailSendResult{
		Provider:  string(ProviderSMTP),
		MessageID: msgID,
	}, nil
}

func (s *SMTPSender) dialClient(ctx context.Context) (*smtp.Client, error) {
	tlsCfg := &tls.Config{
		ServerName: s.serverName,
		MinVersion: tls.VersionTLS12,
		RootCAs:    s.tlsRootCAs,
	}

	var conn net.Conn
	var err error
	dialer := &net.Dialer{Timeout: s.timeout}

	if s.tlsMode == SMTPTLSImplicit {
		tlsDialer := &tls.Dialer{NetDialer: dialer, Config: tlsCfg}
		conn, err = tlsDialer.DialContext(ctx, "tcp", s.addr)
	} else {
		conn, err = dialer.DialContext(ctx, "tcp", s.addr)
	}
	if err != nil {
		return nil, fmt.Errorf("smtp dial failed: %w", err)
	}
	if deadline, ok := ctx.Deadline(); ok {
		_ = conn.SetDeadline(deadline)
	}

	client, err := smtp.NewClient(conn, s.host)
	if err != nil {
		_ = conn.Close()
		return nil, fmt.Errorf("smtp client handshake failed: %w", err)
	}

	if s.tlsMode == SMTPTLSStartTLS {
		if ok, _ := client.Extension("STARTTLS"); !ok {
			_ = client.Close()
			return nil, errors.New("smtp server does not support required STARTTLS")
		}
		if err := client.StartTLS(tlsCfg); err != nil {
			_ = client.Close()
			return nil, fmt.Errorf("smtp STARTTLS negotiation failed: %w", err)
		}
	}

	return client, nil
}

func classifySMTPError(stage string, err error) error {
	if err == nil {
		return nil
	}
	msg := err.Error()
	// Kode SMTP 5xx menandakan penolakan permanen; 4xx bersifat sementara.
	if strings.HasPrefix(strings.TrimSpace(msg), "5") {
		return &PermanentError{Err: fmt.Errorf("smtp %s rejected: %w", stage, err)}
	}
	return fmt.Errorf("smtp %s failed: %w", stage, err)
}
