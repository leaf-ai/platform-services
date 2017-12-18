package main

// This file contains the implementation of REST handlers
// for API requests made to the experiment manager

import (
	"github.com/gorilla/mux"
)

func Router() *mux.Router {
	r := mux.NewRouter()
	r.HandleFunc("/info", info).Methods("GET")
	return r
}
