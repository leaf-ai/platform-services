package main

// This file contains the implementation of REST handlers
// for API requests made to the experiment manager

import (
	"github.com/gorilla/mux"

	"github.com/KarlMutch/MeshTest/cmd/expmanager/health"
	"github.com/KarlMutch/MeshTest/cmd/expmanager/info"
)

func Router() *mux.Router {
	r := mux.NewRouter()
	r.HandleFunc("/health", health.Health).Methods("GET")
	r.HandleFunc("/info", info.Info).Methods("GET")
	return r
}
