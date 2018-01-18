package main

import (
	"flag"
	"testing"
)

var (
	auth0TestToken = flag.String("auth0-token", "", "The raw token string received from an invocation of {auth0-domain}/oauth/token}")
)

func TestAuth0(t *testing.T) {
	err := validateToken("Bearer "+*auth0TestToken, "all:experiments")
	if err != nil {
		t.Error("Expected nil, got ", err.Error())
	}
}
