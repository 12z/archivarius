package test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/12z/archivarius/arch"
	"github.com/12z/archivarius/server"
)

type fileDef struct {
	name string
	data []byte
}

var files = map[int]fileDef{
	1:  {".tmp/test/src/one.txt", []byte("1")},
	2:  {".tmp/test/src/two.txt", []byte("12")},
	3:  {".tmp/test/src/three.txt", []byte("123")},
	4:  {".tmp/test/src/four.txt", []byte("1234")},
	5:  {".tmp/test/src/five.txt", []byte("12345")},
	6:  {".tmp/test/src/six.txt", []byte("123456")},
	7:  {".tmp/test/src/seven.txt", []byte("1234567")},
	8:  {".tmp/test/src/eight.txt", []byte("12345678")},
	9:  {".tmp/test/src/nine.txt", []byte("123456789")},
	10: {".tmp/test/src/ten.txt", []byte("1234567890")},
	11: {".tmp/test/src/eleven.txt", []byte("12345678901")},
	12: {".tmp/test/src/twelve.txt", []byte("123456789012")},

	21: {".tmp/test/src/uno.json", []byte(`["blue", "green"]`)},

	101: {".tmp/test/src/inner/inner1.txt", []byte("inner")},
	102: {".tmp/test/src/inner/inner/inner2.txt", []byte("double inner")},
}

