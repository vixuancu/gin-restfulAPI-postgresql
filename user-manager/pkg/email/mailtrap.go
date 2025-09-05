package email

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"time"
	"user-management-api/internal/utils"
	"user-management-api/pkg/logger"

	"github.com/rs/zerolog"
)

type MailtrapConfig struct {
	MailSender     string
	NameSender     string
	MailtrapURL    string
	MailtrapAPIKey string
}

type MailTrapProvider struct {
	client *http.Client
	config *MailtrapConfig
	logger *zerolog.Logger
}

func NewMailTrapProvider(config *MailConfig) (EmailProviderService, error) {
	mailtrapConfig, ok := config.ProviderConfig["mailtrap"].(map[string]any)
	if !ok {
		return nil, utils.NewError("invalid mailtrap configuration", utils.ErrorCodeInternalServer)
	}

	return &MailTrapProvider{
		client: &http.Client{Timeout: config.Timeout},
		config: &MailtrapConfig{
			MailSender:     mailtrapConfig["mail_sender"].(string),
			NameSender:     mailtrapConfig["name_sender"].(string),
			MailtrapURL:    mailtrapConfig["mailtrap_url"].(string),
			MailtrapAPIKey: mailtrapConfig["mailtrap_api_key"].(string),
		},
		logger: config.Logger,
	}, nil
}
func (m *MailTrapProvider) SendEmail(ctx context.Context, email *Email) error {
	traceID := logger.GetTraceID(ctx)
	start := time.Now()
	email.From = Address{
		Email: m.config.MailSender,
		Name:  m.config.NameSender,
	}
	payload, err := json.Marshal(email)
	if err != nil {
		return utils.NewError("failed to marshal email payload", utils.ErrorCodeInternalServer)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, m.config.MailtrapURL, bytes.NewReader(payload))
	if err != nil {
		return utils.NewError("failed to create email request", utils.ErrorCodeInternalServer)
	}
	req.Header.Add("Authorization", "Bearer "+strings.TrimSpace(m.config.MailtrapAPIKey))
	req.Header.Add("Content-Type", "application/json")

	res, err := m.client.Do(req)
	if err != nil {
		m.logger.Error().Str("trace_id", traceID).
			Dur("duration", time.Since(start)).
			Str("operation", "SendEmail").
			Err(err).
			Msg("❌ Failed to send email request:")
		return utils.NewError("failed to send email request", utils.ErrorCodeInternalServer)
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(res.Body)
		m.logger.Error().Str("trace_id", traceID).
			Dur("duration", time.Since(start)).
			Str("operation", "SendEmail").
			Int("status_code", res.StatusCode).
			Str("response_body", string(body)).
			Msg("Unxpected response from email mailtrap")
		return utils.NewError("unexpected response from email provider", utils.ErrorCodeInternalServer)
	}
	return nil
}
