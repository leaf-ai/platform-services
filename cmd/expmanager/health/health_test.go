package health

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestHealthContent(t *testing.T) {
	w := httptest.NewRecorder()
	Health(w, nil)

	resp := w.Result()
	if resp.StatusCode != http.StatusOK {
		t.Errorf("/info resource could not be retrieved %d", resp.StatusCode)
	}
}