func TestBasic(t *testing.T) {
	tests := []struct {
		name string

		compFilter string
		extrFilter string

		compLimit int
		extrLimit int

		files    []int
		expFiles []int
	}{
		{
			name:     "basic",
			files:    []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12},
			expFiles: []int{3, 4, 5, 6, 7, 8, 9, 10, 11, 12},
		},
		{
			name:     "two files",
			files:    []int{1, 4},
			expFiles: []int{1, 4},
		},
		{
			name:     "10 files",
			files:    []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10},
			expFiles: []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10},
		},
		{
			name:     "no files",
			files:    []int{},
			expFiles: []int{},
		},
		{
			name:     "subdirectories",
			files:    []int{5, 8, 101, 102},
			expFiles: []int{5, 8},
		},
		{
			name:       "compress filter",
			compFilter: "*.txt",
			files:      []int{1, 2, 3, 21},
			expFiles:   []int{1, 2, 3},
		},
		{
			name:       "extract filter",
			extrFilter: "*.txt",
			files:      []int{4, 5, 6, 21},
			expFiles:   []int{4, 5, 6},
		},
		{
			name:      "compress limit",
			compLimit: 3,
			files:     []int{5, 6, 7, 8},
			expFiles:  []int{6, 7, 8},
		},
		{
			name:      "extract limit",
			extrLimit: 3,
			files:     []int{5, 6, 7, 8, 9},
			expFiles:  []int{7, 8, 9},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			srv := setupServer()
			setupTestBasicData(t, tt.files)
			defer teardownTestBasicData(t)
			sUrl := srv.URL
			serverUrl, err := url.Parse(sUrl)
			if err != nil {
				t.Errorf("error while parse testserver url: %v", err)
				return
			}

			compReq := arch.Request{
				ArchiveName: ".tmp/test/archive.zip",
				Directory:   ".tmp/test/src/",
				Filter:      tt.compFilter,
				Limit:       tt.compLimit,
			}

			compReqData, err := json.Marshal(compReq)
			if err != nil {
				t.Fatal(err)
			}

			serverUrl.Path = "/api/v1/compress"
			compResp, err := http.Post(
				serverUrl.String(), "application/json", bytes.NewReader(compReqData))
			if err != nil {
				t.Fatal(err)
			}
			if compResp.StatusCode != 200 {
				t.Fatalf("not 200 response %d", compResp.StatusCode)
			}

			extReq := arch.Request{
				ArchiveName: ".tmp/test/archive.zip",
				Directory:   ".tmp/test/dst",
				Filter:      tt.extrFilter,
				Limit:       tt.extrLimit,
			}

			extReqData, err := json.Marshal(extReq)
			if err != nil {
				t.Fatal(err)
			}

			serverUrl.Path = "/api/v1/extract"
			extResp, err := http.Post(
				serverUrl.String(), "application/json", bytes.NewReader(extReqData))
			if err != nil {
				t.Fatal(err)
			}
			if extResp.StatusCode != 200 {
				t.Fatalf("not 200 response %d", extResp.StatusCode)
			}

			elems, err := os.ReadDir(".tmp/test/dst")
			if err != nil {
				t.Fatal(err)
			}

			if len(tt.expFiles) != len(elems) {
				t.Fatalf("expected %d files, got %d", len(tt.expFiles), len(elems))
			}

			expFilenames := func() []string {
				names := make([]string, 0, len(tt.expFiles))
				for _, i := range tt.expFiles {
					name := filepath.Base(files[i].name)
					names = append(names, name)
				}
				return names
			}()
			actFilenames := func() []string {
				names := make([]string, 0, len(elems))
				for _, elem := range elems {
					names = append(names, elem.Name())
				}
				return names
			}()
			if !fileListsEqual(expFilenames, actFilenames) {
				t.Fatalf("wrong files restored: expected %s, got %s", expFilenames, actFilenames)
			}
		})

		t.Run(fmt.Sprintf("async %s", tt.name), func(t *testing.T) {
			sleepInterval := time.Millisecond * 10
			srv := setupServer()
			setupTestBasicData(t, tt.files)
			defer teardownTestBasicData(t)
			sUrl := srv.URL
			serverUrl, err := url.Parse(sUrl)
			if err != nil {
				t.Errorf("error while parse testserver url: %v", err)
				return
			}

			compReq := arch.Request{
				ArchiveName: ".tmp/test/archive.zip",
				Directory:   ".tmp/test/src/",
				Filter:      tt.compFilter,
				Limit:       tt.compLimit,
			}

			compReqData, err := json.Marshal(compReq)
			if err != nil {
				t.Fatal(err)
			}

			serverUrl.Path = "/api/v1/compress/async"
			compResp, err := http.Post(
				serverUrl.String(), "application/json", bytes.NewReader(compReqData))
			if err != nil {
				t.Fatal(err)
			}
			if compResp.StatusCode != 200 {
				t.Fatalf("not 200 response %d", compResp.StatusCode)
			}
			pRespData, err := io.ReadAll(compResp.Body)
			if err != nil {
				t.Fatal(err)
			}
			var pResp server.AsyncPostResponse
			err = json.Unmarshal(pRespData, &pResp)
			if err != nil {
				t.Fatal(err)
			}
			sessionId := pResp.SessionId

			q := serverUrl.Query()
			q.Add("session_id", sessionId)
			serverUrl.RawQuery = q.Encode()

			time.Sleep(sleepInterval)
			compGetResp, err := http.Get(serverUrl.String())
			if err != nil {
				t.Fatal(err)
			}
			if compGetResp.StatusCode != 200 {
				t.Fatalf("not 200 response %d", compGetResp.StatusCode)
			}
			gRespData, err := io.ReadAll(compGetResp.Body)
			if err != nil {
				t.Fatal(err)
			}
			var gResp server.AsyncGetResponse
			err = json.Unmarshal(gRespData, &gResp)
			if err != nil {
				t.Fatal(err)
			}
			if gResp.Status != server.Finished {
				t.Fatalf("session not finished: status %s", gResp.Status)
			}
			if gResp.Result.Code != 200 {
				t.Fatalf("session result code is not 200, %d", gResp.Result.Code)
			}

			extReq := arch.Request{
				ArchiveName: ".tmp/test/archive.zip",
				Directory:   ".tmp/test/dst",
				Filter:      tt.extrFilter,
				Limit:       tt.extrLimit,
			}

			extReqData, err := json.Marshal(extReq)
			if err != nil {
				t.Fatal(err)
			}

			serverUrl.Path = "/api/v1/extract/async"
			serverUrl.RawQuery = ""
			extResp, err := http.Post(
				serverUrl.String(), "application/json", bytes.NewReader(extReqData))
			if err != nil {
				t.Fatal(err)
			}
			if extResp.StatusCode != 200 {
				t.Fatalf("not 200 response %d", extResp.StatusCode)
			}

			pRespData, err = io.ReadAll(extResp.Body)
			if err != nil {
				t.Fatal(err)
			}
			err = json.Unmarshal(pRespData, &pResp)
			if err != nil {
				t.Fatal(err)
			}
			sessionId = pResp.SessionId

			q = serverUrl.Query()
			q.Add("session_id", sessionId)
			serverUrl.RawQuery = q.Encode()

			time.Sleep(sleepInterval)
			getResp, err := http.Get(serverUrl.String())
			if err != nil {
				t.Fatal(err)
			}
			if compGetResp.StatusCode != 200 {
				t.Fatalf("not 200 response %d", compGetResp.StatusCode)
			}
			gRespData, err = io.ReadAll(getResp.Body)
			if err != nil {
				t.Fatal(err)
			}
			err = json.Unmarshal(gRespData, &gResp)
			if err != nil {
				t.Fatal(err)
			}
			if gResp.Status != server.Finished {
				t.Fatalf("session not finished: status %s", gResp.Status)
			}
			if gResp.Result.Code != 200 {
				t.Fatalf("session result code is not 200, %d", gResp.Result.Code)
			}

			elems, err := os.ReadDir(".tmp/test/dst")
			if err != nil {
				t.Fatal(err)
			}

			if len(tt.expFiles) != len(elems) {
				t.Fatalf("expected %d files, got %d", len(tt.expFiles), len(elems))
			}

			expFilenames := func() []string {
				names := make([]string, 0, len(tt.expFiles))
				for _, i := range tt.expFiles {
					name := filepath.Base(files[i].name)
					names = append(names, name)
				}
				return names
			}()
			actFilenames := func() []string {
				names := make([]string, 0, len(elems))
				for _, elem := range elems {
					names = append(names, elem.Name())
				}
				return names
			}()
			if !fileListsEqual(expFilenames, actFilenames) {
				t.Fatalf("wrong files restored: expected %s, got %s", expFilenames, actFilenames)
			}
		})

	}

}

