// Copyright 2023 CodeMaker AI Inc. All rights reserved.

package cli

import (
	"fmt"
	"github.com/codemakerai/codemaker-sdk-go/client"
	"github.com/joho/godotenv"
	"os"
	"path/filepath"
)

const (
	apiKeyEnvironmentVariable = "CODEMAKER_API_KEY"
)

func createConfig() (*client.Config, error) {
	apiKey := os.Getenv(apiKeyEnvironmentVariable)

	if len(apiKey) == 0 {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return nil, fmt.Errorf("failed to load configuration: no user home dir found")
		}

		env, err := godotenv.Read(filepath.Join(homeDir, ".codemaker", "config"))
		if err != nil {
			return nil, fmt.Errorf("failed to load configuration from config file")
		}

		apiKey = env[apiKeyEnvironmentVariable]
	}

	if len(apiKey) == 0 {
		return nil, fmt.Errorf("failed to resolve CODEMAKER_API_KEY")
	}

	return &client.Config{
		ApiKey: apiKey,
	}, nil
}
