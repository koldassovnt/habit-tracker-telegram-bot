package config

import (
	"fmt"
	"os"
)

type Config struct {
	Telegram TelegramConfig
	Database DatabaseConfig
}

type TelegramConfig struct {
	Token string
}

type DatabaseConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	Name     string
}

func Load() (*Config, error) {
	cfg := &Config{
		Telegram: TelegramConfig{
			Token: os.Getenv("TELEGRAM_HABIT_TRACKER_TOKEN"),
		},
		Database: DatabaseConfig{
			Host:     os.Getenv("DB_HOST"),
			Port:     os.Getenv("DB_PORT"),
			User:     os.Getenv("DB_USER"),
			Password: os.Getenv("DB_PASSWORD"),
			Name:     os.Getenv("DB_NAME"),
		},
	}

	if err := cfg.validate(); err != nil {
		return nil, err
	}

	return cfg, nil
}

func (c *Config) validate() error {
	if c.Telegram.Token == "" {
		return fmt.Errorf("TELEGRAM_HABIT_TRACKER_TOKEN is not set")
	}
	if c.Database.Host == "" {
		return fmt.Errorf("DB_HOST is not set")
	}
	if c.Database.Port == "" {
		return fmt.Errorf("DB_PORT is not set")
	}
	if c.Database.User == "" {
		return fmt.Errorf("DB_USER is not set")
	}
	if c.Database.Password == "" {
		return fmt.Errorf("DB_PASSWORD is not set")
	}
	if c.Database.Name == "" {
		return fmt.Errorf("DB_NAME is not set")
	}
	return nil
}

// PgxDSN is used by pgxpool
func (d *DatabaseConfig) PgxDSN() string {
	return fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		d.Host, d.Port, d.User, d.Password, d.Name)
}

// FlywayDSN is used by Flyway JDBC URL
func (d *DatabaseConfig) FlywayDSN() string {
	return fmt.Sprintf("%s:%s/%s?user=%s&password=%s",
		d.Host, d.Port, d.Name, d.User, d.Password)
}
