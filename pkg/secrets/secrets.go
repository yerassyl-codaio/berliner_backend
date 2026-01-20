package secrets

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/secretsmanager"
)

// Client wraps AWS Secrets Manager client
type Client struct {
	svc    *secretsmanager.Client
	region string
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
	input := &secretsmanager.GetSecretValueInput{
		SecretId: &secretName,
	}

	result, err := c.svc.GetSecretValue(ctx, input)
	if err != nil {
		return "", fmt.Errorf("failed to get secret %s: %w", secretName, err)
	}

	if result.SecretString == nil {
		return "", fmt.Errorf("secret %s has no string value", secretName)
	}

	return *result.SecretString, nil
}
