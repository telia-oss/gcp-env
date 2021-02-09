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
