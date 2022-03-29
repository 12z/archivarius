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
	sm     *SessionManager
}

type Response struct {
	Status  string `json:"status"`
	Message string `json:"message,omitempty"`
}

type AsyncResult struct {
	Code     int      `json:"status_code,omitempty"`
	Response Response `json:"response,omitempty"`
}

type AsyncGetResponse struct {
	Status string      `json:"status"`
	Result AsyncResult `json:"result,omitempty"`
}

type AsyncPostResponse struct {
	SessionId string `json:"session_id"`
	Status    string `json:"status"`
	Message   string `json:"message,omitempty"`
}

// NewServer creates an instance of Server
func NewServer(srv *http.Server) *Server {
	sm := NewSessionManager()
	srv.Handler = Router(sm)
	server := &Server{
		server: srv,
		sm:     sm,
	}

	return server
}

func Router(sm *SessionManager) *http.ServeMux {
	mux := http.NewServeMux()

	mux.HandleFunc(fmt.Sprintf("%s/compress", apiPrefix), compressHandler)
	mux.HandleFunc(fmt.Sprintf("%s/extract", apiPrefix), extractHandler)
	mux.HandleFunc(fmt.Sprintf("%s/compress/async", apiPrefix),
		func(w http.ResponseWriter, r *http.Request) {
			compressHandlerAsync(w, r, sm)
		})
	mux.HandleFunc(fmt.Sprintf("%s/extract/async", apiPrefix),
		func(w http.ResponseWriter, r *http.Request) {
			extractHandlerAsync(w, r, sm)
		})

	return mux
}

// Serve starts serving
func (s *Server) Serve() {
	s.server.ListenAndServe()
}

func compressHandler(rw http.ResponseWriter, r *http.Request) {
	syncHandler(rw, r, arch.Compress)
}

func extractHandler(rw http.ResponseWriter, r *http.Request) {
	syncHandler(rw, r, arch.Extract)
}

func syncHandler(rw http.ResponseWriter, r *http.Request,
	processor func(req arch.Request) (int, error),
) {
	if r.Method != "POST" {
		rw.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	statusCode, resp := processSync(r.Body, processor)
	respData, err := json.Marshal(resp)
	if err != nil {
		rw.WriteHeader(http.StatusInternalServerError)
		return
	}
	rw.WriteHeader(statusCode)
	rw.Write(respData)
}

func processSync(r io.ReadCloser, processor func(req arch.Request) (int, error)) (int, Response) {
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

	var req arch.Request
	err = json.Unmarshal(data, &req)
	if err != nil {
		statusCode = 400
		resp.Status = "nok"
		resp.Message = fmt.Sprintf("incorrect rquest format (%s)", err.Error())
		return statusCode, resp
	}

	stCode, err := processor(req)
	if err != nil {
		statusCode = stCode
		resp.Status = "nok"
		resp.Message = fmt.Sprintf("unable to process (%s)", err.Error())
		return statusCode, resp
	}

	resp.Status = "ok"

	return statusCode, resp
}

func compressHandlerAsync(rw http.ResponseWriter, r *http.Request, sm *SessionManager) {
	handlerAsync(rw, r, sm, arch.Compress)
}

func extractHandlerAsync(rw http.ResponseWriter, r *http.Request, sm *SessionManager) {
	handlerAsync(rw, r, sm, arch.Extract)
}

func handlerAsync(rw http.ResponseWriter, r *http.Request, sm *SessionManager,
	processor func(req arch.Request) (int, error),
) {
	switch r.Method {
	case "POST":
		session_id, session := sm.CreateSession()
		statusCode, resp := processAsync(r.Body, session, processor)

		pResp := AsyncPostResponse{session_id, resp.Status, resp.Message}
		respData, err := json.Marshal(pResp)
		if err != nil {
			rw.WriteHeader(http.StatusInternalServerError)
			return
		}
		rw.WriteHeader(statusCode)
		rw.Write(respData)

	case "GET":
		sessionId := r.URL.Query().Get("session_id")
		session := sm.Get(sessionId)
		if session == nil {
			rw.WriteHeader(http.StatusNotFound)
			return
		}
		status, res := session.Result()
		resp := AsyncGetResponse{status, res}
		respData, err := json.Marshal(resp)
		if err != nil {
			rw.WriteHeader(http.StatusInternalServerError)
			return
		}
		rw.Write(respData)

	case "DELETE":
		sessionId := r.URL.Query().Get("session_id")
		sm.Delete(sessionId)

	default:
		rw.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
}

func processAsync(r io.ReadCloser, session *Session,
	processor func(req arch.Request) (int, error),
) (int, Response) {
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

	var req arch.Request
	err = json.Unmarshal(data, &req)
	if err != nil {
		statusCode = 400
		resp.Status = "nok"
		resp.Message = fmt.Sprintf("incorrect rquest format (%s)", err.Error())
		return statusCode, resp
	}

	go session.Run(req, processor)

	resp.Status = "ok"

	return statusCode, resp
}
