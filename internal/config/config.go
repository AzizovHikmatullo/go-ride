package config

import (
	"fmt"
	"os"
	"time"

	"github.com/joho/godotenv"
)

type Config struct {
	Server struct {
		Port string `mapstructure:"port"`
	} `mapstructure:"server"`

	Database struct {
		Host     string `mapstructure:"host"`
		Port     string `mapstructure:"port"`
		DBName   string `mapstructure:"dbname"`
		User     string `mapstructure:"username"`
		Password string
	} `mapstructure:"db"`

	JWT struct {
		Secret          string
		AccessTokenTTL  time.Duration `mapstructure:"access_token_ttl"`
		RefreshTokenTTL time.Duration `mapstructure:"refresh_token_ttl"`
	} `mapstructure:"jwt"`
}

func LoadConfig() (*Config, error) {
	err := godotenv.Load()
	if err != nil {
		return nil, fmt.Errorf("error reading .env file: %w", err)
	}

	cfg := &Config{}
	cfg.Server.Port = os.Getenv("SERVER_PORT")

	cfg.Database.Host = os.Getenv("DB_HOST")
	cfg.Database.Port = os.Getenv("DB_PORT")
	cfg.Database.DBName = os.Getenv("DB_DBNAME")
	cfg.Database.User = os.Getenv("DB_USERNAME")
	cfg.Database.Password = os.Getenv("DB_PASSWORD")

	cfg.JWT.Secret = os.Getenv("JWT_SECRET")

	accesTokenTTL, err := time.ParseDuration(os.Getenv("JWT_ACCESS_TOKEN_TTL"))
	if err != nil {
		return nil, fmt.Errorf("failed to convert access_token_ttl: %w", err)
	}
	cfg.JWT.AccessTokenTTL = accesTokenTTL

	refreshTokenTTL, err := time.ParseDuration(os.Getenv("JWT_REFRESH_TOKEN_TTL"))
	if err != nil {
		return nil, fmt.Errorf("failed to convert refresh_token_ttl: %w", err)
	}
	cfg.JWT.RefreshTokenTTL = refreshTokenTTL

	return cfg, nil
}
