package brevo

import (
	"context"
	"errors"
	"fmt"
	"net/mail"
	"strings"


	brevo "github.com/getbrevo/brevo-go/lib"
	"stitchforce-server/config"
	"stitchforce-server/pkg/email"
)

type BrevoSender struct {
	from   string
	client *brevo.APIClient
}

func New(cfg config.BrevoConfig) (*BrevoSender, error) {
	if strings.TrimSpace(cfg.APIKey) == "" {
		return nil, errors.New("brevo API key is required")
	}

	if strings.TrimSpace(cfg.From) == "" {
		return nil, errors.New("email sender address is required")
	}

	brevoCfg := brevo.NewConfiguration()
	brevoCfg.AddDefaultHeader("api-key", cfg.APIKey)

	return &BrevoSender{
		from:   cfg.From,
		client: brevo.NewAPIClient(brevoCfg),
	}, nil
}

func parseTo(values []string) ([]brevo.SendSmtpEmailTo, error) {
	recipients := make([]brevo.SendSmtpEmailTo, 0, len(values))

	for _, value := range values {
		address, err := mail.ParseAddress(strings.TrimSpace(value))
		if err != nil {
			return nil, fmt.Errorf("invalid email address %q: %w", value, err)
		}

		recipients = append(recipients, brevo.SendSmtpEmailTo{
			Email: address.Address,
			Name:  address.Name,
		})
	}

	return recipients, nil
}

func parseCc(values []string) ([]brevo.SendSmtpEmailCc, error) {
	recipients := make([]brevo.SendSmtpEmailCc, 0, len(values))

	for _, value := range values {
		address, err := mail.ParseAddress(strings.TrimSpace(value))
		if err != nil {
			return nil, fmt.Errorf("invalid email address %q: %w", value, err)
		}

		recipients = append(recipients, brevo.SendSmtpEmailCc{
			Email: address.Address,
			Name:  address.Name,
		})
	}

	return recipients, nil
}

func parseBcc(values []string) ([]brevo.SendSmtpEmailBcc, error) {
	recipients := make([]brevo.SendSmtpEmailBcc, 0, len(values))

	for _, value := range values {
		address, err := mail.ParseAddress(strings.TrimSpace(value))
		if err != nil {
			return nil, fmt.Errorf("invalid email address %q: %w", value, err)
		}

		recipients = append(recipients, brevo.SendSmtpEmailBcc{
			Email: address.Address,
			Name:  address.Name,
		})
	}

	return recipients, nil
}

func parseSender(value string) (*brevo.SendSmtpEmailSender, error) {
	address, err := mail.ParseAddress(strings.TrimSpace(value))
	if err != nil {
		return nil, fmt.Errorf("invalid email address %q: %w", value, err)
	}

	return &brevo.SendSmtpEmailSender{
		Email: address.Address,
		Name:  address.Name,
	}, nil
}

func parseReplyTo(value string) (*brevo.SendSmtpEmailReplyTo, error) {
	address, err := mail.ParseAddress(strings.TrimSpace(value))
	if err != nil {
		return nil, fmt.Errorf("invalid email address %q: %w", value, err)
	}

	return &brevo.SendSmtpEmailReplyTo{
		Email: address.Address,
		Name:  address.Name,
	}, nil
}

func (s *BrevoSender) Send(
	ctx context.Context,
	message email.Message,
) error {
	if len(message.To) == 0 {
		return errors.New("at least one recipient is required")
	}

	if strings.TrimSpace(message.Subject) == "" {
		return errors.New("email subject is required")
	}

	if message.HTML == "" && message.Text == "" {
		return errors.New("email body is required")
	}

	from := message.From
	if strings.TrimSpace(from) == "" {
		from = s.from
	}

	sender, err := parseSender(from)
	if err != nil {
		return fmt.Errorf("invalid sender: %w", err)
	}

	to, err := parseTo(message.To)
	if err != nil {
		return fmt.Errorf("invalid recipients: %w", err)
	}

	cc, err := parseCc(message.Cc)
	if err != nil {
		return fmt.Errorf("invalid cc recipients: %w", err)
	}

	bcc, err := parseBcc(message.Bcc)
	if err != nil {
		return fmt.Errorf("invalid bcc recipients: %w", err)
	}

	params := brevo.SendSmtpEmail{
		Sender:      sender,
		To:          to,
		Cc:          cc,
		Bcc:         bcc,
		Subject:     message.Subject,
		HtmlContent: message.HTML,
		TextContent: message.Text,
	}

	if len(message.ReplyTo) > 0 {
		replyTo, err := parseReplyTo(message.ReplyTo)
		if err != nil {
			return fmt.Errorf("invalid reply-to address: %w", err)
		}

		params.ReplyTo = replyTo
	}

	_, _, err = s.client.TransactionalEmailsApi.SendTransacEmail(ctx, params)
	if err != nil {
		return fmt.Errorf("send email through brevo: %w", err)
	}

	return nil
}
