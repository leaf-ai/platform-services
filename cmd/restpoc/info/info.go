package info

// This file implements a simple info resource for the
// REST server

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/leaf-ai/platform-services/internal/version"
)

type info struct {
	GitHash string `json:"VersionHash,omitempty"`
}

var (
	// On initialization will contain the version information as JSON
	imprint = ""
)

func init() {
	inf := &info{
		GitHash: version.GitHash,
	}
	b, err := json.Marshal(inf)
	if err != nil {
		return
	}
	imprint = string(b)
}

func Info(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	fmt.Fprint(w, imprint)
}
