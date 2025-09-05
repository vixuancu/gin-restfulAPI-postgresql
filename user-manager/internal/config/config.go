package config

import (
	"fmt"
	"os"
	"user-management-api/internal/utils"
)

type DatabaseConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	DBName   string
	SSLMode  string
}

type Config struct {
	ServerAddress string 
	DB DatabaseConfig
	MailProviderType   string
	MailProviderConfig map[string]any
}

func NewConfig() *Config {
	mailProviderConfig := make(map[string]any)
	mailProviderType := utils.GetEnv("MAIL_PROVIDER_TYPE", "mailtrap")
	if mailProviderType == "mailtrap" {
		mailtrapConfig := map[string]any{
			"mail_sender": utils.GetEnv("MAILTRAP_SENDER_EMAIL","vixuancu2004@gmail.com"),
			"name_sender": utils.GetEnv("MAILTRAP_SENDER_NAME","Vi Xuân Cử"),
			"mailtrap_url": utils.GetEnv("MAILTRAP_URL","https://sandbox.api.mailtrap.io/api/send/4011715"),
			"mailtrap_api_key": utils.GetEnv("MAILTRAP_API_KEY","11c42a24a41917f166b9938cb87190f6"),
		}
		mailProviderConfig["mailtrap"] = mailtrapConfig
	}
	return &Config{
		ServerAddress: fmt.Sprintf(":%s", os.Getenv("SERVER_PORT")),
		DB: DatabaseConfig{
			Host:     utils.GetEnv("DB_HOST", "localhost"),
			Port:     utils.GetEnv("DB_PORT", "5433"),
			User:     utils.GetEnv("DB_USER", "vixuancu"),
			Password: utils.GetEnv("DB_PASSWORD", "123456"),
			DBName:   utils.GetEnv("DB_NAME", "master-golang"),
			SSLMode:  utils.GetEnv("DB_SSLMODE", "disable"),
		},
		MailProviderType:   mailProviderType,
		MailProviderConfig: mailProviderConfig, 
	}
}

func (c *Config) DNS() string {
	return fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		c.DB.Host, c.DB.Port, c.DB.User, c.DB.Password, c.DB.DBName, c.DB.SSLMode)
}