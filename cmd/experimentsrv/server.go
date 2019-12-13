package main

import (
	"context"
	"fmt"
	"net"
	"strings"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/health/grpc_health_v1"

	"google.golang.org/grpc/reflection"

	grpc_middleware "github.com/grpc-ecosystem/go-grpc-middleware"

	"github.com/go-stack/stack"
	"github.com/karlmutch/errors"

	model "github.com/leaf-ai/platform-services/internal/experiment"
	experiment "github.com/leaf-ai/platform-services/internal/gen/experimentsrv"

	"github.com/grpc-ecosystem/grpc-opentracing/go/otgrpc"
	opentracing "github.com/opentracing/opentracing-go"
	jaeger "github.com/uber/jaeger-client-go"
	"github.com/uber/jaeger-client-go/transport/zipkin"
)

type ExperimentServer struct {
}

func (*ExperimentServer) MeshCheck(ctx context.Context, in *experiment.CheckRequest) (resp *experiment.CheckResponse, err error) {

	resp = &experiment.CheckResponse{
		Modules: []string{},
	}

	ctxTimeout, cancel := context.WithTimeout(ctx, time.Duration(10*time.Second))
	defer cancel()

	if ds := aliveDownstream(ctxTimeout, in.Live); len(ds) != 0 {
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

	// Start the opentracing framework using Jaeger as the tracer implementation, and
	// zipkin HTTP backend interface pointing at the honeycomb opentracing proxy
	backendURI := "http://honeycomb-opentracing-proxy:9411/api/v1/spans"
	transport, errGo := zipkin.NewHTTPTransport(backendURI, zipkin.HTTPLogger(jaeger.StdLogger))
	if errGo != nil {
		logger.Warn(fmt.Sprint(errors.Wrap(errGo).With("stack", stack.Trace().TrimRuntime())))
	}

	reporter := jaeger.NewRemoteReporter(transport)
	sampler := jaeger.NewConstSampler(true) // Always output a trace when requested

	zstracer, zscloser := jaeger.NewTracer("experiment", sampler, reporter)
	opentracing.SetGlobalTracer(zstracer) // Setup the Jaeger as the global default
	defer zscloser.Close()

	streams := grpc_middleware.ChainStreamServer(
		authStreamInterceptor,
		otgrpc.OpenTracingStreamServerInterceptor(opentracing.GlobalTracer()),
	)
	unaries := grpc_middleware.ChainUnaryServer(
		authUnaryInterceptor,
		otgrpc.OpenTracingServerInterceptor(opentracing.GlobalTracer()),
	)

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
		<-ctx.Done()
		server.Stop()
	}()
	return errC
}
