package mailer

import (
	"bytes"
	"context"
	"fmt"
	htmltmpl "html/template"
	"net/url"
	"strings"
	texttmpl "text/template"
	"time"

	domainservice "github.com/integrione/erp-retail/backend/services/iam/domain/service"
)

type emailTemplateData struct {
	RecipientName string
	ActionURL     string
	RawToken      string
	ExpiresIn     string
	Year          int
}

var (
	passwordResetHTMLTmpl = htmltmpl.Must(htmltmpl.New("reset_html").Parse(`<!DOCTYPE html>
<html lang="id">
<head><meta charset="UTF-8"><title>Pemulihan Kata Sandi IntegriOne</title></head>
<body style="margin:0;padding:24px;background:#f4f7f8;font-family:-apple-system,BlinkMacSystemFont,'Segoe UI',Roboto,sans-serif;color:#17202a;">
  <div style="max-width:560px;margin:0 auto;background:#ffffff;border:1px solid #dde3e8;border-radius:16px;overflow:hidden;">
    <div style="background:#0b2439;padding:24px 28px;color:#f9fafb;">
      <div style="font-size:12px;letter-spacing:0.12em;text-transform:uppercase;color:#da414a;font-weight:700;">INTEGRIONE IDENTITY</div>
      <h1 style="margin:8px 0 0;font-size:20px;font-weight:600;">Permintaan Atur Ulang Kata Sandi</h1>
    </div>
    <div style="padding:28px;font-size:15px;line-height:1.6;">
      <p style="margin-top:0;">Halo <strong>{{.RecipientName}}</strong>,</p>
      <p>Kami menerima permintaan untuk mengatur ulang kata sandi akun IntegriOne Anda. Klik tombol berikut untuk membuat kata sandi baru:</p>
      <p style="margin:24px 0;">
        <a href="{{.ActionURL}}" style="display:inline-block;padding:12px 22px;background:#133554;color:#ffffff;text-decoration:none;border-radius:10px;font-weight:600;">Atur Ulang Kata Sandi</a>
      </p>
      <p style="font-size:13px;color:#66717c;">Tautan ini berlaku selama <strong>{{.ExpiresIn}}</strong> dan hanya dapat dipakai satu kali. Jika tombol di atas tidak berfungsi, salin kode token berikut ke halaman pemulihan:</p>
      <pre style="padding:12px;background:#f4f7f8;border:1px solid #dde3e8;border-radius:8px;font-size:13px;word-break:break-all;white-space:pre-wrap;">{{.RawToken}}</pre>
      <p style="font-size:13px;color:#66717c;margin-bottom:0;">Jika Anda tidak meminta pengaturan ulang kata sandi, abaikan email ini. Jangan pernah membagikan tautan atau token pemulihan kepada siapa pun.</p>
    </div>
    <div style="padding:16px 28px;background:#f9fafb;border-top:1px solid #dde3e8;font-size:12px;color:#66717c;">
      &copy; {{.Year}} IntegriOne Retail ERP &amp; Commerce Security
    </div>
  </div>
</body>
</html>`))

	passwordResetTextTmpl = texttmpl.Must(texttmpl.New("reset_text").Parse(`Halo {{.RecipientName}},

Kami menerima permintaan untuk mengatur ulang kata sandi akun IntegriOne Anda.

Buka tautan berikut untuk membuat kata sandi baru (berlaku {{.ExpiresIn}}, hanya 1 kali pakai):
{{.ActionURL}}

Jika diperlukan, Anda juga dapat memasukkan token pemulihan berikut secara manual pada halaman /login:
{{.RawToken}}

Jika Anda tidak meminta perubahan ini, abaikan email ini dan jangan bagikan token ini kepada siapa pun.`))

	verifyEmailHTMLTmpl = htmltmpl.Must(htmltmpl.New("verify_html").Parse(`<!DOCTYPE html>
<html lang="id">
<head><meta charset="UTF-8"><title>Verifikasi Alamat Email IntegriOne</title></head>
<body style="margin:0;padding:24px;background:#f4f7f8;font-family:-apple-system,BlinkMacSystemFont,'Segoe UI',Roboto,sans-serif;color:#17202a;">
  <div style="max-width:560px;margin:0 auto;background:#ffffff;border:1px solid #dde3e8;border-radius:16px;overflow:hidden;">
    <div style="background:#0b2439;padding:24px 28px;color:#f9fafb;">
      <div style="font-size:12px;letter-spacing:0.12em;text-transform:uppercase;color:#da414a;font-weight:700;">INTEGRIONE IDENTITY</div>
      <h1 style="margin:8px 0 0;font-size:20px;font-weight:600;">Verifikasi Alamat Email Anda</h1>
    </div>
    <div style="padding:28px;font-size:15px;line-height:1.6;">
      <p style="margin-top:0;">Halo <strong>{{.RecipientName}}</strong>,</p>
      <p>Silakan konfirmasi kepemilikan alamat email ini untuk mengaktifkan akun IntegriOne Anda:</p>
      <p style="margin:24px 0;">
        <a href="{{.ActionURL}}" style="display:inline-block;padding:12px 22px;background:#133554;color:#ffffff;text-decoration:none;border-radius:10px;font-weight:600;">Verifikasi Alamat Email</a>
      </p>
      <p style="font-size:13px;color:#66717c;margin-bottom:0;">Tautan verifikasi ini berlaku selama <strong>{{.ExpiresIn}}</strong> dan hanya dapat dipakai satu kali. Jika Anda tidak merasa mendaftar atau mengubah email, abaikan pesan ini.</p>
    </div>
    <div style="padding:16px 28px;background:#f9fafb;border-top:1px solid #dde3e8;font-size:12px;color:#66717c;">
      &copy; {{.Year}} IntegriOne Retail ERP &amp; Commerce Security
    </div>
  </div>
</body>
</html>`))

	verifyEmailTextTmpl = texttmpl.Must(texttmpl.New("verify_text").Parse(`Halo {{.RecipientName}},

Silakan konfirmasi kepemilikan alamat email Anda dengan membuka tautan berikut (berlaku {{.ExpiresIn}}, hanya 1 kali pakai):
{{.ActionURL}}

Jika Anda tidak merasa melakukan permintaan ini, abaikan email ini.`))
)

