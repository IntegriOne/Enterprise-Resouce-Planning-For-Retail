package mailer

import (
	"bufio"
	"bytes"
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"math/big"
	"net"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/sesv2"
	"github.com/aws/smithy-go"

	domainservice "github.com/integrione/erp-retail/backend/services/iam/domain/service"
)

type fakeSESClient struct {
	mu     sync.Mutex
	inputs []*sesv2.SendEmailInput
	fn     func(context.Context, *sesv2.SendEmailInput) (*sesv2.SendEmailOutput, error)
}

func (f *fakeSESClient) SendEmail(ctx context.Context, params *sesv2.SendEmailInput, _ ...func(*sesv2.Options)) (*sesv2.SendEmailOutput, error) {
	f.mu.Lock()
	f.inputs = append(f.inputs, params)
	f.mu.Unlock()
	if f.fn != nil {
		return f.fn(ctx, params)
	}
	return &sesv2.SendEmailOutput{MessageId: aws.String("ses-msg-001")}, nil
}

type fakeAPIError struct {
	code string
	msg  string
}

func (e fakeAPIError) Error() string                 { return e.code + ": " + e.msg }
func (e fakeAPIError) ErrorCode() string             { return e.code }
func (e fakeAPIError) ErrorMessage() string          { return e.msg }
func (e fakeAPIError) ErrorFault() smithy.ErrorFault { return smithy.FaultClient }

func TestConfigValidationAndSecurityPolicies(t *testing.T) {
	t.Parallel()

	if err := DefaultConfig().Validate(false); err != nil {
		t.Fatalf("DefaultConfig().Validate(false) error = %v", err)
	}

	// SES tanpa region harus ditolak fail-closed.
	sesNoRegion := Config{Provider: ProviderSES, From: "no-reply@integrione.id"}
	if err := sesNoRegion.Validate(true); err == nil {
		t.Fatal("expected SES config without region to fail validation")
	}

	// Kredensial statis AWS tidak boleh setengah terisi.
	sesHalfCreds := Config{Provider: ProviderSES, From: "no-reply@integrione.id", SESRegion: "ap-southeast-1", SESAccessKeyID: "AKIA123"}
	if err := sesHalfCreds.Validate(false); err == nil {
		t.Fatal("expected partial AWS credentials to fail validation")
	}

	// Plaintext SMTP dilarang pada production maupun jika AllowPlaintext tidak diaktifkan.
	plainSMTP := Config{
		Provider:       ProviderSMTP,
		From:           "no-reply@integrione.id",
		SMTPHost:       "mailpit",
		SMTPPort:       1025,
		SMTPTLSMode:    SMTPTLSNone,
		AllowPlaintext: true,
	}
	if err := plainSMTP.Validate(true); err == nil {
		t.Fatal("expected plaintext SMTP to be rejected in production")
	}
	if err := plainSMTP.Validate(false); err != nil {
		t.Fatalf("expected plaintext SMTP allowed in dev with opt-in, got: %v", err)
	}

	// SendGrid wajib memiliki API Key.
	if err := (Config{Provider: ProviderSendGrid, From: "no-reply@integrione.id"}).Validate(false); err == nil {
		t.Fatal("expected SendGrid without API key to fail validation")
	}
}

func TestTemplatesEscapeHTMLAndMaskLogs(t *testing.T) {
	t.Parallel()

	msg, err := BuildPasswordResetMessage(
		"https://erp.integrione.id",
		`<script>alert(1)</script>`,
		"user@integrione.id",
		"tok+secret/123=",
		15*time.Minute,
		time.Date(2026, 11, 15, 6, 0, 0, 0, time.UTC),
	)
	if err != nil {
		t.Fatalf("BuildPasswordResetMessage() error = %v", err)
	}
	if !strings.Contains(msg.HTMLBody, "&lt;script&gt;alert(1)&lt;/script&gt;") {
		t.Fatalf("expected HTML escaping for recipient name, got HTML: %s", msg.HTMLBody)
	}
	if !strings.Contains(msg.HTMLBody, "https://erp.integrione.id/login#reset_token=tok%2Bsecret%2F123%3D") {
		t.Fatalf("expected fragment URL in reset email HTML, got: %s", msg.HTMLBody)
	}
	if strings.Contains(msg.HTMLBody, "?reset_token=") {
		t.Fatalf("reset URL must use # fragment instead of ? query string")
	}

	verMsg, err := BuildEmailVerificationMessage(
		"https://erp.integrione.id/",
		"Budi",
		"budi@integrione.id",
		"verify-abc",
		24*time.Hour,
		time.Now(),
	)
	if err != nil {
		t.Fatalf("BuildEmailVerificationMessage() error = %v", err)
	}
	if !strings.Contains(verMsg.TextBody, "https://erp.integrione.id/verify-email#token=verify-abc") {
		t.Fatalf("expected fragment verify URL, got: %s", verMsg.TextBody)
	}

	if got := domainservice.MaskEmailForLog("budi.santoso@integrione.id"); got != "b***@integrione.id" {
		t.Fatalf("unexpected masked email: %q", got)
	}
}

