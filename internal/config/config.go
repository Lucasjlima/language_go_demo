package config

import (
	"errors"
	"fmt"
	"strings"

	"github.com/spf13/viper"
)

const defaultGeminiModel = "gemini-2.5-flash"

type Config struct {
	GeminiAPIKey string
	GeminiModel  string
}

func Load() (Config, error) {
	v := viper.New()
	v.SetConfigFile(".env")
	v.SetConfigType("env")
	v.AutomaticEnv()
	v.SetDefault("GEMINI_MODEL", defaultGeminiModel)

	if err := v.ReadInConfig(); err != nil {
		var configNotFound viper.ConfigFileNotFoundError
		if !errors.As(err, &configNotFound) {
			return Config{}, fmt.Errorf("load config: %w", err)
		}
	}

	cfg := Config{
		GeminiAPIKey: strings.TrimSpace(v.GetString("GEMINI_API_KEY")),
		GeminiModel:  strings.TrimSpace(v.GetString("GEMINI_MODEL")),
	}

	if cfg.GeminiModel == "" {
		cfg.GeminiModel = defaultGeminiModel
	}

	if cfg.GeminiAPIKey == "" {
		return Config{}, errors.New("missing GEMINI_API_KEY in .env")
	}

	return cfg, nil
}
