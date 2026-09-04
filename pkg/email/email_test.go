package email_test

import (
	"bytes"
	"context"
	"log/slog"
	"testing"

	"arsen/pkg/email"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNoopSender_ReturnsNil(t *testing.T) {
	sender := email.NewNoopSender(slog.Default())
	err := sender.Send(context.Background(), email.SendParams{
		To:      "user@example.com",
		Subject: "Welcome",
		HTML:    "<h1>Hello</h1>",
	})
	assert.NoError(t, err)
}

func TestNoopSender_LogsTheSend(t *testing.T) {
	var buf bytes.Buffer
	logger := slog.New(slog.NewJSONHandler(&buf, nil))

	sender := email.NewNoopSender(logger)
	err := sender.Send(context.Background(), email.SendParams{
		To:      "logged@example.com",
		Subject: "Test Subject",
		HTML:    "<p>body</p>",
	})
	require.NoError(t, err)

	output := buf.String()
	assert.Contains(t, output, "logged@example.com", "log output should contain the 'to' address")
	assert.Contains(t, output, "Test Subject", "log output should contain the subject")
}

func TestNoopSender_DoesNotErrorForAnyInput(t *testing.T) {
	sender := email.NewNoopSender(slog.Default())

	tests := []struct {
		name   string
		params email.SendParams
	}{
		{
			name:   "all empty strings",
			params: email.SendParams{To: "", Subject: "", HTML: ""},
		},
		{
			name:   "only to populated",
			params: email.SendParams{To: "a@b.c", Subject: "", HTML: ""},
		},
		{
			name:   "only subject populated",
			params: email.SendParams{To: "", Subject: "hey", HTML: ""},
		},
		{
			name:   "only html populated",
			params: email.SendParams{To: "", Subject: "", HTML: "<b>bold</b>"},
		},
		{
			name:   "unicode input",
			params: email.SendParams{To: "user@example.com", Subject: "Hej verden", HTML: "<p>Content</p>"},
		},
		{
			name:   "very long strings",
			params: email.SendParams{To: string(make([]byte, 10000)), Subject: string(make([]byte, 10000)), HTML: string(make([]byte, 10000))},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := sender.Send(context.Background(), tc.params)
			assert.NoError(t, err)
		})
	}
}

func TestNewResendSender_CreatesValidSender(t *testing.T) {
	sender := email.NewResendSender("fake-api-key", "noreply@example.com", "Test App")
	require.NotNil(t, sender, "NewResendSender should return a non-nil sender")
}
