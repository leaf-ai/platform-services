package main

import (
	"context"
	"flag"
	"fmt"
	"net"
	"strings"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/health/grpc_health_v1"

	"google.golang.org/grpc/reflection"

	grpc_middleware "github.com/grpc-ecosystem/go-grpc-middleware"

	"github.com/davecgh/go-spew/spew"
	"github.com/honeycombio/libhoney-go"

	"github.com/go-stack/stack"
	"github.com/karlmutch/errors"

	model "github.com/leaf-ai/platform-services/internal/experiment"
	experiment "github.com/leaf-ai/platform-services/internal/gen/experimentsrv"
	"github.com/leaf-ai/platform-services/internal/platform"
)

var (
	honeycombKey  = flag.String("o11y-key", "24cbba447cc8198c9c215d77ec159ed3", "An API key used to activate, and for use with the honeycomb.io service")
	honeycombData = flag.String("o11y-dataset", "platform-services", "The name for the dataset into which observability data is to be written")
)

type ExperimentServer struct {
}

func (*ExperimentServer) MeshCheck(ctx context.Context, in *experiment.CheckRequest) (resp *experiment.CheckResponse, err error) {

	resp = &experiment.CheckResponse{
		Modules: []string{},
	}

	if ds := aliveDownstream(); len(ds) != 0 {
		resp.Modules = append(resp.Modules, ds)
	}

	return resp, nil
}

func (*ExperimentServer) Create(ctx context.Context, in *experiment.CreateRequest) (resp *experiment.CreateResponse, err error) {
	if in == nil {
		return nil, fmt.Errorf("request is missing a message to experiment")
	}

	exp, err := model.InsertExperiment(ctx, in.Experiment)
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

	resp.Experiment, err = model.SelectExperimentWide(ctx, in.Uid)
	if err != nil {
		return nil, err
	}
	if resp.Experiment == nil {
		return nil, fmt.Errorf("no matching experiments found matching user specified input parameters")
	}

	return resp, nil
}

func (es *ExperimentServer) Check(ctx context.Context, in *grpc_health_v1.HealthCheckRequest) (resp *grpc_health_v1.HealthCheckResponse, err error) {
	return grpcHealth(ctx, in)
}

func (es *ExperimentServer) Watch(in *grpc_health_v1.HealthCheckRequest, server grpc_health_v1.Health_WatchServer) (err error) {
	return errors.New(grpc_health_v1.HealthCheckResponse_UNKNOWN.String())
}

func o11yUnaryInterceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, errGo error) {
		ev := platform.CreateEvent(ctx, "experiment")

		defer func() {
			if errEv := ev.Send(); errEv != nil {
				logger.Warn(fmt.Sprint(errors.Wrap(errEv).With("stack", stack.Trace().TrimRuntime())))
			}
		}()

		start := time.Now()

		// add fields to identify this event
		ev.Add(map[string]interface{}{
			"name":       info.FullMethod,
			"grpc.input": req,
		})

		handlerCtx := context.Context(ctx)
		for k, v := range ev.Fields() {
			handlerCtx = context.WithValue(handlerCtx, k, v)
		}
		if resp, errGo = handler(handlerCtx, req); errGo != nil {
			ev.AddField("grpc.error", errGo)
		}

		ev.AddField("duration_ms", float64(time.Since(start))/float64(time.Millisecond))
		logger.Debug(stack.Trace().TrimRuntime().String())
		logger.Debug(spew.Sdump(ev))
		logger.Debug(spew.Sdump(ctx))
		logger.Debug(spew.Sdump(handlerCtx))
		logger.Debug(spew.Sdump(req))

		return resp, errGo
	}
}

type stream struct {
	grpc.ServerStream
	ctx context.Context
}

