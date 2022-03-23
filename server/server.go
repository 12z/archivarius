package server

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/12z/archivarius/arch"
)

const apiPrefix = "/api/v1"

type Server struct {
	server http.Server
}

// NewServer creates an instance of Server
func NewServer() *Server {
	mux := http.NewServeMux()

	mux.HandleFunc(fmt.Sprintf("%s/compress", apiPrefix), compressHandler)
	mux.HandleFunc(fmt.Sprintf("%s/extract", apiPrefix), extractHandler)

	server := &Server{
		server: http.Server{
			Handler: mux,
		},
	}

	return server
}

// Serve starts serving
func (s *Server) Serve() {
	// context?

	s.server.ListenAndServe()
}

func compressHandler(rw http.ResponseWriter, r *http.Request) {
	data, err := ioutil.ReadAll(r.Body)
	if err != nil {
		rw.WriteHeader(http.StatusInternalServerError)
		return
	}
	defer r.Body.Close()

	var compReq arch.CompressionRequest
	err = json.Unmarshal(data, &compReq)
	if err != nil {
		rw.WriteHeader(http.StatusBadRequest)
		return
	}

	err = arch.Compress(compReq)
	if err != nil {
		rw.WriteHeader(http.StatusInternalServerError)
	}
}

func extractHandler(rw http.ResponseWriter, r *http.Request) {
	data, err := ioutil.ReadAll(r.Body)
	if err != nil {
		rw.WriteHeader(http.StatusInternalServerError)
		return
	}
	defer r.Body.Close()

	var extReq arch.ExtractRequest
	err = json.Unmarshal(data, &extReq)
	if err != nil {
		rw.WriteHeader(http.StatusBadRequest)
		return
	}
	err = arch.Excract(extReq)
	if err != nil {
		rw.WriteHeader(http.StatusInternalServerError)
	}
}
