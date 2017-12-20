package main

// This file contains the implementation of REST handlers
// for API requests made to the experiment manager

import (
	"flag"
	"fmt"
	"time"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"github.com/go-chi/docgen"
	"github.com/go-chi/render"

	"github.com/karlmutch/MeshTest/cmd/expmanager/health"
	"github.com/karlmutch/MeshTest/cmd/expmanager/info"

	"github.com/karlmutch/errors"
)

var (
	routes = flag.Bool("routes", false, "Generate router documentation")
)

func Router(errI []errors.Error) (r chi.Router, errO []errors.Error) {

	r = chi.NewRouter()

	// A good base middleware stack
	r.Use(middleware.RequestID)       // Injects a request ID into the context of each request
	r.Use(middleware.DefaultCompress) // Gzip compression for clients that accept compressed responses
	r.Use(middleware.RealIP)          // Sets a http.Request's RemoteAddr to either X-Forwarded-For or X-Real-IP
	r.Use(middleware.Logger)          // Logs the start and end of each request with the elapsed processing time
	r.Use(middleware.Recoverer)       // Gracefully absorb panics and prints the stack trace

	r.Use(render.SetContentType(render.ContentTypeJSON))

	// Examples of other middleware that should be eventually placed here includes
	// Throttle, and Heartbeat

	// Set a timeout value on the request context (ctx), that will signal
	// through ctx.Done() that the request has timed out and further
	// processing should be stopped.
	r.Use(middleware.Timeout(60 * time.Second))

	r.Get("/health", health.Health)
	r.Get("/info", info.Info)

	// Passing -routes to the program will generate docs for the above
	// router definition. See the `routes.json` file in this folder for
	// the output.
	if *routes {
		// fmt.Println(docgen.JSONRoutesDoc(r))
		fmt.Println(docgen.MarkdownRoutesDoc(r, docgen.MarkdownOpts{
			ProjectPath: "github.com/karlMutch/MeshTest",
			Intro:       "Welcome to the experiment manager generated docs.",
		}))
		return nil, errI
	}

	return r, errI
}
