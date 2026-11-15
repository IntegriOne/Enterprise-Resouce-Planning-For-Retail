package user

import (
	"context"
	"log/slog"
	"time"

	domainservice "github.com/integrione/erp-retail/backend/services/iam/domain/service"
	"github.com/integrione/erp-retail/backend/services/iam/infrastructure/mailer"
)

func (s *Service) dispatchPasswordResetEmail(
	ctx context.Context,
	recipientName string,
	recipientEmail string,
	issued *IssuedTokenResult,
) {
	if s == nil || s.emailSender == nil || issued == nil {
		return
	}
	msg, err := mailer.BuildPasswordResetMessage(
		s.cfg.PublicBaseURL,
		recipientName,
		recipientEmail,
		issued.PlaintextToken,
		s.cfg.PasswordResetTTL,
		s.clock.Now(),
	)
	s.sendTransactionalMessage(ctx, msg, err)
}

func (s *Service) dispatchVerificationEmail(
	ctx context.Context,
	recipientName string,
	recipientEmail string,
	rawToken string,
	ttl time.Duration,
) {
	if s == nil || s.emailSender == nil || rawToken == "" {
		return
	}
	msg, err := mailer.BuildEmailVerificationMessage(
		s.cfg.PublicBaseURL,
		recipientName,
		recipientEmail,
		rawToken,
		ttl,
		s.clock.Now(),
	)
	s.sendTransactionalMessage(ctx, msg, err)
}

func (s *Service) sendTransactionalMessage(ctx context.Context, msg domainservice.EmailMessage, buildErr error) {
	if buildErr != nil {
		s.logEmailFailure(msg.Purpose, msg.To, buildErr)
		return
	}
	if _, err := s.emailSender.Send(ctx, msg); err != nil {
		// Keamanan anti-enumerasi: kegagalan transport email dicatat terstruktur tanpa membocorkan status akun ke pemanggil publik.
		s.logEmailFailure(msg.Purpose, msg.To, err)
	}
}

func (s *Service) logEmailFailure(purpose domainservice.EmailPurpose, to string, err error) {
	if s.logger == nil || err == nil {
		return
	}
	s.logger.Error("failed to dispatch IAM transactional email",
		slog.String("purpose", string(purpose)),
		slog.String("recipient", domainservice.MaskEmailForLog(to)),
		slog.String("error", err.Error()),
	)
}
