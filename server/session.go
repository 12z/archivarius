package server

import (
	"fmt"
	"sync"

	"github.com/google/uuid"

	"github.com/12z/archivarius/arch"
)

const (
	Created  = "created"
	Started  = "started"
	Finished = "finished"
)

type Session struct {
	status string
	result AsyncResult
	mutex  sync.Mutex
}

func (s *Session) Run(req arch.Request, processor func(req arch.Request) (int, error)) {
	s.mutex.Lock()
	s.status = Started
	s.mutex.Unlock()

	var resp = Response{
		Status: "ok",
	}
	statusCode, err := processor(req)
	if err != nil {
		resp.Status = "nok"
		resp.Message = fmt.Sprintf("unable to process (%s)", err.Error())
	}

	s.mutex.Lock()
	defer s.mutex.Unlock()

	s.result.Code = statusCode
	s.result.Response = resp
	s.status = Finished
}

func (s *Session) Result() (string, AsyncResult) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	return s.status, s.result
}

type SessionManager struct {
	sessions map[string]*Session
	mutex    sync.Mutex
}

func NewSessionManager() *SessionManager {
	return &SessionManager{
		sessions: make(map[string]*Session),
	}
}

func (m *SessionManager) Get(id string) *Session {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	return m.sessions[id]
}

func (m *SessionManager) CreateSession() (string, *Session) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	id := uuid.New().String()
	session := &Session{
		status: Created,
	}
	m.sessions[id] = session

	return id, session
}

func (m *SessionManager) Delete(id string) {
	delete(m.sessions, id)
}
