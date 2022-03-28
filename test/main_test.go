package test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"testing"

	"github.com/12z/archivarius/arch"
	"github.com/12z/archivarius/server"
)

func setupServer() *httptest.Server {
	mux := server.Router()
	testServer := httptest.NewServer(mux)
	return testServer
}

func TestBasic(t *testing.T) {
	srv := setupServer()
	setupTestBasicData(t)
	sUrl := srv.URL
	serverUrl, err := url.Parse(sUrl)
	if err != nil {
		t.Errorf("error while parse testserver url: %v", err)
		return
	}

	compReq := arch.CompressionRequest{
		ArchiveName: ".tmp/test1/archive.zip",
		Directory:   ".tmp/test1/src/",
	}

	compReqData, err := json.Marshal(compReq)
	if err != nil {
		t.Fatal(err)
	}

	serverUrl.Path = "/api/v1/compress"
	compResp, err := http.Post(serverUrl.String(), "application/json", bytes.NewReader(compReqData))
	if err != nil {
		t.Fatal(err)
	}
	if compResp.StatusCode != 200 {
		t.Fatalf("not 200 response %d", compResp.StatusCode)
	}

	extReq := arch.ExtractRequest{
		ArchiveName: ".tmp/test1/archive.zip",
		Directory:   ".tmp/test1/dst",
	}

	extReqData, err := json.Marshal(extReq)
	if err != nil {
		t.Fatal(err)
	}

	serverUrl.Path = "/api/v1/extract"
	extResp, err := http.Post(serverUrl.String(), "application/json", bytes.NewReader(extReqData))
	if err != nil {
		t.Fatal(err)
	}
	if extResp.StatusCode != 200 {
		t.Fatalf("not 200 response %d", extResp.StatusCode)
	}

	elems, err := os.ReadDir(".tmp/test1/dst")
	if err != nil {
		t.Fatal(err)
	}

	if len(elems) != 10 {
		t.Fatalf("expected 10 files, got %d", len(elems))
	}

	teardownTestBasicData(t)
}

func setupTestBasicData(t *testing.T) {
	if err := os.MkdirAll(".tmp/test1/src", os.ModePerm); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(".tmp/test1/dst", os.ModePerm); err != nil {
		t.Fatal(err)
	}

	os.WriteFile(".tmp/test1/src/one.txt", []byte{'1'}, os.ModePerm)
	os.WriteFile(".tmp/test1/src/two.txt", []byte{'1', '2'}, os.ModePerm)
	os.WriteFile(".tmp/test1/src/three.txt", []byte{'1', '2', '3'}, os.ModePerm)
	os.WriteFile(".tmp/test1/src/four.txt", []byte{'1', '2', '3', '4'}, os.ModePerm)
	os.WriteFile(".tmp/test1/src/five.txt", []byte{'1', '2', '3', '4', '5'}, os.ModePerm)
	os.WriteFile(".tmp/test1/src/six.txt", []byte{'1', '2', '3', '4', '5', '6'}, os.ModePerm)
	os.WriteFile(".tmp/test1/src/seven.txt", []byte{'1', '2', '3', '4', '5', '6', '7'}, os.ModePerm)
	os.WriteFile(".tmp/test1/src/eight.txt", []byte{'1', '2', '3', '4', '5', '6', '7', '8'}, os.ModePerm)
	os.WriteFile(".tmp/test1/src/nine.txt", []byte{'1', '2', '3', '4', '5', '6', '7', '8', '9'}, os.ModePerm)
	os.WriteFile(".tmp/test1/src/ten.txt", []byte{'1', '2', '3', '4', '5', '6', '7', '8', '9', '0'}, os.ModePerm)
	os.WriteFile(".tmp/test1/src/eleven.txt", []byte{'1', '2', '3', '4', '5', '6', '7', '8', '9', '0', '1'}, os.ModePerm)
	os.WriteFile(".tmp/test1/src/twelve.txt", []byte{'1', '2', '3', '4', '5', '6', '7', '8', '9', '0', '1', '2'}, os.ModePerm)

}

func teardownTestBasicData(t *testing.T) {
	if err := os.RemoveAll(".tmp/test1"); err != nil {
		t.Fatal(err)
	}
}