func TestSESSenderPayloadAndPermanentErrorClassification(t *testing.T) {
	t.Parallel()

	client := &fakeSESClient{}
	sender, err := newSESSenderWithClient(client, Config{
		Provider:            ProviderSES,
		From:                "IntegriOne <security@integrione.id>",
		ReplyTo:             "support@integrione.id",
		SESRegion:           "ap-southeast-1",
		SESConfigurationSet: "iam-transactional",
	})
	if err != nil {
		t.Fatalf("newSESSenderWithClient() error = %v", err)
	}

	msg, _ := BuildPasswordResetMessage("https://erp.integrione.id", "Siti", "siti@integrione.id", "raw-tok-1", 15*time.Minute, time.Now())
	res, err := sender.Send(context.Background(), msg)
	if err != nil {
		t.Fatalf("Send() error = %v", err)
	}
	if res.Provider != "ses" || res.MessageID != "ses-msg-001" {
		t.Fatalf("unexpected SES result: %+v", res)
	}
	if len(client.inputs) != 1 {
		t.Fatalf("expected 1 SES SendEmail call, got %d", len(client.inputs))
	}
	in := client.inputs[0]
	if aws.ToString(in.FromEmailAddress) != "IntegriOne <security@integrione.id>" || aws.ToString(in.ConfigurationSetName) != "iam-transactional" {
		t.Fatalf("unexpected SES metadata: from=%q cfgSet=%q", aws.ToString(in.FromEmailAddress), aws.ToString(in.ConfigurationSetName))
	}
	if len(in.ReplyToAddresses) != 1 || in.ReplyToAddresses[0] != "support@integrione.id" {
		t.Fatalf("unexpected SES ReplyTo: %v", in.ReplyToAddresses)
	}

	// Tolak CRLF injection sebelum memanggil SDK.
	badMsg := msg
	badMsg.Subject = "Hello\r\nBcc: evil@attacker.tld"
	if _, err := sender.Send(context.Background(), badMsg); !IsPermanent(err) {
		t.Fatalf("expected PermanentError on header injection, got: %v", err)
	}

	// Error 400 SES MessageRejected harus diklasifikasikan permanen agar tidak retry sia-sia.
	client.fn = func(context.Context, *sesv2.SendEmailInput) (*sesv2.SendEmailOutput, error) {
		return nil, fakeAPIError{code: "MessageRejected", msg: "Email address is not verified"}
	}
	if _, err := sender.Send(context.Background(), msg); !IsPermanent(err) {
		t.Fatalf("expected PermanentError for SES MessageRejected, got: %v", err)
	}
}

func TestHTTPProvidersSendGridResendPostmark(t *testing.T) {
	t.Parallel()

	cases := []struct {
		provider   Provider
		expectPath string
		authHdr    string
		authVal    string
		respHdr    string
		respBody   string
		wantMsgID  string
	}{
		{
			provider:   ProviderSendGrid,
			expectPath: "/v3/mail/send",
			authHdr:    "Authorization",
			authVal:    "Bearer sg-key-123",
			respHdr:    "sg-msg-xyz",
			respBody:   ``,
			wantMsgID:  "sg-msg-xyz",
		},
		{
			provider:   ProviderResend,
			expectPath: "/emails",
			authHdr:    "Authorization",
			authVal:    "Bearer re-key-123",
			respBody:   `{"id":"re_001"}`,
			wantMsgID:  "re_001",
		},
		{
			provider:   ProviderPostmark,
			expectPath: "/email",
			authHdr:    "X-Postmark-Server-Token",
			authVal:    "pm-tok-123",
			respBody:   `{"MessageID":"pm-uuid-01"}`,
			wantMsgID:  "pm-uuid-01",
		},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(string(tc.provider), func(t *testing.T) {
			t.Parallel()
			var gotBody map[string]any
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.URL.Path != tc.expectPath {
					t.Fatalf("path = %q, want %q", r.URL.Path, tc.expectPath)
				}
				if gotAuth := r.Header.Get(tc.authHdr); gotAuth != tc.authVal {
					t.Fatalf("header %s = %q, want %q", tc.authHdr, gotAuth, tc.authVal)
				}
				raw, _ := io.ReadAll(r.Body)
				_ = json.Unmarshal(raw, &gotBody)
				if tc.respHdr != "" {
					w.Header().Set("X-Message-Id", tc.respHdr)
				}
				w.WriteHeader(http.StatusAccepted)
				_, _ = w.Write([]byte(tc.respBody))
			}))
			defer srv.Close()

			sender, err := NewHTTPProviderSender(Config{
				Provider:   tc.provider,
				From:       "IntegriOne <noreply@integrione.id>",
				APIKey:     strings.TrimPrefix(tc.authVal, "Bearer "),
				APIBaseURL: srv.URL,
				Timeout:    2 * time.Second,
			})
			if err != nil {
				t.Fatalf("NewHTTPProviderSender() error = %v", err)
			}

			msg, _ := BuildPasswordResetMessage("https://erp.integrione.id", "Rina", "rina@integrione.id", "tok-http-1", 15*time.Minute, time.Now())
			res, err := sender.Send(context.Background(), msg)
			if err != nil {
				t.Fatalf("Send() error = %v", err)
			}
			if res.MessageID != tc.wantMsgID || len(gotBody) == 0 {
				t.Fatalf("unexpected result=%+v body=%v", res, gotBody)
			}
		})
	}
}

