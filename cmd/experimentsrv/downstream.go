package main

import (
	"context"
	"sync"
	"time"

	"github.com/davecgh/go-spew/spew"
	"github.com/go-stack/stack"
	"github.com/karlmutch/errors"
	"go.opencensus.io/trace"

	"google.golang.org/grpc"

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

func aliveDownstream(ctx context.Context, onlineCheck bool) (server string) {

	if onlineCheck {
		ctx, span := trace.StartSpan(ctx, "experiment/aliveDownstream",
			trace.WithSampler(trace.ProbabilitySampler(100.0)),
			trace.WithSpanKind(trace.SpanKindClient))
		defer span.End()
		if err := seen.checkDownstream(ctx); err == nil {
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

func (server *lastSeen) checkDownstream(ctx context.Context) (err errors.Error) {
	server.Lock()
	hostAndPort := server.hostAndPort
	server.Unlock()

	// Internal connections are protected using the mTLS features of the Istio side-car
	conn, errGo := grpc.Dial(hostAndPort, grpc.WithInsecure())
	if errGo != nil {
		return errors.Wrap(errGo).With("address", hostAndPort).With("stack", stack.Trace().TrimRuntime())
	}
	defer conn.Close()

	client := downstream.NewServiceClient(conn)

	ctx, span := trace.StartSpan(ctx, "experiment/checkDownstream",
		trace.WithSpanKind(trace.SpanKindClient))
	defer span.End()
	spew.Dump(ctx)

	if _, errGo = client.Ping(ctx, &downstream.PingRequest{}); errGo != nil {
		return errors.Wrap(errGo).With("address", hostAndPort).With("stack", stack.Trace().TrimRuntime())
	}
	server.Lock()
	server.when = time.Now()
	server.Unlock()
	return nil
}

func initiateDownstream(ctx context.Context, hostAndPort string, refresh time.Duration) (err errors.Error) {
	seen.Lock()
	seen.hostAndPort = hostAndPort
	seen.Unlock()

	go func() {
		for {
			select {
			case <-time.After(refresh):
				timeout := refresh - time.Duration(-2*time.Second)
				ctxTimeout, cancel := context.WithTimeout(ctx, time.Duration(timeout))

				if err := seen.checkDownstream(ctxTimeout); err != nil {
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
