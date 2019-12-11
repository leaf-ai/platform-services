package main

// This module checks the internal state of the server
// and will set the appropriate health values based on the checking

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/davecgh/go-spew/spew"

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

	logger.Info(spew.Sdump(ctx))
	serverHealth.Lock()
	defer serverHealth.Unlock()
	return serverHealth.h.Check(ctx, in)
}

func initHealthTracker(serviceName string, quitC <-chan struct{}) {

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

		// This next section of code is used to setup reporting times for the server state in our logs.
		// The intervals are in seconds and a Fibonacci series to prevent logs filling up with rubbish.
		// The series position is reset on changes in state.
		timeUpdate := 0
		updateTimes := []int{1, 1, 2, 3, 5, 8, 13, 21, 34, 55, 89, 144, 233}
		nextReportTime := time.Now().Add(5 * time.Second)
		lastStateMsg := ""

		for {
			select {
			case <-time.After(time.Second):
				msg := fmt.Sprintf("server running state is %v", running)
				if time.Now().Before(nextReportTime) {
					if msg == lastStateMsg {
						continue
					}
				}
				// output the msg and then determine when it should next be shown
				logger.Info(msg)
				if msg == lastStateMsg {
					if timeUpdate+1 < len(updateTimes) {
						timeUpdate += 1
					}
				} else {
					lastStateMsg = msg
					timeUpdate = 0
				}
				nextReportTime = time.Now().Add(time.Duration(time.Duration(updateTimes[timeUpdate]) * time.Second))

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
