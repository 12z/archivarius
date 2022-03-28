package server

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/12z/archivarius/arch"
)

const apiPrefix = "/api/v1"

type Server struct {
	server *http.Server
}

type Response struct {
	Status  string `json:"status"`
	Message string `json:"message,omitempty"`
}

// NewServer creates an instance of Server
func NewServer(srv *http.Server) *Server {
	mux := http.NewServeMux()

	mux.HandleFunc(fmt.Sprintf("%s/compress", apiPrefix), compressHandler)
	mux.HandleFunc(fmt.Sprintf("%s/extract", apiPrefix), extractHandler)

	srv.Handler = mux
	server := &Server{
		server: srv,
	}

	return server
}

func Router() *http.ServeMux {
	mux := http.NewServeMux()

	mux.HandleFunc(fmt.Sprintf("%s/compress", apiPrefix), compressHandler)
	mux.HandleFunc(fmt.Sprintf("%s/extract", apiPrefix), extractHandler)

	return mux
}

// Serve starts serving
func (s *Server) Serve() {
	// context?

	s.server.ListenAndServe()
}

func compressHandler(rw http.ResponseWriter, r *http.Request) {
	statusCode, resp := processCompress(r.Body)
	respData, err := json.Marshal(resp)
	if err != nil {
		rw.WriteHeader(http.StatusInternalServerError)
	}
	rw.WriteHeader(statusCode)
	rw.Write(respData)
}

func processCompress(r io.ReadCloser) (int, Response) {
	var statusCode = 200
	var resp Response

	data, err := io.ReadAll(r)
	if err != nil {
		statusCode = 500
		resp.Status = "nok"
		resp.Message = "unable to read request"
		return statusCode, resp
	}
	defer r.Close()

	var compReq arch.CompressionRequest
	err = json.Unmarshal(data, &compReq)
	if err != nil {
		statusCode = 400
		resp.Status = "nok"
		resp.Message = fmt.Sprintf("incorrect rquest format (%s)", err.Error())
		return statusCode, resp
	}

	stCode, err := arch.Compress(compReq)
	if err != nil {
		statusCode = stCode
		resp.Status = "nok"
		resp.Message = fmt.Sprintf("unable to compress (%s)", err.Error())
		return statusCode, resp
	}

	resp.Status = "ok"

	return statusCode, resp
}

func extractHandler(rw http.ResponseWriter, r *http.Request) {
	statusCode, resp := processExtract(r.Body)
	respData, err := json.Marshal(resp)
	if err != nil {
		rw.WriteHeader(http.StatusInternalServerError)
	}
	rw.WriteHeader(statusCode)
	rw.Write(respData)
}

func processExtract(r io.ReadCloser) (int, Response) {
	var statusCode = 200
	var resp Response

	data, err := io.ReadAll(r)
	if err != nil {
		statusCode = 500
		resp.Status = "nok"
		resp.Message = "unable to read request"
		return statusCode, resp
	}
	defer r.Close()

	var extReq arch.ExtractRequest
	err = json.Unmarshal(data, &extReq)
	if err != nil {
		statusCode = 400
		resp.Status = "nok"
		resp.Message = fmt.Sprintf("incorrect rquest format (%s)", err.Error())
		return statusCode, resp
	}

	stCode, err := arch.Excract(extReq)
	if err != nil {
		statusCode = stCode
		resp.Status = "nok"
		resp.Message = fmt.Sprintf("unable to compress (%s)", err.Error())
		return statusCode, resp
	}

	resp.Status = "ok"

	return statusCode, resp
}
