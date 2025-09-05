package email

import (
	"context"
	"time"
	"user-management-api/internal/config"
	"user-management-api/internal/utils"
	"user-management-api/pkg/logger"

	"github.com/rs/zerolog"
)

type Email struct {
	From     Address   `json:"from"`
	To       []Address `json:"to"`
	Subject  string    `json:"subject"`
	Text     string    `json:"text"`
	Category string    `json:"category"`
}

type Address struct {
	Email string `json:"email"`
	Name  string `json:"name,omitempty"`
}
type MailConfig struct {
	ProviderConfig map[string]any
	ProviderType   ProviderType
	MaxRetries     int
	Timeout        time.Duration
	Logger         *zerolog.Logger
}
type MailService struct {
	config   *MailConfig
	provider EmailProviderService
	logger   *zerolog.Logger
}

func NewMailService(cfg *config.Config, logger *zerolog.Logger, providerFactory ProviderFactory) (EmailProviderService, error) {
	config := &MailConfig{
		ProviderConfig: cfg.MailProviderConfig,
		ProviderType:   ProviderType(cfg.MailProviderType),
		MaxRetries:     3,
		Timeout:        10 * time.Second,
		Logger:         logger,
	}
	provider,err := providerFactory.CreateProvider(config)
	if err != nil {
		return nil, err
	}
	return &MailService{
		config:   config,
		provider: provider,
		logger:   logger,
	}, nil
}

func (ms *MailService) SendEmail(ctx context.Context, email *Email) error {
	traceID := logger.GetTraceID(ctx)
	start := time.Now()

	var lastErr error
	for attempt := 1; attempt <= ms.config.MaxRetries; attempt++ {
		startAttempt := time.Now()
		err := ms.provider.SendEmail(ctx, email)
		if err == nil {
			ms.logger.Info().Str("trace_id", traceID).
				Dur("duration", time.Since(startAttempt)).
				Str("operation", "SendEmail").
				Interface("to", email.To).
				Interface("subject", email.Subject).
				Str("category", email.Category).
				Int("attempt", attempt).
				Msg("Email sent successfully")
			return nil
		}
		lastErr = err
		ms.logger.Warn().Str("trace_id", traceID).
			Dur("duration", time.Since(startAttempt)).
			Str("operation", "SendEmail").
			Int("attempt", attempt).
			Err(err).
			Msg("Failed to send email, retrying...")
		time.Sleep(time.Duration(attempt) * time.Second)
	}
	ms.logger.Error().Str("trace_id", traceID).
		Dur("duration", time.Since(start)).
		Str("operation", "SendEmail").
		Int("attempts", ms.config.MaxRetries).
		Err(lastErr).
		Msg("All attempts to send email failed")
	return utils.WrapError(lastErr, "failed to send email after retries", utils.ErrorCodeInternalServer)
}
