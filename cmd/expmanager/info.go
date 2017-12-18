package main

// This file implements a simple info resource for the
// REST server

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/KarlMutch/MeshTest/version"
)

type Info struct {
	GitHash    string `json:"VersionHash,omitempty"`
	BuildStamp string `json:"BuildTimeStamp,omitempty"`
}

var (
	// On initialization will contain the version information as JSON
	imprint = ""
)

func init() {
	inf := &Info{
		GitHash:    version.GitHash,
		BuildStamp: version.BuildTime,
	}
	b, err := json.Marshal(inf)
	if err != nil {
		return
	}
	imprint = string(b)
}

func info(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	fmt.Fprint(w, imprint)
}
