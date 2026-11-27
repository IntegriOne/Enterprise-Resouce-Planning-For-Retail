package mailer

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/mail"
	"regexp"
	"strings"

	domainservice "github.com/integrione/erp-retail/backend/services/iam/domain/service"
)

var providerEmailRedactor = regexp.MustCompile(`[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}`)

const (
	defaultSendGridBaseURL = "https://api.sendgrid.com"
	defaultResendBaseURL   = "https://api.resend.com"
	defaultPostmarkBaseURL = "https://api.postmarkapp.com"
)

type httpProviderSpec struct {
	provider      Provider
	endpointPath  string
	authHeader    string
	authPrefix    string
	messageIDHdr  string
	buildPayload  func(sender *HTTPProviderSender, toAddr string, msg domainservice.EmailMessage) (any, error)
	extractRespID func(body []byte, hdr http.Header) string
}

var httpProviderSpecs = map[Provider]httpProviderSpec{
	ProviderSendGrid: {
		provider:     ProviderSendGrid,
		endpointPath: "/v3/mail/send",
		authHeader:   "Authorization",
		authPrefix:   "Bearer ",
		messageIDHdr: "X-Message-Id",
		buildPayload: buildSendGridPayload,
		extractRespID: func(_ []byte, hdr http.Header) string {
			return strings.TrimSpace(hdr.Get("X-Message-Id"))
		},
	},
	ProviderResend: {
		provider:     ProviderResend,
		endpointPath: "/emails",
		authHeader:   "Authorization",
		authPrefix:   "Bearer ",
		buildPayload: buildResendPayload,
		extractRespID: func(body []byte, _ http.Header) string {
			var parsed struct {
				ID string `json:"id"`
			}
			_ = json.Unmarshal(body, &parsed)
			return strings.TrimSpace(parsed.ID)
		},
	},
	ProviderPostmark: {
		provider:     ProviderPostmark,
		endpointPath: "/email",
		authHeader:   "X-Postmark-Server-Token",
		authPrefix:   "",
		buildPayload: buildPostmarkPayload,
		extractRespID: func(body []byte, _ http.Header) string {
			var parsed struct {
				MessageID string `json:"MessageID"`
			}
			_ = json.Unmarshal(body, &parsed)
			return strings.TrimSpace(parsed.MessageID)
		},
	},
}

// HTTPProviderSender melayani pengiriman email melalui REST JSON API pihak ketiga (Twilio SendGrid, Resend, Postmark) dengan satu alur seragam.
type HTTPProviderSender struct {
	spec       httpProviderSpec
	baseURL    string
	apiKey     string
	fromRaw    string
	fromAddr   mail.Address
	replyToRaw string
	httpClient *http.Client
}

// NewHTTPProviderSender membentuk pengirim HTTP terparametrisasi untuk SendGrid, Resend, atau Postmark.
func NewHTTPProviderSender(rawCfg Config) (*HTTPProviderSender, error) {
	cfg := rawCfg.Normalize()
	spec, ok := httpProviderSpecs[cfg.Provider]
	if !ok {
		return nil, fmt.Errorf("unsupported HTTP email provider %q", cfg.Provider)
	}
	if strings.TrimSpace(cfg.APIKey) == "" {
		return nil, fmt.Errorf("API key is required for provider %s", cfg.Provider)
	}
	parsedFrom, err := mail.ParseAddress(cfg.From)
	if err != nil {
		return nil, fmt.Errorf("invalid from address: %w", err)
	}

	baseURL := cfg.APIBaseURL
	if baseURL == "" {
		switch cfg.Provider {
		case ProviderSendGrid:
			baseURL = defaultSendGridBaseURL
		case ProviderResend:
			baseURL = defaultResendBaseURL
		case ProviderPostmark:
			baseURL = defaultPostmarkBaseURL
		}
	}

	return &HTTPProviderSender{
		spec:       spec,
		baseURL:    strings.TrimRight(baseURL, "/"),
		apiKey:     strings.TrimSpace(cfg.APIKey),
		fromRaw:    cfg.From,
		fromAddr:   *parsedFrom,
		replyToRaw: cfg.ReplyTo,
		httpClient: &http.Client{Timeout: cfg.Timeout},
	}, nil
}

// ProviderName mengembalikan nama provider aktif.
func (s *HTTPProviderSender) ProviderName() string {
	if s == nil {
		return ""
	}
	return string(s.spec.provider)
}

