package main

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/karlmutch/errors"
)

func TestInfo(t *testing.T) {
	r, _ := Router([]errors.Error{})

	ts := httptest.NewServer(r)
	defer ts.Close()

	res, err := http.Get(ts.URL + "/health")
	if err != nil {
		t.Fatal(err)
	}
	if res.StatusCode != http.StatusOK {
		t.Errorf("unable to locate the /health resource %d", res.StatusCode)
	}

	res, err = http.Post(ts.URL+"/health", "text/plain", nil)
	if err != nil {
		t.Fatal(err)
	}
	if res.StatusCode != http.StatusMethodNotAllowed {
		t.Errorf("resource /health appears to allow PUTs unexpectedly %d", res.StatusCode)
	}

	res, err = http.Get(ts.URL + "/info")
	if err != nil {
		t.Fatal(err)
	}
	if res.StatusCode != http.StatusOK {
		t.Errorf("unable to locate the /info resource %d", res.StatusCode)
	}

	res, err = http.Post(ts.URL+"/info", "text/plain", nil)
	if err != nil {
		t.Fatal(err)
	}
	if res.StatusCode != http.StatusMethodNotAllowed {
		t.Errorf("resource /info appears to allow PUTs unexpectedly %d", res.StatusCode)
	}

	res, err = http.Get(ts.URL + "/not-exists")
	if err != nil {
		t.Fatal(err)
	}
	if res.StatusCode != http.StatusNotFound {
		t.Errorf("invalid resource did not fail a GET test")
	}
}
