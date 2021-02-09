package secrets_test

import (
	"context"
	"errors"
	"reflect"
	"testing"

	secrets "github.com/telia-oss/gcp-env/internal/secrets"
	"github.com/telia-oss/gcp-env/internal/secrets/secretsfakes"
	secretspb "google.golang.org/genproto/googleapis/cloud/secretmanager/v1"
)

func TestSecretsProvider_ResolveSecrets(t *testing.T) {
	type fields struct {
		sm secrets.GoogleSecretsManagerAPI
	}
	type args struct {
		ctx  context.Context
		vars []string
	}
	tests := []struct {
		name                       string
		fields                     fields
		args                       args
		secretsfakeserviceProvider func(context.Context, *secretsfakes.FakeGoogleSecretsManagerAPI) secrets.Provider
		want                       []string
		wantErr                    bool
	}{
		{
			name: "get implicit (latest) version single secret from Secrets Manager",
			args: args{
				ctx: context.TODO(),
				vars: []string{
					"sm://projects/test-project-id/secrets/test-secret",
				},
			},
			want: []string{
				"test-secret-value",
			},
			secretsfakeserviceProvider: func(ctx context.Context, fakeSecretManagerAPI *secretsfakes.FakeGoogleSecretsManagerAPI) secrets.Provider {
				//	req := secretspb.AccessSecretVersionRequest{
				//		Name: "projects/test-project-id/secrets/test-secret/versions/latest",
				//	}
				res := &secretspb.AccessSecretVersionResponse{Payload: &secretspb.SecretPayload{
					Data: []byte("test-secret-value"),
				}}
				fakeSecretManagerAPI.AccessSecretVersionReturns(res, nil)
				sp := secrets.Provider{SMClient: fakeSecretManagerAPI}
				return sp
			},
		},
		{
			name: "get explicit single secret version from Secrets Manager",
			args: args{
				ctx: context.TODO(),
				vars: []string{
					"sm://projects/test-project-id/secrets/test-secret/versions/5",
				},
			},
			want: []string{
				"test-secret-value",
			},
			secretsfakeserviceProvider: func(ctx context.Context, fakeSecretManagerAPI *secretsfakes.FakeGoogleSecretsManagerAPI) secrets.Provider {
				//req := secretspb.AccessSecretVersionRequest{
				//	Name: "projects/test-project-id/secrets/test-secret/versions/5",
				//}
				res := &secretspb.AccessSecretVersionResponse{Payload: &secretspb.SecretPayload{
					Data: []byte("test-secret-value"),
				}}
				fakeSecretManagerAPI.AccessSecretVersionReturns(res, nil)
				sp := secrets.Provider{SMClient: fakeSecretManagerAPI}
				return sp
			},
		},
		{
			name: "get 2 secrets from Secrets Manager",
			args: args{
				ctx: context.TODO(),
				vars: []string{
					"sm://projects/test-project-id/secrets/test-secret/versions/5",
					"hello",
					"sm://projects/test-project-id/secrets/test-secret",
				},
			},
			want: []string{
				"test-secret-value-1",
				"hello",
				"test-secret-value-2",
			},
			secretsfakeserviceProvider: func(ctx context.Context, fakeSecretManagerAPI *secretsfakes.FakeGoogleSecretsManagerAPI) secrets.Provider {
				sp := secrets.Provider{SMClient: fakeSecretManagerAPI}
				vars := map[string]string{
					"sm://projects/test-project-id/secrets/test-secret/versions/5":      "test-secret-value-1",
					"sm://projects/test-project-id/secrets/test-secret/versions/latest": "test-secret-value-2",
				}
				call := 0
				for _, v := range vars {

					//name := n
					value := v
					//req := secretspb.AccessSecretVersionRequest{
					//	Name: name,
					//}
					res := &secretspb.AccessSecretVersionResponse{Payload: &secretspb.SecretPayload{
						Data: []byte(value),
					}}
					//fakeSecretManagerAPI.AccessSecretVersionReturns(res, nil)
					fakeSecretManagerAPI.AccessSecretVersionReturnsOnCall(call, res, nil)
					call++
				}
				return sp
			},
		},
		{
			name: "no secrets",
			args: args{
				ctx: context.TODO(),
				vars: []string{
					"hello-1",
					"hello-2",
				},
			},
			want: []string{
				"hello-1",
				"hello-2",
			},
			secretsfakeserviceProvider: func(ctx context.Context, fakeSecretManagerAPI *secretsfakes.FakeGoogleSecretsManagerAPI) secrets.Provider {
				return secrets.Provider{SMClient: fakeSecretManagerAPI}
			},
		},
		{
			name: "error getting secret from Secrets Manager",
			args: args{
				ctx: context.TODO(), vars: []string{
					"gcp+sm://projects/test-project-id/secrets/test-secret",
					"hello",
				},
			},
			want: []string{
				"",
				"hello",
			},
			wantErr: true,
			secretsfakeserviceProvider: func(ctx context.Context, fakeSecretManagerAPI *secretsfakes.FakeGoogleSecretsManagerAPI) secrets.Provider {
				sp := secrets.Provider{SMClient: fakeSecretManagerAPI}
				//req := secretspb.AccessSecretVersionRequest{
				//	Name: "projects/test-project-id/secrets/test-secret/versions/latest",
				//}
				fakeSecretManagerAPI.AccessSecretVersionReturns(nil, errors.New("test-error"))
				return sp
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			secretsfakesM := &secretsfakes.FakeGoogleSecretsManagerAPI{}
			sp := tt.secretsfakeserviceProvider(tt.args.ctx, secretsfakesM)
			got, err := sp.ResolveSecrets(tt.args.vars)
			if (err != nil) != tt.wantErr {
				t.Errorf("SecretsProvider.ResolveSecrets() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("SecretsProvider.ResolveSecrets() = %v, want %v", got, tt.want)
				return
			}
		})
	}
}
