package mailer

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/sesv2"
	"github.com/aws/aws-sdk-go-v2/service/sesv2/types"
	"github.com/aws/smithy-go"

	domainservice "github.com/integrione/erp-retail/backend/services/iam/domain/service"
)

// sesV2API mengabstraksikan Client resmi AWS SDK v2 SESv2 agar unit test dapat memverifikasi payload tanpa jaringan AWS.
type sesV2API interface {
	SendEmail(ctx context.Context, params *sesv2.SendEmailInput, optFns ...func(*sesv2.Options)) (*sesv2.SendEmailOutput, error)
}

// SESSender mengirim email transaksional melalui Amazon SES API v2 resmi.
type SESSender struct {
	client           sesV2API
	fromAddress      string
	replyTo          string
	configurationSet string
}

// NewSESSender memuat konfigurasi kredensial AWS (static key opsional atau default IAM/IRSA chain) dan membentuk client SESv2.
func NewSESSender(ctx context.Context, rawCfg Config) (*SESSender, error) {
	cfg := rawCfg.Normalize()
	if cfg.SESRegion == "" {
		return nil, errors.New("aws region is required for ses provider")
	}

	opts := []func(*awsconfig.LoadOptions) error{
		awsconfig.WithRegion(cfg.SESRegion),
	}
	if cfg.SESAccessKeyID != "" && cfg.SESSecretAccessKey != "" {
		opts = append(opts, awsconfig.WithCredentialsProvider(
			credentials.NewStaticCredentialsProvider(cfg.SESAccessKeyID, cfg.SESSecretAccessKey, cfg.SESSessionToken),
		))
	}

	awsCfg, err := awsconfig.LoadDefaultConfig(ctx, opts...)
	if err != nil {
		return nil, fmt.Errorf("failed to load AWS config for SES: %w", err)
	}

	var sesOpts []func(*sesv2.Options)
	if cfg.SESEndpointURL != "" {
		endpoint := cfg.SESEndpointURL
		sesOpts = append(sesOpts, func(o *sesv2.Options) {
			o.BaseEndpoint = aws.String(endpoint)
		})
	}

	client := sesv2.NewFromConfig(awsCfg, sesOpts...)
	return &SESSender{
		client:           client,
		fromAddress:      cfg.From,
		replyTo:          cfg.ReplyTo,
		configurationSet: cfg.SESConfigurationSet,
	}, nil
}

// newSESSenderWithClient membentuk SESSender dengan client khusus untuk keperluan pengujian terkontrol.
func newSESSenderWithClient(client sesV2API, rawCfg Config) (*SESSender, error) {
	if client == nil {
		return nil, errors.New("ses client is required")
	}
	cfg := rawCfg.Normalize()
	return &SESSender{
		client:           client,
		fromAddress:      cfg.From,
		replyTo:          cfg.ReplyTo,
		configurationSet: cfg.SESConfigurationSet,
	}, nil
}

// ProviderName mengembalikan pengenal provider Amazon SES v2.
func (s *SESSender) ProviderName() string {
	return string(ProviderSES)
}

// Send memanggil operasi sesv2 SendEmail berformat UTF-8 untuk pesan teks maupun HTML.
func (s *SESSender) Send(ctx context.Context, msg domainservice.EmailMessage) (*domainservice.EmailSendResult, error) {
	if s == nil || s.client == nil {
		return nil, errors.New("ses sender is not initialized")
	}
	toAddr, err := extractEnvelopeAddress(msg.To)
	if err != nil {
		return nil, &PermanentError{Err: err}
	}
	if err := validateOutboundMessage(msg); err != nil {
		return nil, &PermanentError{Err: err}
	}

	body := &types.Body{}
	if strings.TrimSpace(msg.TextBody) != "" {
		body.Text = &types.Content{
			Data:    aws.String(msg.TextBody),
			Charset: aws.String("UTF-8"),
		}
	}
	if strings.TrimSpace(msg.HTMLBody) != "" {
		body.Html = &types.Content{
			Data:    aws.String(msg.HTMLBody),
			Charset: aws.String("UTF-8"),
		}
	}

	input := &sesv2.SendEmailInput{
		FromEmailAddress: aws.String(s.fromAddress),
		Destination: &types.Destination{
			ToAddresses: []string{toAddr},
		},
		Content: &types.EmailContent{
			Simple: &types.Message{
				Subject: &types.Content{
					Data:    aws.String(msg.Subject),
					Charset: aws.String("UTF-8"),
				},
				Body: body,
			},
		},
	}

	if s.replyTo != "" {
		replyAddr, err := extractEnvelopeAddress(s.replyTo)
		if err == nil {
			input.ReplyToAddresses = []string{replyAddr}
		}
	}
	if s.configurationSet != "" {
		input.ConfigurationSetName = aws.String(s.configurationSet)
	}
	if msg.Purpose != "" {
		input.EmailTags = []types.MessageTag{
			{
				Name:  aws.String("email_purpose"),
				Value: aws.String(sanitizeSESTagValue(string(msg.Purpose))),
			},
		}
	}

	out, err := s.client.SendEmail(ctx, input)
	if err != nil {
		return nil, classifySESError(err)
	}

	msgID := ""
	if out != nil {
		msgID = aws.ToString(out.MessageId)
	}
	return &domainservice.EmailSendResult{
		Provider:  string(ProviderSES),
		MessageID: msgID,
	}, nil
}

func sanitizeSESTagValue(v string) string {
	var b strings.Builder
	for _, r := range v {
		if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r == '_' || r == '-' {
			b.WriteRune(r)
		} else {
			b.WriteByte('_')
		}
	}
	res := b.String()
	if res == "" {
		return "iam_transactional"
	}
	if len(res) > 256 {
		return res[:256]
	}
	return res
}

func classifySESError(err error) error {
	if apiErr, ok := errors.AsType[smithy.APIError](err); ok {
		code := apiErr.ErrorCode()
		switch code {
		case "MessageRejected", "BadRequestException", "MailFromDomainNotVerifiedException", "InvalidParameterValue":
			return &PermanentError{Err: fmt.Errorf("ses rejected request (%s): %w", code, err)}
		}
	}
	return fmt.Errorf("ses send failed: %w", err)
}
