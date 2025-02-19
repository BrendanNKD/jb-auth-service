package secretmanager

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"os"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/secretsmanager"
)

// LoadSecretToEnv fetches a secret from AWS Secrets Manager for the given secretName and region,
// parses the secret (expected to be a JSON object), maps the keys as needed, and sets each
// key/value pair as an environment variable.
func LoadSecretToEnv(secretName, region string) error {
	// Load AWS configuration.
	cfg, err := config.LoadDefaultConfig(context.TODO(), config.WithRegion(region))
	if err != nil {
		return fmt.Errorf("failed to load AWS config: %w", err)
	}

	// Create a Secrets Manager client.
	svc := secretsmanager.NewFromConfig(cfg)

	// Build the request.
	input := &secretsmanager.GetSecretValueInput{
		SecretId: aws.String(secretName),
	}

	// Retrieve the secret.
	result, err := svc.GetSecretValue(context.TODO(), input)
	if err != nil {
		return fmt.Errorf("failed to get secret value: %w", err)
	}

	// Get the secret value as a string.
	var secretValue string
	if result.SecretString != nil {
		secretValue = *result.SecretString
	} else if result.SecretBinary != nil {
		decoded, err := base64.StdEncoding.DecodeString(string(result.SecretBinary))
		if err != nil {
			return fmt.Errorf("failed to decode secret binary: %w", err)
		}
		secretValue = string(decoded)
	}

	// Attempt to parse the secretValue as JSON.
	var secretMap map[string]interface{}
	if err := json.Unmarshal([]byte(secretValue), &secretMap); err == nil {
		// Define key mappings: secret JSON key -> environment variable key.
		keyMapping := map[string]string{
			"username":             "POSTGRES_USER",
			"password":             "POSTGRES_PASSWORD",
			"host":                 "POSTGRES_HOST",
			"port":                 "POSTGRES_PORT",
			"dbInstanceIdentifier": "POSTGRES_DB",
		}

		// Iterate through the secret map.
		for key, val := range secretMap {
			// Convert the value to a string.
			var value string
			switch v := val.(type) {
			case string:
				value = v
			case float64:
				// Convert numeric values (like port) to a string.
				value = fmt.Sprintf("%.0f", v)
			default:
				value = fmt.Sprintf("%v", v)
			}

			// If there is a mapping, use it; otherwise, use the original key.
			if mappedKey, exists := keyMapping[key]; exists {
				os.Setenv(mappedKey, value)
			} else {
				os.Setenv(key, value)
			}
		}
	} else {
		// If not JSON, store the raw value using the secretName as the key.
		os.Setenv(secretName, secretValue)
	}

	return nil
}