func TestCompressNegative(t *testing.T) {
	tests := []struct {
		name    string
		req     arch.Request
		expCode int
		expResp []byte
	}{
		{
			name: "nonexistent source dir",
			req: arch.Request{
				ArchiveName: ".tmp/test/archive.zip",
				Directory:   ".tmp/test2/src/",
			},
			expCode: 400,
		},
		{
			name: "no rights to create archive",
			req: arch.Request{
				ArchiveName: "/archive.zip",
				Directory:   ".tmp/test2/src/",
			},
			expCode: 400,
		},
		{
			name:    "empty data in request",
			expCode: 400,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			srv := setupServer()
			setupTestBasicData(t, []int{})
			defer teardownTestBasicData(t)
			sUrl := srv.URL
			serverUrl, err := url.Parse(sUrl)
			if err != nil {
				t.Errorf("error while parse testserver url: %v", err)
				return
			}

			compReqData, err := json.Marshal(tt.req)
			if err != nil {
				t.Fatal(err)
			}

			serverUrl.Path = "/api/v1/compress"
			compResp, err := http.Post(
				serverUrl.String(), "application/json", bytes.NewReader(compReqData))
			if err != nil {
				t.Fatal(err)
			}
			if tt.expCode != compResp.StatusCode {
				t.Fatalf("Response code expected %d got %d", tt.expCode, compResp.StatusCode)
			}
		})
	}
	t.Run("illegal compression request", func(t *testing.T) {
		srv := setupServer()
		setupTestBasicData(t, []int{})
		defer teardownTestBasicData(t)
		sUrl := srv.URL
		serverUrl, err := url.Parse(sUrl)
		if err != nil {
			t.Errorf("error while parse testserver url: %v", err)
			return
		}

		req := struct {
			Blah string `json:"blah"`
		}{
			Blah: "data",
		}

		compReqData, err := json.Marshal(req)
		if err != nil {
			t.Fatal(err)
		}

		serverUrl.Path = "/api/v1/compress"
		compResp, err := http.Post(serverUrl.String(), "application/json", bytes.NewReader(compReqData))
		if err != nil {
			t.Fatal(err)
		}
		if compResp.StatusCode != 400 {
			t.Fatalf("Response code expected %d got %d", 400, compResp.StatusCode)
		}
	})
	t.Run("non POST compression request", func(t *testing.T) {
		srv := setupServer()
		setupTestBasicData(t, []int{})
		defer teardownTestBasicData(t)
		sUrl := srv.URL
		serverUrl, err := url.Parse(sUrl)
		if err != nil {
			t.Errorf("error while parse testserver url: %v", err)
			return
		}

		serverUrl.Path = "/api/v1/compress"
		compResp, err := http.Get(serverUrl.String())
		if err != nil {
			t.Fatal(err)
		}
		if compResp.StatusCode != 405 {
			t.Fatalf("Response code expected %d got %d", 405, compResp.StatusCode)
		}
	})
}

