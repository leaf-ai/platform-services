package main

import (
	"context"
	"sync"
	"time"

	"github.com/go-stack/stack"
	"github.com/karlmutch/errors"

	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"

	zipkin "github.com/openzipkin/zipkin-go"

	downstream "github.com/leaf-ai/platform-services/internal/gen/downstream"
)

var (
	seen = lastSeen{}
)

type lastSeen struct {
	hostAndPort string
	when        time.Time
	sync.Mutex
}

func aliveDownstream(ctx context.Context, tracer *zipkin.Tracer, onlineCheck bool) (server string) {

	if onlineCheck {
		if err := seen.checkDownstream(ctx, tracer); err == nil {
			return "downstream"
		}
		return ""
	}

	if seen.recentlySeen() {
		return "downstream"
	}
	return ""
}

func (server *lastSeen) recentlySeen() (recent bool) {
	server.Lock()
	defer server.Unlock()
	return server.when.After(time.Now().Add(-15 * time.Second))
}

func (server *lastSeen) checkDownstream(ctx context.Context, tracer *zipkin.Tracer) (err errors.Error) {

	span, ctx := tracer.StartSpanFromContext(ctx, "checkDownstream")
	defer span.Finish()

	server.Lock()
	hostAndPort := server.hostAndPort
	server.Unlock()

	md, _ := metadata.FromIncomingContext(ctx)
	ctx = metadata.NewOutgoingContext(ctx, md)

	// Internal connections are protected using the mTLS features of the Istio side-car
	conn, errGo := grpc.Dial(hostAndPort, grpc.WithInsecure())
	if errGo != nil {
		return errors.Wrap(errGo).With("address", hostAndPort).With("stack", stack.Trace().TrimRuntime())
	}
	defer conn.Close()

	client := downstream.NewServiceClient(conn)

	if _, errGo = client.Ping(ctx, &downstream.PingRequest{}); errGo != nil {
		return errors.Wrap(errGo).With("address", hostAndPort).With("stack", stack.Trace().TrimRuntime())
	}
	server.Lock()
	server.when = time.Now()
	server.Unlock()
	return nil
}

func initiateDownstream(ctx context.Context, tracer *zipkin.Tracer, hostAndPort string, refresh time.Duration) (err errors.Error) {
	seen.Lock()
	seen.hostAndPort = hostAndPort
	seen.Unlock()

	go func() {
		for {
			select {
			case <-time.After(refresh):
				timeout := refresh - time.Duration(-2*time.Second)
				ctxTimeout, cancel := context.WithTimeout(ctx, time.Duration(timeout))

				if err := seen.checkDownstream(ctxTimeout, tracer); err != nil {
					logger.Warn(err.Error())
				}
				cancel()
			case <-ctx.Done():
				return
			}
		}
	}()

	return nil
}
