package email

import (
	"context"
	"fmt"

	"github.com/resend/resend-go/v2"
)

type SendParams struct {
	To      string
	Subject string
	HTML    string
}

type Sender interface {
	Send(ctx context.Context, params SendParams) error
}

type ResendSender struct {
	client      *resend.Client
	fromAddress string
	fromName    string
}

func NewResendSender(apiKey string, fromAddress string, fromName string) *ResendSender {
	return &ResendSender{
		client:      resend.NewClient(apiKey),
		fromAddress: fromAddress,
		fromName:    fromName,
	}
}

func (s *ResendSender) Send(ctx context.Context, params SendParams) error {
	_, err := s.client.Emails.SendWithContext(ctx, &resend.SendEmailRequest{
		From:    fmt.Sprintf("%s <%s>", s.fromName, s.fromAddress),
		To:      []string{params.To},
		Subject: params.Subject,
		Html:    params.HTML,
	})
	return err
}
