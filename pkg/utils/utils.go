package utils

import (
	"io/ioutil"
	"os"

	"github.com/mitchellh/go-homedir"
	"golang.org/x/oauth2"
)

// PathOrContents reads content or file path
func PathOrContents(poc string) (string, bool, error) {
	if poc == "" {
		return poc, false, nil
	}

	path := poc
	if path[0] == '~' {
		var err error
		path, err = homedir.Expand(path)
		if err != nil {
			return path, true, err
		}
	}

	if _, err := os.Stat(path); err == nil {
		contents, err := ioutil.ReadFile(path)
		if err != nil {
			return string(contents), true, err
		}
		return string(contents), true, nil
	}

	return poc, false, nil
}

// StaticTokenSource is used to be able to identify static token sources without reflection.
type StaticTokenSource struct {
	oauth2.TokenSource
}
