package main

import (
	"context"
	"fmt"
	"net"
	"strings"

	"github.com/go-stack/stack"
	"github.com/karlmutch/errors"

	"google.golang.org/grpc"
	"google.golang.org/grpc/health"
	"google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/reflection"

	model "github.com/SentientTechnologies/platform-services/experiment"
	experiment "github.com/SentientTechnologies/platform-services/gen/experimentsrv"
)

type ExperimentServer struct {
	health *health.Server
}

func (*ExperimentServer) Create(ctx context.Context, in *experiment.CreateRequest) (resp *experiment.CreateResponse, err error) {
	if in == nil {
		return nil, fmt.Errorf("request is missing a message to experiment")
	}

	exp, err := model.InsertExperiment(in.Experiment)
	if err != nil {
		return nil, err
	}

	resp = &experiment.CreateResponse{
		Uid: exp.Uid,
	}

	return resp, nil
}

func (*ExperimentServer) Get(ctx context.Context, in *experiment.GetRequest) (resp *experiment.GetResponse, err error) {
	if in == nil || len(strings.TrimSpace(in.Uid)) == 0 {
		return nil, fmt.Errorf("request is missing input parameters")
	}

	resp = &experiment.GetResponse{}

	resp.Experiment, err = model.SelectExperimentWide(in.Uid)
	if err != nil {
		return nil, err
	}
	if resp.Experiment == nil {
		return nil, fmt.Errorf("no matching experiments found matching user specified input parameters")
	}

	return resp, nil
}

func (es *ExperimentServer) Check(ctx context.Context, in *grpc_health_v1.HealthCheckRequest) (resp *grpc_health_v1.HealthCheckResponse, err error) {
	return es.health.Check(ctx, in)
}

func runServer(ctx context.Context, serviceName string, ipPort string) (errC chan errors.Error) {

	errC = make(chan errors.Error, 3)

	server := grpc.NewServer(grpc.UnaryInterceptor(authInterceptor))
	experimentSrv := &ExperimentServer{health: health.NewServer()}

	experiment.RegisterServiceServer(server, experimentSrv)
	grpc_health_v1.RegisterHealthServer(server, experimentSrv)

	reflection.Register(server)

	netListen, err := net.Listen("tcp", ipPort)
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
