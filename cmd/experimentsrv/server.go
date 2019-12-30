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
	"google.golang.org/grpc/metadata"

	"google.golang.org/grpc/reflection"

	"github.com/davecgh/go-spew/spew"
	"github.com/dgrijalva/jwt-go"
	grpc_middleware "github.com/grpc-ecosystem/go-grpc-middleware"
	zipkin "github.com/openzipkin/zipkin-go"
	zipkingrpc "github.com/openzipkin/zipkin-go/middleware/grpc"

	"github.com/go-stack/stack"
	"github.com/karlmutch/errors"

	model "github.com/leaf-ai/platform-services/internal/experiment"
	experiment "github.com/leaf-ai/platform-services/internal/gen/experimentsrv"
)

var (
	honeycombKey  = flag.String("o11y-key", "", "An API key used to activate, and for use with the honeycomb.io service")
	honeycombData = flag.String("o11y-dataset", "", "The name for the dataset into which observability data is to be written")
)

type ExperimentServer struct {
	tracer *zipkin.Tracer
}

// GetUserFromClaims show how given a custom rule in Auth0 extrat metadata related to the claims of
// what is assumed to be an authenticated user can be extracted and used, for example the users
// email address.  The custom rule would appear as follows:
//
// function (user, context, callback) {
//  context.accessToken["http://cognizant-ai.dev/user"] = user.email;
//  callback(null, user, context);
// }
//
func GetUserFromClaims(ctx context.Context) {
	md, _ := metadata.FromIncomingContext(ctx)

	user, err := func() (user string, err errors.Error) {
		if auth := md.Get("authorization"); len(auth) > 0 {
			splitToken := strings.Split(auth[0], "Bearer")
			if len(splitToken) != 2 {
				return "", errors.New("badly formatted token").With("stack", stack.Trace().TrimRuntime())
			}

			type CustomClaims struct {
				Email string `json:"http://cognizant-ai.dev/user"`
				jwt.StandardClaims
			}

			func() {
				defer func() {
					logger.Warn(spew.Sdump(recover()))
				}()
				token, _ := jwt.ParseWithClaims(strings.TrimSpace(splitToken[1]), &CustomClaims{}, func(token *jwt.Token) (interface{}, error) {
					if _, ok := token.Method.(*jwt.SigningMethodRSA); !ok {
						return nil, fmt.Errorf("Unexpected signing method: %v", token.Header["alg"])
					}
					// Dont supply a valuable key as we assume that because this is embeeded in a mesh that the mTLS has
					// secured the JWT in transit between the Auth module of the ingress and the service itself.
					// Should you wish to fully secure this then a private key should be supplied here so that the signing
					// can be checked at least.  THis however has risk and should be designed in by the final system
					// implementation people.
					return nil, nil
				})
				// Untrusted dely on the mTLS to secure it and prevent meddling
				if claims, ok := token.Claims.(*CustomClaims); ok {
					user = claims.Email
				}
			}()
		}
		return user, nil
	}()
	if err != nil {
		logger.Warn(err.Error())
	} else {
		if len(user) != 0 {
			logger.Info(user)
		}
	}
}

func (eServer *ExperimentServer) MeshCheck(ctx context.Context, in *experiment.CheckRequest) (resp *experiment.CheckResponse, err error) {

	GetUserFromClaims(ctx)

	resp = &experiment.CheckResponse{
		Modules: []string{},
	}

	ctxTimeout, cancel := context.WithTimeout(ctx, time.Duration(10*time.Second))
	defer cancel()

	if ds := aliveDownstream(ctxTimeout, eServer.tracer, in.Live); len(ds) != 0 {
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

func runServer(ctx context.Context, tracer *zipkin.Tracer, serviceName string, ipPort string) (errC chan errors.Error) {

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

	streams := grpc_middleware.ChainStreamServer(
		authStreamInterceptor,
	)
	unaries := grpc_middleware.ChainUnaryServer(
		authUnaryInterceptor,
	)

	// Set up the server with the OpenCensus
	// stats handler to enable stats and tracing
	server := grpc.NewServer(
		grpc.StatsHandler(zipkingrpc.NewServerHandler(tracer)),
		grpc.StreamInterceptor(streams),
		grpc.UnaryInterceptor(unaries),
	)

	experimentSrv := &ExperimentServer{tracer: tracer}

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
