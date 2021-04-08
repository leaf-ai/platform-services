package main

// This client can be used to access a fully deployed service mesh example using both
// a LetsEcrypt certificate to secure the connection and a JWT token to authenticate
// on top of the secured connection

import (
	"crypto/tls"
	"flag"
	"fmt"
	"os"

	"github.com/leaf-ai/platform-services/internal/platform"

	"github.com/davecgh/go-spew/spew"
	"github.com/go-stack/stack"
	"github.com/karlmutch/errors"

	"golang.org/x/net/context"
	"golang.org/x/oauth2"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/oauth"

	experimentsrv "github.com/leaf-ai/platform-services/internal/gen/experimentsrv"
)

var (
	logger = platform.NewLogger("cli-experiment")

	serverAddr = flag.String("server-addr", "127.0.0.1:443", "The server address in the format of host:port")
	optToken   = flag.String("auth0-token", "", "The authentication token to be used for the user, (JWT Token)")
)

func getToken() (token *oauth2.Token, errGo errors.Error) {
	return &oauth2.Token{
		AccessToken: *optToken,
	}, errGo
}

func main() {
	flag.Parse()

	// Obtain a JWT and prepare it for use with the gRPC connection and API calls
	token, err := getToken()
	if err != nil {
		logger.Fatal(fmt.Sprint(err.With("address", *serverAddr).With("stack", stack.Trace().TrimRuntime())))
	}
	perRPC := oauth.NewOauthAccess(token)

	opts := []grpc.DialOption{
		grpc.WithPerRPCCredentials(perRPC),
		grpc.WithTransportCredentials(credentials.NewTLS(&tls.Config{})),
	}
	opts = append(opts, grpc.WithBlock())

	conn, errGo := grpc.Dial(*serverAddr, opts...)
	if errGo != nil {
		logger.Fatal(fmt.Sprint(errors.Wrap(errGo).With("address", *serverAddr).With("stack", stack.Trace().TrimRuntime())))
	}
	defer conn.Close()
	client := experimentsrv.NewExperimentsClient(conn)

	response, errGo := client.MeshCheck(context.Background(), &experimentsrv.CheckRequest{Live: true})
	if errGo != nil {
		logger.Error(fmt.Sprint(errors.Wrap(errGo).With("address", *serverAddr).With("stack", stack.Trace().TrimRuntime())))
		os.Exit(-1)
	}
	spew.Fdump(os.Stdout, response)
}
