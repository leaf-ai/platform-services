package main

// This module checks the internal state of the server
// and will set the appropriate health values based on the checking

import (
	"context"
	"fmt"
	"sync"

	"google.golang.org/grpc/health"
	"google.golang.org/grpc/health/grpc_health_v1"
)

type healthCheck struct {
	h *health.Server
	sync.Mutex
}

var (
	serverHealth = &healthCheck{
		h: health.NewServer(),
	}
)

func grpcHealth(ctx context.Context, in *grpc_health_v1.HealthCheckRequest) (resp *grpc_health_v1.HealthCheckResponse, err error) {
	serverHealth.Lock()
	defer serverHealth.Unlock()
	return serverHealth.h.Check(ctx, in)
}

func initHealthTracker(ctx context.Context, serviceName string) {

	serverHealth.Lock()
	serverHealth.h.SetServingStatus(serviceName, grpc_health_v1.HealthCheckResponse_NOT_SERVING)
	serverHealth.Unlock()

	listenerC := make(chan bool)

	modules := &Modules{}
	modules.AddListener(listenerC)

	go func() {

		defer close(listenerC)

		defer func() {
			serverHealth.Lock()
			serverHealth.h.SetServingStatus(serviceName, grpc_health_v1.HealthCheckResponse_NOT_SERVING)
			serverHealth.Unlock()
		}()

		running := false
		lastKnown := running

		for {
			select {
			case up := <-listenerC:
				running = up
				serverHealth.Lock()
				if up {
					serverHealth.h.SetServingStatus(serviceName, grpc_health_v1.HealthCheckResponse_SERVING)
				} else {
					serverHealth.h.SetServingStatus(serviceName, grpc_health_v1.HealthCheckResponse_NOT_SERVING)
				}
				serverHealth.Unlock()
				if lastKnown != running {
					lastKnown = running
					logger.Info(fmt.Sprintf("server running state is %v", running))
				}
			case <-ctx.Done():
				return
			}
		}
	}()
}
