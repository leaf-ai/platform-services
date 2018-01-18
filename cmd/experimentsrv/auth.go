package main

// This file contains the implementation of per request authentication
// for our gRPC server.  Authetication and validation are done using the Auth0
// platform.

import (
	"flag"
	"fmt"
	"strings"

	"net/http"

	"github.com/go-stack/stack"
	"github.com/karlmutch/errors"

	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"

	"github.com/auth0-community/go-auth0"
	"gopkg.in/square/go-jose.v2"
)

var (
	auth0Domain = flag.String("auth0-domain", "sentientai.auth0.com", "The domain assigned to the server API by Auth0")
)

func validateToken(token string, claimCheck string) (err errors.Error) {

	client := auth0.NewJWKClient(auth0.JWKClientOptions{
		URI: "https://" + *auth0Domain + "/.well-known/jwks.json",
	})
	audience := []string{
		"http://api.sentient.ai/experimentsrv",
	}

	configuration := auth0.NewConfiguration(client, audience, "https://"+*auth0Domain+"/", jose.RS256)
	validator := auth0.NewValidator(configuration)

	headerTokenRequest, errGo := http.NewRequest("", audience[0], nil)
	if errGo != nil {
		return errors.Wrap(errGo).With("stack", stack.Trace().TrimRuntime())
	}
	headerValue := fmt.Sprintf("Bearer %s", *auth0TestToken)
	headerTokenRequest.Header.Add("Authorization", headerValue)

	validResp, errGo := validator.ValidateRequest(headerTokenRequest)
	if errGo != nil {
		return errors.Wrap(errGo).With("stack", stack.Trace().TrimRuntime())
	}

	if len(claimCheck) == 0 {
		return nil
	}

	claims := map[string]interface{}{}
	errGo = validator.Claims(headerTokenRequest, validResp, &claims)
	if errGo != nil {
		return errors.Wrap(errGo).With("stack", stack.Trace().TrimRuntime())
	}

	if !strings.Contains(claims["scope"].(string), claimCheck) {
		return errors.New(fmt.Sprintf("the authenticated user does not have the appropriate '%s' scope", claimCheck)).With("stack", stack.Trace().TrimRuntime())
	}
	return nil
}

func authInterceptor(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
	meta, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return nil, grpc.Errorf(codes.Unauthenticated, "missing context metadata")
	}
	//if !authenticate(md["username"], md["password"]) {
	//	return nil, codes.Unauthenticated
	//}

	if len(meta["authorization"]) != 1 {
		return nil, grpc.Errorf(codes.Unauthenticated, "invalid token")
	}
	if err := validateToken(meta["authorization"][0], "all:experiments"); err != nil {
		return nil, grpc.Errorf(codes.Unauthenticated, err.Error())
	}

	return handler(ctx, req)
}
