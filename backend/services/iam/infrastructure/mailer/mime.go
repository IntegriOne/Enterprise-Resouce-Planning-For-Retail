package mailer

import (
	"bytes"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"mime"
	"mime/multipart"
	"net/mail"
	"net/textproto"
	"strings"
	"time"

	domainservice "github.com/integrione/erp-retail/backend/services/iam/domain/service"
)

func validateHeaderSafe(field, val string) error {
	if strings.ContainsAny(val, "\r\n\x00") {
		return fmt.Errorf("invalid %s header value: forbidden control characters", field)
	}
	return nil
}

func validateOutboundMessage(msg domainservice.EmailMessage) error {
	to := strings.TrimSpace(msg.To)
	if to == "" {
		return errors.New("recipient email address is required")
	}
	if err := validateHeaderSafe("to", to); err != nil {
		return err
	}
	if _, err := mail.ParseAddress(to); err != nil {
		return fmt.Errorf("invalid recipient email address: %w", err)
	}

	subj := strings.TrimSpace(msg.Subject)
	if subj == "" {
		return errors.New("email subject is required")
	}
	if err := validateHeaderSafe("subject", subj); err != nil {
		return err
	}
	if strings.TrimSpace(msg.TextBody) == "" && strings.TrimSpace(msg.HTMLBody) == "" {
		return errors.New("email body is required")
	}
	return nil
}

func extractEnvelopeAddress(raw string) (string, error) {
	if err := validateHeaderSafe("envelope", raw); err != nil {
		return "", err
	}
	addr, err := mail.ParseAddress(strings.TrimSpace(raw))
	if err != nil {
		return "", fmt.Errorf("invalid envelope address: %w", err)
	}
	if err := validateHeaderSafe("envelope_address", addr.Address); err != nil {
		return "", err
	}
	return addr.Address, nil
}

// buildMIMEMessage menyusun payload MIME multipart/alternative berstandar RFC 5322 yang aman terhadap header injection.
func buildMIMEMessage(fromHeader, replyToHeader string, msg domainservice.EmailMessage, now time.Time) ([]byte, string, error) {
	if err := validateOutboundMessage(msg); err != nil {
		return nil, "", &PermanentError{Err: err}
	}
	if err := validateHeaderSafe("from", fromHeader); err != nil {
		return nil, "", &PermanentError{Err: err}
	}
	fromAddr, err := mail.ParseAddress(fromHeader)
	if err != nil {
		return nil, "", &PermanentError{Err: fmt.Errorf("invalid from address: %w", err)}
	}

	toAddr, err := mail.ParseAddress(msg.To)
	if err != nil {
		return nil, "", &PermanentError{Err: fmt.Errorf("invalid recipient address: %w", err)}
	}

	var replyToFormatted string
	if strings.TrimSpace(replyToHeader) != "" {
		if err := validateHeaderSafe("reply_to", replyToHeader); err != nil {
			return nil, "", &PermanentError{Err: err}
		}
		parsedReply, err := mail.ParseAddress(replyToHeader)
		if err != nil {
			return nil, "", &PermanentError{Err: fmt.Errorf("invalid reply-to address: %w", err)}
		}
		replyToFormatted = parsedReply.String()
	}

	if now.IsZero() {
		now = time.Now().UTC()
	}
	domainPart := "integrione.local"
	if at := strings.LastIndexByte(fromAddr.Address, '@'); at > 0 && at < len(fromAddr.Address)-1 {
		domainPart = fromAddr.Address[at+1:]
	}
	msgID := fmt.Sprintf("<%d.%s@%s>", now.UnixNano(), randomHexSuffix(), domainPart)

	var bodyBuf bytes.Buffer
	writer := multipart.NewWriter(&bodyBuf)
	boundary := writer.Boundary()

	if strings.TrimSpace(msg.TextBody) != "" {
		textHdr := textproto.MIMEHeader{}
		textHdr.Set("Content-Type", "text/plain; charset=UTF-8")
		textHdr.Set("Content-Transfer-Encoding", "8bit")
		part, err := writer.CreatePart(textHdr)
		if err != nil {
			return nil, "", err
		}
		if _, err := part.Write([]byte(normalizeLineEndings(msg.TextBody))); err != nil {
			return nil, "", err
		}
	}

	if strings.TrimSpace(msg.HTMLBody) != "" {
		htmlHdr := textproto.MIMEHeader{}
		htmlHdr.Set("Content-Type", "text/html; charset=UTF-8")
		htmlHdr.Set("Content-Transfer-Encoding", "8bit")
		part, err := writer.CreatePart(htmlHdr)
		if err != nil {
			return nil, "", err
		}
		if _, err := part.Write([]byte(normalizeLineEndings(msg.HTMLBody))); err != nil {
			return nil, "", err
		}
	}

	if err := writer.Close(); err != nil {
		return nil, "", err
	}

	var out bytes.Buffer
	out.WriteString("From: " + fromAddr.String() + "\r\n")
	out.WriteString("To: " + toAddr.Address + "\r\n")
	if replyToFormatted != "" {
		out.WriteString("Reply-To: " + replyToFormatted + "\r\n")
	}
	out.WriteString("Subject: " + mime.QEncoding.Encode("UTF-8", msg.Subject) + "\r\n")
	out.WriteString("Date: " + now.UTC().Format(time.RFC1123Z) + "\r\n")
	out.WriteString("Message-ID: " + msgID + "\r\n")
	out.WriteString("MIME-Version: 1.0\r\n")
	fmt.Fprintf(&out, "Content-Type: multipart/alternative; boundary=%q\r\n", boundary)
	out.WriteString("\r\n")
	out.Write(bodyBuf.Bytes())

	return out.Bytes(), msgID, nil
}

func normalizeLineEndings(s string) string {
	s = strings.ReplaceAll(s, "\r\n", "\n")
	s = strings.ReplaceAll(s, "\r", "\n")
	return strings.ReplaceAll(s, "\n", "\r\n")
}

func randomHexSuffix() string {
	var b [8]byte
	if _, err := rand.Read(b[:]); err != nil {
		return "00000000"
	}
	return hex.EncodeToString(b[:])
}
