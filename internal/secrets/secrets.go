package secrets

import (
	"context"
	"encoding/base64"
	"fmt"
	"os"
	"strings"

	kms "cloud.google.com/go/kms/apiv1"
	secretmanager "cloud.google.com/go/secretmanager/apiv1"
	kmspb "google.golang.org/genproto/googleapis/cloud/kms/v1"
	secretmanagerpb "google.golang.org/genproto/googleapis/cloud/secretmanager/v1"

	"github.com/googleapis/gax-go/v2"
	"github.com/pkg/errors"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/option"
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

// Provider Google Cloud API provider
type Provider struct {
	KMSClient GoogleKeyManagementAPI
	SMClient  GoogleSecretsManagerAPI
	ctx       context.Context
}

// NewClient is a global exported function that creates a new client
func NewClient(ctx context.Context, creds *google.Credentials) (*Provider, error) {
	kmsClient, err := kms.NewKeyManagementClient(ctx, option.WithCredentials(creds))
	if err != nil {
		return nil, errors.Wrap(err, "failed to initialize Google Cloud KMS SDK")
	}

	smClient, err := secretmanager.NewClient(ctx, option.WithCredentials(creds))
	if err != nil {
		return nil, errors.Wrap(err, "failed to initialize Google Cloud Secret Manager SDK")
	}
	if err != nil {
		return nil, err
	}
	client := NewSecretsProvider(ctx, kmsClient, smClient)
	if err != nil {
		return nil, err
	}
	return client, nil
}

// NewSecretsProvider is a global exported function that creates a new client
func NewSecretsProvider(ctx context.Context, kmsClient GoogleKeyManagementAPI, smClient GoogleSecretsManagerAPI) *Provider {
	return &Provider{
		KMSClient: kmsClient,
		SMClient:  smClient,
		ctx:       ctx,
	}
}

// ResolveSecret provides and interface to resolve a secret
func (s *Provider) ResolveSecret(value string) (secret string, err error) {
	if strings.HasPrefix(value, kmsPrefix) {
		secret, err = s.decrypt(strings.TrimPrefix(value, kmsPrefix))
		if err != nil {
			return "", fmt.Errorf("failed to decrypt kms secret: '%s': %s", value, err)
		}
	} else if strings.HasPrefix(value, smPrefix) {
		secret, err = s.getSecretValue(strings.TrimPrefix(value, smPrefix))
		if err != nil {
			return "", fmt.Errorf("failed to fetch sm secret: '%s': %s", value, err)
		}
	} else {
		return "", fmt.Errorf("failed to fetch unsupported secret: '%s': %s", value, err)
	}

	return secret, nil
}

// ResolveSecrets provides and interface to resolve a list of secrets
func (s *Provider) ResolveSecrets(values []string) ([]string, error) {
	var secretlist []string
	var parseError error
	var errorValues []string
	for _, v := range values {
		value, err := s.ResolveSecret(v)
		if err != nil {
			errorValues = append(errorValues, value)
			parseError = err
		}
		secretlist = append(secretlist, value)
	}
	if len(errorValues) > 0 {
		parseError = fmt.Errorf("failed to resolve secrets: '%s': %s", strings.Join(errorValues, ","), parseError)
	}
	return secretlist, parseError
}

func (s *Provider) getSecretValue(path string) (string, error) {
	// if no version specified add latest
	if !strings.Contains(path, "/versions/") {
		path += "/versions/latest"
	}
	// get secret value
	accessReq := &secretmanagerpb.AccessSecretVersionRequest{
		Name: path,
	}

	secret, err := s.SMClient.AccessSecretVersion(s.ctx, accessReq)
	if err != nil {
		return "", errors.Wrap(err, "failed to access secret from Google Secret Manager")
	}
	return string(secret.Payload.GetData()), nil
}

func (s *Provider) decrypt(ciphertext string) (string, error) {
	data, err := base64.StdEncoding.DecodeString(ciphertext)
	if err != nil {
		return "", fmt.Errorf("failed to decode base64 cipher: %s", err)
	}
	keyID := os.Getenv("KMS_KEY_ID")
	if len(keyID) < 1 {
		return "", fmt.Errorf("missing required keyID to decrypt: %s", s)
	}
	// decrypt secret value
	req := &kmspb.DecryptRequest{
		Name:       keyID,
		Ciphertext: data,
	}
	resp, err := s.KMSClient.Decrypt(s.ctx, req)
	if err != nil {
		return "", errors.Wrap(err, "failed to decrypt from Google Secret Manager")
	}
	return strings.TrimSpace(string(resp.Plaintext)), nil
}

// GoogleSecretsManagerAPI represents KeyManagementClient interface for stub
//go:generate go run github.com/maxbrunsfeld/counterfeiter/v6 . GoogleSecretsManagerAPI

type GoogleSecretsManagerAPI interface {
	// ref. https://pkg.go.dev/cloud.google.com/go/secretmanager/apiv1?tab=doc#example-Client.AccessSecretVersion
	AccessSecretVersion(ctx context.Context, req *secretmanagerpb.AccessSecretVersionRequest, opts ...gax.CallOption) (*secretmanagerpb.AccessSecretVersionResponse, error) // go:nolint
	GetSecret(ctx context.Context, req *secretmanagerpb.GetSecretRequest, opts ...gax.CallOption) (*secretmanagerpb.Secret, error)
}

// GoogleKeyManagementAPI represents KeyManagementClient interface for stub
//go:generate go run github.com/maxbrunsfeld/counterfeiter/v6 . GoogleKeyManagementAPI

type GoogleKeyManagementAPI interface {
	// ref. https://pkg.go.dev/cloud.google.com/go/kms/apiv1?tab=doc#example-KeyManagementClient.Decrypt
	Decrypt(ctx context.Context, req *kmspb.DecryptRequest, opts ...gax.CallOption) (*kmspb.DecryptResponse, error)
}
