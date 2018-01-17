package main

import (
	"flag"
	"fmt"
	"testing"

	"github.com/sstarcher/go-okta"
)

var (
	oktaOrg      = flag.String("okta-org", "", "The organization against which security tokens are verified")
	oktaUser     = flag.String("okta-user", "", "The user name being used for okta testing")
	oktaPassword = flag.String("okta-password", "", "The password of the user being used for okta testing")
)

func TestOKTAAcceptance(t *testing.T) {
	client := okta.NewClient(*oktaOrg)
	client.Url = "oktapreview.com"

	ret, err := client.Authenticate(*oktaUser, *oktaPassword)
	if err != nil {
		t.Error("Expected nil, ", *oktaOrg, "got ", err.Error())
	}
	logger.Info(fmt.Sprintf("%+v", ret))
}

func TestOKTASession(t *testing.T) {
	client := okta.NewClient(*oktaOrg)
	client.Url = "oktapreview.com"

	ret, err := client.Authenticate(*oktaUser, *oktaPassword)
	if err != nil {
		t.Error("Expected nil, ", *oktaOrg, "got ", err.Error())
	}
	sess, err := client.Session(ret.SessionToken)
	if err != nil {
		t.Error("Expected nil, got ", err.Error())
	}
	logger.Info(fmt.Sprintf("%+v", sess))
}
