package resend

import (
	"context"
	"errors"
	"stitchforce-server/config"
	"stitchforce-server/pkg/email"
	"strings"

	"github.com/resend/resend-go/v4"
)

type ResendConfig struct {
    APIKey string
    From   string
}

type ResendSender struct {
    client *resend.Client
    from   string
}

func New(cfg config.ResendConfig) (*ResendSender, error) {
    if strings.TrimSpace(cfg.APIKey) == "" {
        return nil, errors.New("resend API key is required")
    }

    if strings.TrimSpace(cfg.From) == "" {
        return nil, errors.New("email sender address is required")
    }

    return &ResendSender{
        client: resend.NewClient(cfg.APIKey),
        from:   cfg.From,
    }, nil
}

func (s *ResendSender) Send(
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

    params := &resend.SendEmailRequest{
        From:    from,
        To:      message.To,
        Cc:      message.Cc,
        Bcc:     message.Bcc,
        ReplyTo: message.ReplyTo,
        Subject: message.Subject,
        Html:    message.HTML,
        Text:    message.Text,
    }

    _, err := s.client.Emails.SendWithContext(ctx, params)
    if err != nil {
        return err
    }

    return nil
}
