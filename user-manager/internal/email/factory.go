package email

import (
	"fmt"
	"user-management-api/internal/utils"
)

type ProviderType string

const (
	ProviderMailtrap ProviderType = "mailtrap"
)
type ProviderFactory interface {
	CreateProvider(config *MailConfig) (EmailProviderService, error)
}

type MailTrapProviderFactory struct{

}

func (f *MailTrapProviderFactory) CreateProvider(config *MailConfig) (EmailProviderService, error) {

	return NewMailTrapProvider(config)
}

func NewProviderFactory(ProviderType ProviderType) (ProviderFactory, error) {
	switch ProviderType {
	case ProviderMailtrap:
		return &MailTrapProviderFactory{}, nil
	default:
		return nil, utils.NewError(fmt.Sprintf("Unsupported email provider type: %s",ProviderType), utils.ErrorCodeInternalServer)
	}

}
