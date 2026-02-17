package config

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/joho/godotenv"
)

// Credentials holds the X API authentication keys.
type Credentials struct {
	APIKey            string
	APISecret         string
	AccessToken       string
	AccessTokenSecret string
	BearerToken       string
}

// LoadCredentials loads credentials from env vars, with .env fallback.
// Search order: ~/.config/x-cli/.env → ./.env → environment variables.
func LoadCredentials() (*Credentials, error) {
	// Try ~/.config/x-cli/.env
	home, err := os.UserHomeDir()
	if err == nil {
		configEnv := filepath.Join(home, ".config", "x-cli", ".env")
		_ = godotenv.Load(configEnv)
	}
	// Try ./.env
	_ = godotenv.Load()

	require := func(name string) (string, error) {
		val := os.Getenv(name)
		if val == "" {
			return "", fmt.Errorf("missing env var: %s. Set X_API_KEY, X_API_SECRET, X_ACCESS_TOKEN, X_ACCESS_TOKEN_SECRET, X_BEARER_TOKEN", name)
		}
		return val, nil
	}

	apiKey, err := require("X_API_KEY")
	if err != nil {
		return nil, err
	}
	apiSecret, err := require("X_API_SECRET")
	if err != nil {
		return nil, err
	}
	accessToken, err := require("X_ACCESS_TOKEN")
	if err != nil {
		return nil, err
	}
	accessTokenSecret, err := require("X_ACCESS_TOKEN_SECRET")
	if err != nil {
		return nil, err
	}
	bearerToken, err := require("X_BEARER_TOKEN")
	if err != nil {
		return nil, err
	}

	return &Credentials{
		APIKey:            apiKey,
		APISecret:         apiSecret,
		AccessToken:       accessToken,
		AccessTokenSecret: accessTokenSecret,
		BearerToken:       bearerToken,
	}, nil
}
