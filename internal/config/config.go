package config

import (
	"fmt"
	"os"
	"time"

	"github.com/joho/godotenv"
	"github.com/spf13/viper"
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

	viper.AddConfigPath("./config")
	if err := viper.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("error reading config file: %w", err)
	}

	cfg := &Config{}
	if err := viper.Unmarshal(cfg); err != nil {
		return nil, fmt.Errorf("failed to decode config file into struct: %w", err)
	}

	cfg.Database.Password = os.Getenv("DB_PASSWORD")
	cfg.JWT.Secret = os.Getenv("JWT_SECRET")

	return cfg, nil
}