func TestExtractNegative(t *testing.T) {
	tests := []struct {
		name    string
		req     arch.Request
		expCode int
		expResp []byte
	}{
		{
			name: "nonexistent archive",
			req: arch.Request{
				ArchiveName: ".tmp/test2/archive.zip",
				Directory:   ".tmp/test/dst",
			},
			expCode: 400,
		},
		{
			name: "no rights to create archive",
			req: arch.Request{
				ArchiveName: ".tmp/test/archive.zip",
				Directory:   "/",
			},
			expCode: 400,
		},
		{
			name:    "empty data in request",
			expCode: 400,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			srv := setupServer()
			setupTestBasicData(t, []int{1, 3, 5, 7})
			defer teardownTestBasicData(t)
			sUrl := srv.URL
			serverUrl, err := url.Parse(sUrl)
			if err != nil {
				t.Errorf("error while parse testserver url: %v", err)
				return
			}

			compReq := arch.Request{
				ArchiveName: ".tmp/test/archive.zip",
				Directory:   ".tmp/test/src/",
			}

			compReqData, err := json.Marshal(compReq)
			if err != nil {
				t.Fatal(err)
			}

			serverUrl.Path = "/api/v1/compress"
			compResp, err := http.Post(
				serverUrl.String(), "application/json", bytes.NewReader(compReqData))
			if err != nil {
				t.Fatal(err)
			}
			if compResp.StatusCode != 200 {
				t.Fatalf("not 200 response %d", compResp.StatusCode)
			}

			extReqData, err := json.Marshal(tt.req)
			if err != nil {
				t.Fatal(err)
			}

			serverUrl.Path = "/api/v1/extract"
			extResp, err := http.Post(
				serverUrl.String(), "application/json", bytes.NewReader(extReqData))
			if err != nil {
				t.Fatal(err)
			}
			if tt.expCode != extResp.StatusCode {
				t.Fatalf("Response code expected %d response %d", tt.expCode, extResp.StatusCode)
			}
		})
	}
	t.Run("illegal extraction request", func(t *testing.T) {
		srv := setupServer()
		setupTestBasicData(t, []int{1, 3, 5, 7})
		defer teardownTestBasicData(t)
		sUrl := srv.URL
		serverUrl, err := url.Parse(sUrl)
		if err != nil {
			t.Errorf("error while parse testserver url: %v", err)
			return
		}

		compReq := arch.Request{
			ArchiveName: ".tmp/test/archive.zip",
			Directory:   ".tmp/test/src/",
		}

		compReqData, err := json.Marshal(compReq)
		if err != nil {
			t.Fatal(err)
		}

		serverUrl.Path = "/api/v1/compress"
		compResp, err := http.Post(
			serverUrl.String(), "application/json", bytes.NewReader(compReqData))
		if err != nil {
			t.Fatal(err)
		}
		if compResp.StatusCode != 200 {
			t.Fatalf("not 200 response %d", compResp.StatusCode)
		}

		req := struct {
			Blah string `json:"blah"`
		}{
			Blah: "data",
		}

		extReqData, err := json.Marshal(req)
		if err != nil {
			t.Fatal(err)
		}

		serverUrl.Path = "/api/v1/extract"
		extResp, err := http.Post(
			serverUrl.String(), "application/json", bytes.NewReader(extReqData))
		if err != nil {
			t.Fatal(err)
		}
		if extResp.StatusCode != 400 {
			t.Fatalf("Response code expected %d response %d", 400, extResp.StatusCode)
		}
	})
	t.Run("non POST extraction request", func(t *testing.T) {
		srv := setupServer()
		setupTestBasicData(t, []int{})
		defer teardownTestBasicData(t)
		sUrl := srv.URL
		serverUrl, err := url.Parse(sUrl)
		if err != nil {
			t.Errorf("error while parse testserver url: %v", err)
			return
		}

		serverUrl.Path = "/api/v1/extract"
		compResp, err := http.Get(serverUrl.String())
		if err != nil {
			t.Fatal(err)
		}
		if compResp.StatusCode != 405 {
			t.Fatalf("Response code expected %d got %d", 405, compResp.StatusCode)
		}
	})
}

func setupServer() *httptest.Server {
	sm := server.NewSessionManager()
	mux := server.Router(sm)
	testServer := httptest.NewServer(mux)
	return testServer
}

func setupTestBasicData(t *testing.T, filesInd []int) {
	if err := os.MkdirAll(".tmp/test/src", os.ModePerm); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(".tmp/test/dst", os.ModePerm); err != nil {
		t.Fatal(err)
	}

	for _, i := range filesInd {
		if err := os.MkdirAll(filepath.Dir(files[i].name), os.ModePerm); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(files[i].name, files[i].data, os.ModePerm); err != nil {
			t.Fatal(err)
		}
	}
}

func fileListsEqual(exp, act []string) bool {
	if len(exp) != len(act) {
		return false
	}
	diff := make(map[string]int, len(exp))
	for _, f := range exp {
		// increment for every file of expected
		diff[f]++
	}
	for _, f := range act {
		if _, ok := diff[f]; !ok {
			return false
		}
		// decrement for every file of actual
		diff[f] -= 1
		if diff[f] == 0 {
			delete(diff, f)
		}
	}
	return len(diff) == 0
}

func teardownTestBasicData(t *testing.T) {
	if err := os.RemoveAll(".tmp"); err != nil {
		t.Fatal(err)
	}
}
