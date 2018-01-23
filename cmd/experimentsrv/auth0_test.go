package main

import (
	"flag"
	"testing"
)

var (
	auth0TestToken = flag.String("auth0-token", "", "The raw token string received from an invocation of {auth0-domain}/oauth/token}")

	auth0TestSkip = flag.Bool("auth0-skip-test", false, "Must be set to prevent token testing from aborting tests if the token is not supplied")
)

func TestAuth0(t *testing.T) {
	// Although this test can be skipped the user must have specified a flag to prevent it from being fatal

	if len(*auth0TestToken) == 0 {
		if *auth0TestSkip {
			return
		} else {
			t.Error("the token test skip flag (-auth0-skip-test) must be set if a token is not supplied")
		}
	}

	err := validateToken("Bearer "+*auth0TestToken, "all:experiments")
	if err != nil {
		t.Error("expected nil, got ", err.Error())
	}
}

// TODO Add tests where the user name and password are supplied and we can use these to generate
// tokens during the test
