package main

import (
	"context"
	"fmt"
	"net"
	"time"

	"github.com/go-stack/stack"
	"github.com/karlmutch/errors"

	"google.golang.org/grpc"
	"google.golang.org/grpc/health"
	"google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/reflection"

	timestamp "github.com/golang/protobuf/ptypes/timestamp"

	experiment "github.com/karlmutch/platform-services/gen/experimentsrv"
)

type experimentServer struct {
	health *health.Server
}

func (*experimentServer) Create(ctx context.Context, in *experiment.CreateRequest) (resp *experiment.CreateResponse, err error) {
	if in == nil {
		return nil, fmt.Errorf("request is missing a message to experiment")
	}

	return &experiment.CreateResponse{
			Id: "",
		},
		nil
}

func (*experimentServer) Get(ctx context.Context, in *experiment.GetRequest) (resp *experiment.GetResponse, err error) {
	if in == nil {
		return nil, fmt.Errorf("request is missing a message to experiment")
	}

	return &experiment.GetResponse{
			&experiment.Experiment{
				Name:        "",
				Description: "",
				CreateTime:  &timestamp.Timestamp{Seconds: time.Now().Unix()},
				InputLayer:  &experiment.InputLayer{},
				OutputLayer: &experiment.OutputLayer{},
			},
		},
		nil
}
func (es *experimentServer) Check(ctx context.Context, in *grpc_health_v1.HealthCheckRequest) (resp *grpc_health_v1.HealthCheckResponse, err error) {
	return es.health.Check(ctx, in)
}

func runServer(ctx context.Context, serviceName string, port int) (errC chan errors.Error) {

	errC = make(chan errors.Error, 3)

	server := grpc.NewServer()
	experimentSrv := &experimentServer{health: health.NewServer()}

	experiment.RegisterExperimentServiceServer(server, experimentSrv)
	grpc_health_v1.RegisterHealthServer(server, experimentSrv)

	reflection.Register(server)

	netListen, err := net.Listen("tcp", fmt.Sprintf("127.0.0.1:%d", port))
	if err != nil {
		errC <- errors.Wrap(err).With("stack", stack.Trace().TrimRuntime())
		return
	}

	go func() {
		experimentSrv.health.SetServingStatus(serviceName, grpc_health_v1.HealthCheckResponse_SERVING)

		// serve API
		if err := server.Serve(netListen); err != nil {
			errC <- errors.Wrap(err).With("stack", stack.Trace().TrimRuntime())
		}
		experimentSrv.health.SetServingStatus(serviceName, grpc_health_v1.HealthCheckResponse_NOT_SERVING)
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
