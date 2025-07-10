package internal

import (
	"fmt"
	"os"
	"strings"

	"github.com/joho/godotenv"
)

var envVars = []string{
	"BLOCKDAEMON_API_KEY",
	"SOLANA_ADDRESSES",
}

func CheckEnvVars() error {
	// err := godotenv.Load("../.env")
	err := godotenv.Load()
	if err != nil {
		return err
	}

	var missing []string
	for _, envVar := range envVars {
		if os.Getenv(envVar) == "" {
			missing = append(missing, envVar)
		}
	}
	if len(missing) > 0 {
		return fmt.Errorf("missing required environment variables: %s", strings.Join(missing, ", "))
	}
	return nil
}
