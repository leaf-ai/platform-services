package main

// This module checks the internal state of the server
// and will set the appropriate health values based on the checking

import (
	"context"
	"fmt"
	"sync"
	"time"

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

func initHealthTracker(serviceName string, quitC <-chan struct{}) {

	serverHealth.Lock()
	serverHealth.h.SetServingStatus(serviceName, grpc_health_v1.HealthCheckResponse_NOT_SERVING)
	serverHealth.Unlock()

	listenerC := make(chan bool)
	defer close(listenerC)

	modules := &Modules{}
	modules.AddListener(listenerC)

	go func() {

		defer func() {
			serverHealth.Lock()
			serverHealth.h.SetServingStatus(serviceName, grpc_health_v1.HealthCheckResponse_NOT_SERVING)
			serverHealth.Unlock()
		}()

		running := false

		for {
			select {
			case <-time.After(5 * time.Second):
				logger.Info(fmt.Sprintf("server running state is %v", running))
			case up := <-listenerC:
				running = up
				serverHealth.Lock()
				if up {
					serverHealth.h.SetServingStatus(serviceName, grpc_health_v1.HealthCheckResponse_SERVING)
				} else {
					serverHealth.h.SetServingStatus(serviceName, grpc_health_v1.HealthCheckResponse_NOT_SERVING)
				}
				serverHealth.Unlock()
			case <-quitC:
				return
			}
		}
	}()
}
