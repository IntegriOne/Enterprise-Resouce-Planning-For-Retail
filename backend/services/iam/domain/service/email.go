package service

import (
	"context"
	"time"

	"github.com/integrione/erp-retail/backend/services/iam/domain/entity"
)

// EmailPurpose mengklasifikasikan jenis pesan email transaksional yang dikirim oleh layanan IAM.
type EmailPurpose string

const (
	// EmailPurposePasswordReset dipakai untuk pengiriman tautan dan token pemulihan kata sandi.
	EmailPurposePasswordReset EmailPurpose = "password_reset"
	// EmailPurposeEmailVerification dipakai untuk verifikasi alamat email baru atau perubahan email tertunda.
	EmailPurposeEmailVerification EmailPurpose = "email_verification"
)

// EmailMessage merepresentasikan muatan email transaksional terstruktur untuk dikirim melalui provider email.
type EmailMessage struct {
	Purpose   EmailPurpose
	To        string
	Subject   string
	TextBody  string
	HTMLBody  string
	Metadata  map[string]string
	CreatedAt time.Time
}

// EmailSendResult memuat metadata hasil pengiriman yang dikembalikan provider email.
type EmailSendResult struct {
	Provider  string
	MessageID string
}

// EmailSender mengabstraksikan pengiriman email transaksional IAM ke Amazon SES, SMTP, Twilio SendGrid, Resend, Postmark, atau mode log.
type EmailSender interface {
	ProviderName() string
	Send(ctx context.Context, msg EmailMessage) (*EmailSendResult, error)
}

// MaskEmailForLog menyamarkan local-part alamat email sebelum dicatat ke log agar PII tetap terlindungi.
func MaskEmailForLog(email string) string {
	return entity.MaskEmail(email)
}
