package main

import (
	"context"
	"sync"
	"time"

	"github.com/go-stack/stack"
	"github.com/karlmutch/errors"

	"google.golang.org/grpc"

	downstream "github.com/leaf-ai/platform-services/internal/gen/downstream"
)

var (
	seen = lastSeen{}
)

type lastSeen struct {
	when time.Time
	sync.Mutex
}

func aliveDownstream() (server string) {
	seen.Lock()
	defer seen.Unlock()

	if seen.when.After(time.Now().Add(-15 * time.Second)) {
		return "downstream"
	}
	return ""
}

func checkDownstream(addr string) (err errors.Error) {
	conn, errGo := grpc.Dial(addr, grpc.WithInsecure())
	if errGo != nil {
		return errors.Wrap(errGo).With("address", addr).With("stack", stack.Trace().TrimRuntime())
	}
	defer conn.Close()

	client := downstream.NewServiceClient(conn)

	ctx, _ := context.WithTimeout(context.Background(), time.Duration(10*time.Second))
	if _, errGo = client.Ping(ctx, &downstream.PingRequest{}); errGo != nil {
		return errors.Wrap(errGo).With("address", addr).With("stack", stack.Trace().TrimRuntime())
	}
	return nil
}

func initiateDownstream(quitC <-chan struct{}) (err errors.Error) {
	go func() {
		internalCheck := time.Duration(time.Second)
		for {
			select {
			case <-time.After(internalCheck):
				if err := checkDownstream("downstream:30001"); err != nil {
					logger.Warn(err.Error())
					continue
				}
				seen.Lock()
				seen.when = time.Now()
				seen.Unlock()
			case <-quitC:
				return
			}
		}
	}()

	return nil
}
