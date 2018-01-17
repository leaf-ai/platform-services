package main

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/auth0-community/go-auth0"
	"gopkg.in/square/go-jose.v2"
)

var (
	auth0Domain    = flag.String("auth0-domain", "sentientai.auth0.com", "The domain assigned to the server API by Auth0")
	auth0TestToken = flag.String("auth0-token", "", "The raw token string recieved from an invocation of {auth0-domain}/oauth/token}")
)

func TestAuth0(t *testing.T) {
	client := auth0.NewJWKClient(auth0.JWKClientOptions{
		URI: "https://" + *auth0Domain + "/.well-known/jwks.json",
	})
	audience := []string{
		"http://api.sentient.ai/experimentsrv",
	}

	configuration := auth0.NewConfiguration(client, audience, "https://"+*auth0Domain+"/", jose.RS256)
	validator := auth0.NewValidator(configuration)

	headerTokenRequest, _ := http.NewRequest("", audience[0], nil)
	headerValue := fmt.Sprintf("Bearer %s", TOKEN_RAW)
	headerTokenRequest.Header.Add("Authorization", headerValue)

	token, errGo := validator.ValidateRequest(headerTokenRequest)
	if errGo != nil {
		t.Error("Expected nil, got ", errGo.Error())
	}
	logger.Info(fmt.Sprintf("%+v", token))
	claims := map[string]interface{}{}
	errGo = validator.Claims(headerTokenRequest, token, &claims)
	if errGo != nil {
		t.Error("Expected nil, got ", errGo.Error())
	}
	logger.Info(fmt.Sprintf("%+v", claims))
}
