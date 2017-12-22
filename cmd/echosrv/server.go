package main

import (
	"context"
	"fmt"
	"net"
	"time"

	"github.com/go-stack/stack"
	"github.com/karlmutch/errors"

	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"

	timestamp "github.com/golang/protobuf/ptypes/timestamp"

	echo "github.com/karlmutch/platform-services/gen/echosrv"
)

type echoServer struct {
}

func (*echoServer) Echo(ctx context.Context, in *echo.EchoRequest) (resp *echo.EchoResponse, err error) {
	if in == nil {
		return nil, fmt.Errorf("request is missing a message to echo")
	}

	return &echo.EchoResponse{Message: in.Message, DateTime: &timestamp.Timestamp{Seconds: time.Now().Unix()}}, nil
}

func runServer(ctx context.Context, port int) (errC chan errors.Error) {

	errC = make(chan errors.Error, 3)

	server := grpc.NewServer()
	echo.RegisterEchoServiceServer(server, &echoServer{})

	reflection.Register(server)

	netListen, err := net.Listen("tcp", fmt.Sprintf("127.0.0.1:%d", port))
	if err != nil {
		errC <- errors.Wrap(err).With("stack", stack.Trace().TrimRuntime())
		return
	}

	go func() {
		// serve API
		if err := server.Serve(netListen); err != nil {
			errC <- errors.Wrap(err).With("stack", stack.Trace().TrimRuntime())
		}
		func() {
			defer recover()
			close(errC)
		}()
	}()

	go func() {
		select {
		case <-ctx.Done():
			server.Stop()
		}
	}()
	return errC
}
