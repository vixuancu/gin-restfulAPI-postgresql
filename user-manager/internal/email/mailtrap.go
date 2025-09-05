package email

import "user-management-api/pkg/logger"

type MailtrapConfig struct {} 


type MailTrapProvider struct {
	config *MailConfig
	logger *logger.Logger
}

func NewMailTrapProvider(config *MailConfig) (EmailProviderService, error) {
	return &MailTrapProvider{
		config: &MailtrapConfig{},
		logger: config.Logger,
	}, nil
}