// BuildPasswordResetMessage menyusun email pemulihan kata sandi ber-template HTML dan teks dengan tautan URL fragment yang aman dari kebocoran referer.
func BuildPasswordResetMessage(
	publicBaseURL string,
	recipientName string,
	recipientEmail string,
	rawToken string,
	ttl time.Duration,
	now time.Time,
) (domainservice.EmailMessage, error) {
	base := strings.TrimRight(strings.TrimSpace(publicBaseURL), "/")
	if base == "" {
		base = DefaultConfig().PublicBaseURL
	}
	if err := validatePublicBaseURL(base); err != nil {
		return domainservice.EmailMessage{}, err
	}
	cleanToken := strings.TrimSpace(rawToken)
	if cleanToken == "" {
		return domainservice.EmailMessage{}, fmt.Errorf("reset token is required")
	}
	// Fragment URL (#reset_token=...) mencegah token tertulis ke log access proxy maupun header HTTP Referer.
	actionURL := fmt.Sprintf("%s/login#reset_token=%s", base, url.QueryEscape(cleanToken))

	data := emailTemplateData{
		RecipientName: fallbackRecipientName(recipientName),
		ActionURL:     actionURL,
		RawToken:      cleanToken,
		ExpiresIn:     formatDurationID(ttl),
		Year:          resolveYear(now),
	}

	var htmlBuf, textBuf bytes.Buffer
	if err := passwordResetHTMLTmpl.Execute(&htmlBuf, data); err != nil {
		return domainservice.EmailMessage{}, fmt.Errorf("failed to render password reset html: %w", err)
	}
	if err := passwordResetTextTmpl.Execute(&textBuf, data); err != nil {
		return domainservice.EmailMessage{}, fmt.Errorf("failed to render password reset text: %w", err)
	}

	return domainservice.EmailMessage{
		Purpose:   domainservice.EmailPurposePasswordReset,
		To:        strings.TrimSpace(recipientEmail),
		Subject:   "Atur Ulang Kata Sandi Akun IntegriOne Anda",
		TextBody:  textBuf.String(),
		HTMLBody:  htmlBuf.String(),
		CreatedAt: now,
	}, nil
}