func o11yStreamInterceptor() grpc.StreamServerInterceptor {
	return func(srv interface{}, strm grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) (errGo error) {
		ctx := strm.Context()
		ev := platform.CreateEvent(ctx, "experiment")
		defer func() {
			if errEv := ev.Send(); errEv != nil {
				logger.Warn(fmt.Sprint(errors.Wrap(errEv).With("stack", stack.Trace().TrimRuntime())))
			}
		}()

		start := time.Now()

		// Debug here
		logger.Debug(spew.Sdump(ctx) + stack.Trace().TrimRuntime().String())
		logger.Debug(spew.Sdump(ev) + stack.Trace().TrimRuntime().String())
		handlerCtx := context.Context(ctx)
		for k, v := range ev.Fields() {
			handlerCtx = context.WithValue(handlerCtx, k, v)
		}
		s := stream{strm, handlerCtx}

		ev.Add(map[string]interface{}{
			"name":        info.FullMethod,
			"duration_ms": float64(time.Since(start)) / float64(time.Millisecond),
		})

		if errGo = handler(srv, s); errGo != nil {
			ev.AddField("grpc.error", errGo)
		}

		return errGo
	}
}

func runServer(ctx context.Context, serviceName string, ipPort string) (errC chan errors.Error) {

	errC = make(chan errors.Error, 3)

	{
		if addrs, errGo := net.InterfaceAddrs(); errGo != nil {
			logger.Warn(fmt.Sprint(errors.Wrap(errGo).With("stack", stack.Trace().TrimRuntime())))
		} else {
			for _, addr := range addrs {
				logger.Debug("", "network", addr.Network(), "addr", addr.String())
			}
		}
	}

	// Start the honeycomb API
	libhoney.Init(libhoney.Config{
		WriteKey: *honeycombKey,
		Dataset:  *honeycombData,
	})

	streams := grpc_middleware.ChainStreamServer(
		authStreamInterceptor,
		o11yStreamInterceptor())
	unaries := grpc_middleware.ChainUnaryServer(
		authUnaryInterceptor,
		o11yUnaryInterceptor())

	server := grpc.NewServer(
		grpc.StreamInterceptor(streams),
		grpc.UnaryInterceptor(unaries),
	)

	experimentSrv := &ExperimentServer{}

	experiment.RegisterServiceServer(server, experimentSrv)
	grpc_health_v1.RegisterHealthServer(server, experimentSrv)

	reflection.Register(server)

	listeners := []net.Listener{}

	listenAddrs := strings.Split(ipPort, ",")
	for _, addr := range listenAddrs {
		// Check for strip off the port number which MUST be present, if found not to be present fail out
		ipPort := strings.Split(addr, ":")
		if len(ipPort) == 1 {
			errC <- errors.New("user specified address did not have a port (xx:NNN)").With("stack", stack.Trace().TrimRuntime()).With("ip-port", ipPort)
			return
		}
		// Look for the port as the last token
		ip := strings.Join(ipPort[:len(ipPort)-1], ":")
		//if err := net.ParseIP(ip); err == nil {
		//		errC <- errors.New("user specified address did not contain a valid IP (XXX:nn)").With("stack", stack.Trace().TrimRuntime()).With("ip-port", ip)
		//		return
		//	}
		proto := "tcp4"
		if strings.Contains(ip, ":") || len(ip) == 0 {
			proto = "tcp6"
		}
		netListen, err := net.Listen(proto, addr)
		if err != nil {
			errC <- errors.Wrap(err).With("stack", stack.Trace().TrimRuntime()).With("ip-port", addr)
			return
		}
		listeners = append(listeners, netListen)
		logger.Debug("", "addr", netListen.Addr())
	}

	for _, listener := range listeners {
		l := listener
		go func(netListen net.Listener, module string) {
			modules := &Modules{}
			modules.SetModule(module, true)

			if err := server.Serve(netListen); err != nil {
				errC <- errors.Wrap(err).With("stack", stack.Trace().TrimRuntime())
			}
			modules.SetModule(module, false)
			func() {
				defer recover()
				close(errC)
			}()
		}(l, l.Addr().String())
	}

	go func() {
		select {
		case <-ctx.Done():
			server.Stop()
		}
	}()
	return errC
}
