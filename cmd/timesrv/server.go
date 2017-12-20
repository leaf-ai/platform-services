package main

import (
	"context"
	"fmt"
	"time"

	"github.com/go-stack/stack"
	"github.com/karlmutch/errors"

	"github.com/go-openapi/loads"
	"github.com/go-openapi/runtime/middleware"
	"github.com/go-openapi/swag"

	"github.com/karlmutch/MeshTest/gen/restapi"
	"github.com/karlmutch/MeshTest/gen/restapi/operations"
)

func runServer(ctx context.Context, port int) (errC chan errors.Error) {

	errC = make(chan errors.Error, 3)

	// load embedded swagger file
	swaggerSpec, err := loads.Analyzed(restapi.SwaggerJSON, "")
	if err != nil {
		errC <- errors.Wrap(err).With("stack", stack.Trace().TrimRuntime())
		return
	}

	// create new service API
	api := operations.NewTimesrvAPI(swaggerSpec)
	server := restapi.NewServer(api)

	server.Port = port

	api.GetTimeHandler = operations.GetTimeHandlerFunc(
		func(params operations.GetTimeParams) middleware.Responder {
			tz := swag.StringValue(params.Timezone)
			if tz == "" {
				tz = "UTC"
			}

			theTime := fmt.Sprintf("%s", time.Now())
			return operations.NewGetTimeOK().WithPayload(theTime)
		})

	go func() {
		// serve API
		if err := server.Serve(); err != nil {
			errC <- errors.Wrap(err).With("stack", stack.Trace().TrimRuntime())
		}
		server.Shutdown()
		func() {
			defer recover()
			close(errC)
		}()
	}()
	return errC
}
