package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"syscall"

	flags "github.com/jessevdk/go-flags"
	environment "github.com/telia-oss/gcp-env/pkg/environment"
	"github.com/telia-oss/gcp-env/pkg/utils"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

var command rootCommand
var version string

type rootCommand struct {
	Version func()      `short:"v" long:"version" description:"Print the version and exit."`
	Exec    execCommand `command:"exec" description:"Execute a command."`
}

const (
	cloudkmsScope      = "https://www.googleapis.com/auth/cloudkms"
	cloudPlatformScope = "https://www.googleapis.com/auth/cloud-platform"
)

type execCommand struct{}

// Execute the exec subcommand.
func (c *execCommand) Execute(args []string) error {
	if len(args) < 1 {
		return errors.New("please supply a command to run")
	}
	var err error

	path, err := exec.LookPath(args[0])
	if err != nil {
		return fmt.Errorf("failed to validate command: %s", err)
	}

	oAuthCredentials := os.Getenv("GOOGLE_OAUTH_ACCESS_TOKEN")
	ctx := context.Background()

	var creds *google.Credentials
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
		return fmt.Errorf("failed to initialize gcp-env: %s", err)
	}
	if err := env.Populate(); err != nil {
		return fmt.Errorf("failed to populate environment: %s", err)
	}

	if err := syscall.Exec(path, args, os.Environ()); err != nil {
		return fmt.Errorf("failed to execute command: %s", err)
	}
	return nil
}

func init() {
	command.Version = func() {
		fmt.Println(version)
		os.Exit(0)
	}
}

func main() {
	_, err := flags.Parse(&command)
	if err != nil {
		if flagsErr, ok := err.(*flags.Error); ok && flagsErr.Type == flags.ErrHelp {
			os.Exit(0)
		} else {
			os.Exit(1)
		}
	}

}
