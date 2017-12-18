package main

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-test/deep"

	"github.com/KarlMutch/MeshTest/version"
)

func TestInfo(t *testing.T) {
	r := Router()

	ts := httptest.NewServer(r)
	defer ts.Close()

	res, err := http.Get(ts.URL + "/info")
	if err != nil {
		t.Fatal(err)
	}
	if res.StatusCode != http.StatusOK {
		t.Errorf("unable to locater the /info resource %d", res.StatusCode)
	}

	res, err = http.Post(ts.URL+"/info", "text/plain", nil)
	if err != nil {
		t.Fatal(err)
	}
	if res.StatusCode != http.StatusMethodNotAllowed {
		t.Errorf("resource/info appears to allow PUTs unexpectedly %d", res.StatusCode)
	}

	res, err = http.Get(ts.URL + "/not-exists")
	if err != nil {
		t.Fatal(err)
	}
	if res.StatusCode != http.StatusNotFound {
		t.Errorf("invalid resource did not fail a GET test")
	}
}

func TestInfoContent(t *testing.T) {
	w := httptest.NewRecorder()
	info(w, nil)

	resp := w.Result()
	if resp.StatusCode != http.StatusOK {
		t.Errorf("/info resource could not be retrieved %d", resp.StatusCode)
	}

	infoData, err := ioutil.ReadAll(resp.Body)
	resp.Body.Close()

	if err != nil {
		t.Fatal(err)
	}

	infoBlock := &Info{}
	if err = json.Unmarshal(infoData, infoBlock); err != nil {
		t.Fatal(err)
	}

	infoRef := &Info{
		GitHash:    version.GitHash,
		BuildStamp: version.BuildTime,
	}
	if diff := deep.Equal(infoBlock, infoRef); diff != nil {
		t.Error(diff)
	}
}