func TestSMTPSenderImplicitTLSWireFormat(t *testing.T) {
	t.Parallel()

	tlsCert, err := generateLocalhostSelfSignedCert()
	if err != nil {
		t.Fatalf("generateLocalhostSelfSignedCert() error = %v", err)
	}
	ln, err := tls.Listen("tcp", "127.0.0.1:0", &tls.Config{Certificates: []tls.Certificate{tlsCert}})
	if err != nil {
		t.Fatalf("tls.Listen() error = %v", err)
	}
	defer func() { _ = ln.Close() }()

	var receivedPayload strings.Builder
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		conn, acceptErr := ln.Accept()
		if acceptErr != nil {
			return
		}
		defer func() { _ = conn.Close() }()
		r := bufio.NewReader(conn)
		_, _ = io.WriteString(conn, "220 localhost ESMTP test\r\n")
		for {
			line, readErr := r.ReadString('\n')
			if readErr != nil {
				return
			}
			upper := strings.ToUpper(strings.TrimSpace(line))
			switch {
			case strings.HasPrefix(upper, "EHLO"), strings.HasPrefix(upper, "HELO"):
				_, _ = io.WriteString(conn, "250-localhost\r\n250 OK\r\n")
			case strings.HasPrefix(upper, "MAIL FROM:"):
				_, _ = io.WriteString(conn, "250 OK\r\n")
			case strings.HasPrefix(upper, "RCPT TO:"):
				_, _ = io.WriteString(conn, "250 OK\r\n")
			case upper == "DATA":
				_, _ = io.WriteString(conn, "354 End data with <CR><LF>.<CR><LF>\r\n")
				for {
					dataLine, errData := r.ReadString('\n')
					if errData != nil {
						return
					}
					if dataLine == ".\r\n" || dataLine == ".\n" {
						break
					}
					receivedPayload.WriteString(dataLine)
				}
				_, _ = io.WriteString(conn, "250 Queued\r\n")
			case upper == "QUIT":
				_, _ = io.WriteString(conn, "221 Bye\r\n")
				return
			default:
				_, _ = io.WriteString(conn, "250 OK\r\n")
			}
		}
	}()

	host, portStr, _ := net.SplitHostPort(ln.Addr().String())
	port, _ := strconv.Atoi(portStr)
	pool := x509.NewCertPool()
	pool.AddCert(tlsCert.Leaf)

	tlsSender, err := NewSMTPSender(Config{
		Provider:       ProviderSMTP,
		From:           "IntegriOne <noreply@integrione.local>",
		SMTPHost:       host,
		SMTPPort:       port,
		SMTPServerName: "localhost",
		SMTPTLSMode:    SMTPTLSImplicit,
		Timeout:        3 * time.Second,
	})
	if err != nil {
		t.Fatalf("NewSMTPSender(tls) error = %v", err)
	}
	tlsSender.tlsRootCAs = pool

	tlsMsg, _ := BuildPasswordResetMessage("https://erp.localhost:8443", "Adi", "adi@integrione.local", "smtp-tls-reset-tok", 15*time.Minute, time.Now())
	if _, err := tlsSender.Send(context.Background(), tlsMsg); err != nil {
		t.Fatalf("SMTPSender(tls).Send() error = %v", err)
	}
	_ = ln.Close()
	wg.Wait()

	wire := receivedPayload.String()
	if !strings.Contains(wire, "Content-Type: multipart/alternative;") || !strings.Contains(wire, "smtp-tls-reset-tok") {
		t.Fatalf("unexpected MIME wire payload:\n%s", wire)
	}
}

