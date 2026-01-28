package secrets

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/secretsmanager"
)

// Client wraps AWS Secrets Manager client
type Client struct {
	svc    *secretsmanager.Client
	region string
}

type SecretUserPass struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

// NewClient creates a new AWS Secrets Manager client
func NewClient(region string) (*Client, error) {
	cfg, err := config.LoadDefaultConfig(context.TODO(), config.WithRegion(region))
	if err != nil {
		return nil, fmt.Errorf("failed to load AWS config: %w", err)
	}

	return &Client{
		svc:    secretsmanager.NewFromConfig(cfg),
		region: region,
	}, nil
}

// GetSecret retrieves a secret value by its name/ARN
func (c *Client) GetSecret(ctx context.Context, secretName string) (string, error) {
	var sup SecretUserPass

	raw, err := c.getSecretRaw(ctx, secretName)
	if err != nil {
		return "", fmt.Errorf("failed to get secret %s: %w", secretName, err)
	}

	if err := json.Unmarshal([]byte(raw), &sup); err != nil {
		// JSON parsed successfully. Use provided fields.
		return "", fmt.Errorf("error unmarshaling: %w", err)
	}

	return sup.Password, nil
}

func (c *Client) getSecretRaw(ctx context.Context, secretName string) (string, error) {
	in := &secretsmanager.GetSecretValueInput{
		SecretId: aws.String(secretName),
	}

	out, err := c.svc.GetSecretValue(ctx, in)
	if err != nil {
		return "", fmt.Errorf("GetSecretValue error: %w", err)
	}

	// Prefer SecretString
	if out.SecretString != nil {
		return aws.ToString(out.SecretString), nil
	}

	// Fallback to SecretBinary (base64-encoded)
	if out.SecretBinary != nil {
		decoded, err := base64.StdEncoding.DecodeString(string(out.SecretBinary))
		if err != nil {
			return "", fmt.Errorf("failed to decode secret binary: %w", err)
		}
		return string(decoded), nil
	}

	return "", errors.New("secret contains no SecretString or SecretBinary")
}
