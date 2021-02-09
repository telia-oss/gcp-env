//Package environment package manages environments
package environment

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/telia-oss/gcp-env/internal/secrets"
	"golang.org/x/oauth2/google"
)

const (
	// The secret should be in the format (optionally with version)
	// kms://{base64}
	kmsPrefix = "kms://"
	// The secret name should be in the format (optionally with version)
	// `sm://projects/{PROJECT_ID}/secrets/{SECRET_NAME}`
	// `sm://projects/{PROJECT_ID}/secrets/{SECRET_NAME}/versions/{VERSION|latest}`
	smPrefix = "sm://"
)

// Manager handles API calls to AWS.
type Manager struct {
	SecretProvider *secrets.Provider
}

// New creates a new manager for populating secret values.
func New(ctx context.Context, creds *google.Credentials) (*Manager, error) {
	provider, err := secrets.NewClient(ctx, creds)
	if err != nil {
		return nil, err
	}
	return &Manager{
		SecretProvider: provider,
	}, nil
}

// Populate environment variables with their secret values from Secrets manager,
func (m *Manager) Populate() error {
	env := make(map[string]string)
	for _, v := range os.Environ() {
		var (
			found  bool
			secret string
			err    error
		)

		name, value := parseEnvironmentVariable(v)

		// Improve this with provider factory
		if strings.HasPrefix(value, kmsPrefix) {
			secret, err = m.SecretProvider.ResolveSecret(value)
			if err != nil {
				return fmt.Errorf("failed to decrypt kms secret: '%s': %s", value, err)
			}
			found = true
		} else if strings.HasPrefix(value, smPrefix) {
			secret, err = m.SecretProvider.ResolveSecret(value)
			if err != nil {
				return fmt.Errorf("failed to decrypt kms secret: '%s': %s", value, err)
			}
			found = true
		}

		if found {
			env[name] = secret
		}
	}

	for name, secret := range env {
		if err := os.Setenv(name, secret); err != nil {
			return fmt.Errorf("failed to set environment variable: '%s': %s", name, err)
		}
	}
	return nil
}

func parseEnvironmentVariable(s string) (string, string) {
	pair := strings.SplitN(s, "=", 2)
	return pair[0], pair[1]
}