func TestDispatcherRetriesTransientAndDrainsOnClose(t *testing.T) {
	t.Parallel()

	var attempts atomic.Int32
	var logBuf bytes.Buffer
	logger := slog.New(slog.NewJSONHandler(&logBuf, nil))

	flaky := &funcEmailSender{
		name: "ses",
		fn: func(_ context.Context, _ domainservice.EmailMessage) (*domainservice.EmailSendResult, error) {
			n := attempts.Add(1)
			if n < 3 {
				return nil, errors.New("temporary network glitch")
			}
			return &domainservice.EmailSendResult{Provider: "ses", MessageID: "msg-retried-3"}, nil
		},
	}

	disp := NewDispatcher(flaky, DispatcherConfig{
		Mode:                DeliveryModeAsync,
		Workers:             1,
		QueueSize:           8,
		SendTimeout:         time.Second,
		MaxAttempts:         3,
		InitialRetryBackoff: 10 * time.Millisecond,
	}, logger)
	disp.Start(context.Background())

	msg, _ := BuildPasswordResetMessage("https://erp.integrione.id", "Dian", "dian@integrione.id", "SUPER_SECRET_RAW_TOKEN_999", 15*time.Minute, time.Now())
	res, err := disp.Send(context.Background(), msg)
	if err != nil || res.MessageID != "queued" {
		t.Fatalf("async Send() = (%+v, %v)", res, err)
	}

	drainCtx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	if err := disp.Close(drainCtx); err != nil {
		t.Fatalf("disp.Close() error = %v", err)
	}
	if attempts.Load() != 3 {
		t.Fatalf("attempts = %d, want 3", attempts.Load())
	}

	logged := logBuf.String()
	if strings.Contains(logged, "SUPER_SECRET_RAW_TOKEN_999") || strings.Contains(logged, "dian@integrione.id") {
		t.Fatalf("logs leaked raw token or unmasked email: %s", logged)
	}
	if !strings.Contains(logged, "d***@integrione.id") {
		t.Fatalf("expected masked email in logs, got: %s", logged)
	}
}

type funcEmailSender struct {
	name string
	fn   func(context.Context, domainservice.EmailMessage) (*domainservice.EmailSendResult, error)
}

func (f *funcEmailSender) ProviderName() string { return f.name }
func (f *funcEmailSender) Send(ctx context.Context, msg domainservice.EmailMessage) (*domainservice.EmailSendResult, error) {
	return f.fn(ctx, msg)
}

func generateLocalhostSelfSignedCert() (tls.Certificate, error) {
	priv, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return tls.Certificate{}, err
	}
	tmpl := x509.Certificate{
		SerialNumber:          big.NewInt(1),
		Subject:               pkix.Name{CommonName: "localhost"},
		NotBefore:             time.Now().Add(-time.Hour),
		NotAfter:              time.Now().Add(24 * time.Hour),
		KeyUsage:              x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		BasicConstraintsValid: true,
		IsCA:                  true,
		DNSNames:              []string{"localhost"},
		IPAddresses:           []net.IP{net.ParseIP("127.0.0.1")},
	}
	der, err := x509.CreateCertificate(rand.Reader, &tmpl, &tmpl, &priv.PublicKey, priv)
	if err != nil {
		return tls.Certificate{}, err
	}
	leaf, err := x509.ParseCertificate(der)
	if err != nil {
		return tls.Certificate{}, err
	}
	return tls.Certificate{
		Certificate: [][]byte{der},
		PrivateKey:  priv,
		Leaf:        leaf,
	}, nil
}

func TestHTTPProviderClassifies4xxPermanentVs5xxTransient(t *testing.T) {
	t.Parallel()

	var statusCode atomic.Int32
	statusCode.Store(http.StatusBadRequest)
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(int(statusCode.Load()))
		_, _ = fmt.Fprint(w, `{"error":"invalid request"}`)
	}))
	defer srv.Close()

	sender, err := NewHTTPProviderSender(Config{
		Provider:   ProviderResend,
		From:       "IntegriOne <noreply@integrione.id>",
		APIKey:     "re-test",
		APIBaseURL: srv.URL,
	})
	if err != nil {
		t.Fatalf("NewHTTPProviderSender() error = %v", err)
	}

	msg := domainservice.EmailMessage{
		Purpose:  domainservice.EmailPurposePasswordReset,
		To:       "test@integrione.id",
		Subject:  "Reset",
		TextBody: "Body",
	}
	if _, err := sender.Send(context.Background(), msg); !IsPermanent(err) {
		t.Fatalf("expected HTTP 400 to be PermanentError, got: %v", err)
	}

	statusCode.Store(http.StatusTooManyRequests)
	if _, err := sender.Send(context.Background(), msg); err == nil || IsPermanent(err) {
		t.Fatalf("expected HTTP 429 to be retryable error, got: %v", err)
	}
}
