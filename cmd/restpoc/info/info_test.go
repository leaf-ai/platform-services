package info

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-test/deep"

	"github.com/leaf-ai/platform-services/version"
)

func TestInfoContent(t *testing.T) {
	w := httptest.NewRecorder()
	Info(w, nil)

	resp := w.Result()
	if resp.StatusCode != http.StatusOK {
		t.Errorf("/info resource could not be retrieved %d", resp.StatusCode)
	}

	infoData, err := ioutil.ReadAll(resp.Body)
	resp.Body.Close()

	if err != nil {
		t.Fatal(err)
	}

	infoBlock := &info{}
	if err = json.Unmarshal(infoData, infoBlock); err != nil {
		t.Fatal(err)
	}

	infoRef := &info{
		GitHash: version.GitHash,
	}
	if diff := deep.Equal(infoBlock, infoRef); diff != nil {
		t.Error(diff)
	}
}
