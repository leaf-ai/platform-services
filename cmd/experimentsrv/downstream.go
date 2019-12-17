package main

import (
	"context"
	"sync"
	"time"

	"github.com/go-stack/stack"
	"github.com/karlmutch/errors"

	"google.golang.org/grpc"

	opentracing "github.com/opentracing/opentracing-go"

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

	if parent := opentracing.SpanFromContext(ctx); parent != nil {
		pctx := parent.Context()
		if tracer := opentracing.GlobalTracer(); tracer != nil {
			span := tracer.StartSpan("dev.cognizant_ai.experiment.Service.aliveDownstream", opentracing.ChildOf(pctx))
			defer span.Finish()
			ctx = opentracing.ContextWithSpan(ctx, span)
		}
	}

	if onlineCheck {
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

	if parent := opentracing.SpanFromContext(ctx); parent != nil {
		pctx := parent.Context()
		if tracer := opentracing.GlobalTracer(); tracer != nil {
			span := tracer.StartSpan("dev.cognizant_ai.experiment.Service.checkDownstream", opentracing.ChildOf(pctx))
			defer span.Finish()
			ctx = opentracing.ContextWithSpan(ctx, span)
		}
	}

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

func initiateDownstream(ctx context.Context, hostAndPort string, refresh time.Duration) (err errors.Error) {
	seen.Lock()
	seen.hostAndPort = hostAndPort
	seen.Unlock()

	go func() {
		for {
			select {
			case <-time.After(refresh):
				func() {
					span := opentracing.StartSpan("dev.cognizant_ai.experiment.Service.initiateDownstream")
					defer span.Finish()

					timeout := refresh - time.Duration(-2*time.Second)
					ctxTimeout, cancel := context.WithTimeout(ctx, time.Duration(timeout))
					defer cancel()

					ctxTimeout = opentracing.ContextWithSpan(ctxTimeout, span)

					if err := seen.checkDownstream(ctxTimeout); err != nil {
						logger.Warn(err.Error())
					}
				}()
			case <-ctx.Done():
				return
			}
		}
	}()

	return nil
}
