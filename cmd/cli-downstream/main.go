package main

// This code showns an example for accessing a gRPC service when a local, unsecured
// version of the service is present.

import (
	"flag"
	"fmt"
	"os"

	"github.com/leaf-ai/platform-services/internal/platform"

	"github.com/davecgh/go-spew/spew"
	"github.com/go-stack/stack"
	"github.com/karlmutch/errors"

	"golang.org/x/net/context"
	"google.golang.org/grpc"

	downstream "github.com/leaf-ai/platform-services/internal/gen/downstream"
)

var (
	logger = platform.NewLogger("cli-downstream")

	serverAddr = flag.String("server-addr", "127.0.0.1:80", "The server address in the format of host:port")
)

func main() {
	flag.Parse()
	opts := []grpc.DialOption{
		grpc.WithInsecure(),
	}
	conn, errGo := grpc.Dial(*serverAddr, opts...)
	if errGo != nil {
		logger.Fatal(fmt.Sprint(errors.Wrap(errGo).With("address", *serverAddr).With("stack", stack.Trace().TrimRuntime())))
	}
	defer conn.Close()
	client := downstream.NewDownstreamClient(conn)

	response, errGo := client.Ping(context.Background(), &downstream.PingRequest{})
	if errGo != nil {
		logger.Error(fmt.Sprint(errors.Wrap(errGo).With("address", *serverAddr).With("stack", stack.Trace().TrimRuntime())))
		os.Exit(-1)
	}
	spew.Fdump(os.Stdout, response)
}