// Send mengirim email ke endpoint REST API provider dan membedakan error sementara (5xx/429) vs permanen (4xx).
func (s *HTTPProviderSender) Send(ctx context.Context, msg domainservice.EmailMessage) (*domainservice.EmailSendResult, error) {
	if s == nil || s.httpClient == nil || s.spec.buildPayload == nil {
		return nil, errors.New("http email sender is not initialized")
	}
	toAddr, err := extractEnvelopeAddress(msg.To)
	if err != nil {
		return nil, &PermanentError{Err: err}
	}
	if err := validateOutboundMessage(msg); err != nil {
		return nil, &PermanentError{Err: err}
	}

	payloadObj, err := s.spec.buildPayload(s, toAddr, msg)
	if err != nil {
		return nil, &PermanentError{Err: err}
	}
	rawBody, err := json.Marshal(payloadObj)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal provider request: %w", err)
	}

	endpoint := s.baseURL + s.spec.endpointPath
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(rawBody))
	if err != nil {
		return nil, &PermanentError{Err: err}
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	req.Header.Set(s.spec.authHeader, s.spec.authPrefix+s.apiKey)

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("%s request failed: %w", s.spec.provider, err)
	}
	if resp == nil {
		return nil, fmt.Errorf("%s request returned empty response", s.spec.provider)
	}
	var respBytes []byte
	if resp.Body != nil {
		defer func() { _ = resp.Body.Close() }()
		respBytes, _ = io.ReadAll(io.LimitReader(resp.Body, 4096))
	}
	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		msgID := s.spec.extractRespID(respBytes, resp.Header)
		return &domainservice.EmailSendResult{
			Provider:  string(s.spec.provider),
			MessageID: msgID,
		}, nil
	}

	// Batasi dan sanitasi ringkasan respons error agar tidak membocorkan PII penerima atau memicu log injection.
	errMsg := fmt.Errorf("%s api returned status %d: %s", s.spec.provider, resp.StatusCode, sanitizeProviderErrorSummary(respBytes))
	// 429 dan 5xx layak dicoba ulang; status 4xx lain bersifat permanen.
	if resp.StatusCode >= 400 && resp.StatusCode < 500 && resp.StatusCode != http.StatusTooManyRequests && resp.StatusCode != http.StatusRequestTimeout {
		return nil, &PermanentError{Err: errMsg}
	}
	return nil, errMsg
}

func buildSendGridPayload(s *HTTPProviderSender, toAddr string, msg domainservice.EmailMessage) (any, error) {
	type sgAddr struct {
		Email string `json:"email"`
		Name  string `json:"name,omitempty"`
	}
	type sgContent struct {
		Type  string `json:"type"`
		Value string `json:"value"`
	}
	type sgPersonalization struct {
		To []sgAddr `json:"to"`
	}
	type sgRequest struct {
		Personalizations []sgPersonalization `json:"personalizations"`
		From             sgAddr              `json:"from"`
		ReplyTo          *sgAddr             `json:"reply_to,omitempty"`
		Subject          string              `json:"subject"`
		Content          []sgContent         `json:"content"`
		Categories       []string            `json:"categories,omitempty"`
	}

	contents := make([]sgContent, 0, 2)
	if strings.TrimSpace(msg.TextBody) != "" {
		contents = append(contents, sgContent{Type: "text/plain", Value: msg.TextBody})
	}
	if strings.TrimSpace(msg.HTMLBody) != "" {
		contents = append(contents, sgContent{Type: "text/html", Value: msg.HTMLBody})
	}
	if len(contents) == 0 {
		return nil, errors.New("email content is empty")
	}

	req := sgRequest{
		Personalizations: []sgPersonalization{{To: []sgAddr{{Email: toAddr}}}},
		From:             sgAddr{Email: s.fromAddr.Address, Name: s.fromAddr.Name},
		Subject:          msg.Subject,
		Content:          contents,
	}
	if s.replyToRaw != "" {
		if parsedReply, err := mail.ParseAddress(s.replyToRaw); err == nil {
			req.ReplyTo = &sgAddr{Email: parsedReply.Address, Name: parsedReply.Name}
		}
	}
	if msg.Purpose != "" {
		req.Categories = []string{string(msg.Purpose)}
	}
	return req, nil
}

func buildResendPayload(s *HTTPProviderSender, toAddr string, msg domainservice.EmailMessage) (any, error) {
	type resendRequest struct {
		From    string   `json:"from"`
		To      []string `json:"to"`
		ReplyTo string   `json:"reply_to,omitempty"`
		Subject string   `json:"subject"`
		Text    string   `json:"text,omitempty"`
		HTML    string   `json:"html,omitempty"`
	}

	return resendRequest{
		From:    s.fromRaw,
		To:      []string{toAddr},
		ReplyTo: s.replyToRaw,
		Subject: msg.Subject,
		Text:    msg.TextBody,
		HTML:    msg.HTMLBody,
	}, nil
}

func buildPostmarkPayload(s *HTTPProviderSender, toAddr string, msg domainservice.EmailMessage) (any, error) {
	type postmarkRequest struct {
		From          string `json:"From"`
		To            string `json:"To"`
		ReplyTo       string `json:"ReplyTo,omitempty"`
		Subject       string `json:"Subject"`
		TextBody      string `json:"TextBody,omitempty"`
		HTMLBody      string `json:"HtmlBody,omitempty"`
		Tag           string `json:"Tag,omitempty"`
		MessageStream string `json:"MessageStream,omitempty"`
	}

	return postmarkRequest{
		From:          s.fromRaw,
		To:            toAddr,
		ReplyTo:       s.replyToRaw,
		Subject:       msg.Subject,
		TextBody:      msg.TextBody,
		HTMLBody:      msg.HTMLBody,
		Tag:           string(msg.Purpose),
		MessageStream: "outbound",
	}, nil
}

func sanitizeProviderErrorSummary(raw []byte) string {
	const maxSummaryLen = 256
	cleaned := strings.Map(func(r rune) rune {
		if r == '\r' || r == '\n' || r == '\t' || (r >= 0 && r < 0x20) || r == 0x7f {
			return ' '
		}
		return r
	}, strings.TrimSpace(string(raw)))
	cleaned = providerEmailRedactor.ReplaceAllString(cleaned, "[REDACTED_EMAIL]")
	if len(cleaned) > maxSummaryLen {
		cleaned = cleaned[:maxSummaryLen] + "..."
	}
	return cleaned
}
