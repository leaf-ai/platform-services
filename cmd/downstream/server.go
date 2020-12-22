package main

import (
	"context"
	"flag"
	"fmt"
	"net"
	"strings"

	"github.com/go-stack/stack"
	"github.com/karlmutch/errors"

	"github.com/golang/protobuf/ptypes"

	"google.golang.org/grpc"
	"google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/reflection"

	downstream "github.com/leaf-ai/platform-services/internal/gen/downstream"
)

var (
	honeycombKey  = flag.String("o11y-key", "", "An API key used to activate, and for use with the honeycomb.io service")
	honeycombData = flag.String("o11y-dataset", "", "The name for the dataset into which observability data is to be written")
)

type DownstreamServer struct {
	downstream.UnimplementedDownstreamServer
}

func (*DownstreamServer) Ping(ctx context.Context, in *downstream.PingRequest) (resp *downstream.PingResponse, err error) {

	if in == nil {
		return nil, fmt.Errorf("request is missing a message to downstream")
	}

	resp = &downstream.PingResponse{
		Tm: ptypes.TimestampNow(),
	}

	return resp, nil
}

func (*DownstreamServer) Check(ctx context.Context, in *grpc_health_v1.HealthCheckRequest) (resp *grpc_health_v1.HealthCheckResponse, err error) {
	return grpcHealth(ctx, in)
}

func (*DownstreamServer) Watch(in *grpc_health_v1.HealthCheckRequest, server grpc_health_v1.Health_WatchServer) (err error) {
	return errors.New(grpc_health_v1.HealthCheckResponse_UNKNOWN.String())
}

func runServer(ctx context.Context, serviceName string, ipPort string) (errC chan errors.Error) {

	{
		if addrs, errGo := net.InterfaceAddrs(); errGo != nil {
			logger.Warn(fmt.Sprint(errors.Wrap(errGo).With("stack", stack.Trace().TrimRuntime())))
		} else {
			for _, addr := range addrs {
				logger.Debug("", "network", addr.Network(), "addr", addr.String())
			}
		}
	}

	// To prevent the server starting before the network listeners report
	// their states we inject a server module ID and set it to false then
	// one the logic to begin listening to the network interfaces is done
	// we set the server module ID is set to true (up) and this allows
	// the health check to visit the network listeners for their states
	//
	modules := &Modules{}
	serverModule := "serverInitDone"
	modules.SetModule(serverModule, false)
	defer modules.SetModule(serverModule, true)

	errC = make(chan errors.Error, 3)

	server := grpc.NewServer()
	handler := &DownstreamServer{}

	downstream.RegisterDownstreamServer(server, handler)
	grpc_health_v1.RegisterHealthServer(server, handler)

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
