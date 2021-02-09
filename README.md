# gcp-env

A small library and binary for securely handling secrets in environment variables on GCP. Supports KMS and Secrets Manager. It's the sister project of [aws-env](https://github.com/telia-oss/aws-env), which itself is inspired by [ssm-env](https://github.com/remind101/ssm-env).

## Usage

Both the library and binary versions of `gcp-env` will loop through the environment and exchange any variables prefixed with
`sm://` and `kms://` with their secret value from Secrets manager or KMS respectively. In order to resolve Google secrets from Google Secret Manager, `gcp-env` should run under IAM role that has permission to access desired secrets.

This can be achieved by assigning IAM Role to Kubernetes Pod with Workload Identity. It's possible to assign IAM Role to GCE instance, where container is running, but this option is less secure.

For instance:
- `export SECRETSMANAGER=sm://<path>`
- `export KMSENCRYPTED=kms://<encrypted-secret>`

Where `<path>` is the name of the secret in secrets manager, and encrypted secret is a base64 cipher text

## Binary

Grab a binary from the [releases](https://github.com/telia-oss/gcp-env/releases) and start your process with:

```bash
gcp-env exec -- <command>
```

This will populate all the secrets in the environment, and hand over the process to your `<command>` with the same PID. The populated secrets are only made available to the `<command>` and 'disappear' when the process exits.

## Library

Import the library and invoke it prior to parsing flags or reading environment variables:

```go
package main

import (
	"context"
	"fmt"
	"os"
	"strings"

	environment "github.com/telia-oss/gcp-env/pkg/environment"
	utils "github.com/telia-oss/gcp-env/pkg/utils"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

const (
	cloudkmsScope      = "https://www.googleapis.com/auth/cloudkms"
	cloudPlatformScope = "https://www.googleapis.com/auth/cloud-platform"
)

func main() {
	// Populate secrets using gcp-env

	ctx := context.Background()

	oAuthCredentials := os.Getenv("GOOGLE_OAUTH_ACCESS_TOKEN")

	var creds *google.Credentials
	var err error
	if len(oAuthCredentials) > 0 {
		var contents string
		contents, _, err = utils.PathOrContents(oAuthCredentials)
		if err != nil {
			panic(fmt.Errorf("failed to initialize credentials for Google Cloud SDK in gcp-env: %s from GOOGLE_OAUTH_ACCESS_TOKEN", err))
		}
		token := &oauth2.Token{AccessToken: contents}
		creds = &google.Credentials{
			TokenSource: utils.StaticTokenSource{oauth2.StaticTokenSource(token)},
		}
	} else {
		creds, err = google.FindDefaultCredentials(ctx, cloudkmsScope, cloudPlatformScope)
		if err != nil {
			panic(fmt.Errorf("failed to initialize credentials for Google Cloud SDK in gcp-env: %s", err))
		}
	}

	env, err := environment.New(ctx, creds)
	if err != nil {
		panic(fmt.Errorf("failed to initialize gcp-env: %s", err))
	}
	if err := env.Populate(); err != nil {
		panic(fmt.Errorf("failed to populate environment: %s", err))
	}

	fmt.Printf("List of environment variables\n")

	envs := make(map[string]string)
	for _, v := range os.Environ() {
		name, value := parseEnvironmentVariable(v)
		envs[name] = value
	}
	fmt.Printf("%v", envs)
}

func parseEnvironmentVariable(s string) (string, string) {
	pair := strings.SplitN(s, "=", 2)
	return pair[0], pair[1]
}
```

## Security

There are a couple of things to keep in mind when using `gcp-env`:

- Spawned processes will inherit their parents environment by default. If your `<command>` spawns new processes they will inherit the environment _with the secrets already populated_, unless you hand-roll the environment for the new process.
- The environment for a running process can be read by the root user (and yourself) _after secrets have been populated_ by running `cat /proc/<pid>/environ` on Linux, and `ps eww <pid>` on OSX. However, if root or the spawning user is compromised a malicious user can just as easily fetch the secrets directly from the GCP API ¯\\_(ツ)_/¯