// BuildEmailVerificationMessage menyusun email verifikasi kepemilikan alamat email menggunakan tautan fragment #token=.
func BuildEmailVerificationMessage(
	publicBaseURL string,
	recipientName string,
	recipientEmail string,
	rawToken string,
	ttl time.Duration,
	now time.Time,
) (domainservice.EmailMessage, error) {
	base := strings.TrimRight(strings.TrimSpace(publicBaseURL), "/")
	if base == "" {
		base = DefaultConfig().PublicBaseURL
	}
	if err := validatePublicBaseURL(base); err != nil {
		return domainservice.EmailMessage{}, err
	}
	cleanToken := strings.TrimSpace(rawToken)
	if cleanToken == "" {
		return domainservice.EmailMessage{}, fmt.Errorf("verification token is required")
	}
	actionURL := fmt.Sprintf("%s/verify-email#token=%s", base, url.QueryEscape(cleanToken))

	data := emailTemplateData{
		RecipientName: fallbackRecipientName(recipientName),
		ActionURL:     actionURL,
		RawToken:      cleanToken,
		ExpiresIn:     formatDurationID(ttl),
		Year:          resolveYear(now),
	}

	var htmlBuf, textBuf bytes.Buffer
	if err := verifyEmailHTMLTmpl.Execute(&htmlBuf, data); err != nil {
		return domainservice.EmailMessage{}, fmt.Errorf("failed to render verification html: %w", err)
	}
	if err := verifyEmailTextTmpl.Execute(&textBuf, data); err != nil {
		return domainservice.EmailMessage{}, fmt.Errorf("failed to render verification text: %w", err)
	}

	return domainservice.EmailMessage{
		Purpose:   domainservice.EmailPurposeEmailVerification,
		To:        strings.TrimSpace(recipientEmail),
		Subject:   "Verifikasi Alamat Email Akun IntegriOne Anda",
		TextBody:  textBuf.String(),
		HTMLBody:  htmlBuf.String(),
		CreatedAt: now,
	}, nil
}

func fallbackRecipientName(name string) string {
	trimmed := strings.Map(func(r rune) rune {
		if r == '\r' || r == '\n' || r == '\t' || (r >= 0 && r < 0x20) || r == 0x7f {
			return -1
		}
		return r
	}, strings.TrimSpace(name))
	if trimmed == "" {
		return "Pengguna"
	}
	return trimmed
}

func resolveYear(now time.Time) int {
	if now.IsZero() {
		return time.Now().Year()
	}
	return now.Year()
}

func formatDurationID(d time.Duration) string {
	if d <= 0 {
		return "beberapa menit"
	}
	if d >= time.Hour {
		hours := int(d.Hours())
		if hours == 1 {
			return "1 jam"
		}
		return fmt.Sprintf("%d jam", hours)
	}
	mins := max(1, int(d.Minutes()))
	return fmt.Sprintf("%d menit", mins)
}

// TransactionalNotifier membungkus EmailSender dan URL dasar publik untuk layanan domain pengguna.
type TransactionalNotifier struct {
	Sender        domainservice.EmailSender
	PublicBaseURL string
}

// NotifyPasswordReset mengirim pesan reset kata sandi berdasarkan URL publik terkonfigurasi.
func (n *TransactionalNotifier) NotifyPasswordReset(ctx context.Context, recipientName, email, rawToken string, ttl time.Duration, now time.Time) (*domainservice.EmailSendResult, error) {
	msg, err := BuildPasswordResetMessage(n.PublicBaseURL, recipientName, email, rawToken, ttl, now)
	if err != nil {
		return nil, err
	}
	return n.Sender.Send(ctx, msg)
}
