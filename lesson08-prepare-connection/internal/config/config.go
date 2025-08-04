package config

import (
	"fmt"
	"hoc-gin/internal/utils"
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
	DB DatabaseConfig
}

func NewConfig() *Config {
	return &Config{
		DB: DatabaseConfig{
			Host:     utils.GetEnv("DB_HOST", "localhost"),
			Port:     utils.GetEnv("DB_PORT", "5433"),
			User:     utils.GetEnv("DB_USER", "vixuancu"),
			Password: utils.GetEnv("DB_PASSWORD", "123456"),
			DBName:   utils.GetEnv("DB_NAME", "master-golang"),
			SSLMode:  utils.GetEnv("DB_SSLMODE", "disable"),
		}}
}

func (c *Config) DNS() string {
	return fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		c.DB.Host, c.DB.Port, c.DB.User, c.DB.Password, c.DB.DBName, c.DB.SSLMode)
}
