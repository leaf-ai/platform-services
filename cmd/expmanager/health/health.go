package health

// This file implements a simple info resource for the
// REST server

import (
	"net/http"
)

func Health(w http.ResponseWriter, _ *http.Request) {
	w.WriteHeader(http.StatusOK)
}
