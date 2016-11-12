package soso

import "sync"

var Sessions = NewSessionList()

type Session interface {
	// Id returns a session id
	ID() string
	// Recv reads one text frame from session
	Recv() ([]byte, error)
	// Send sends one text frame to session
	Send(string) error
	// Close closes the session with provided code and reason.
	Close(status uint32, reason string) error
	IsClosed() bool
}

type SessionList interface {
	// Push adds session to collection
	Push(session Session, uid string) int
	// Get retries all active sessions for the user
	Get(uid string) []Session
	// Pull removes session object from collection
	Pull(session Session) bool
	// Size returns count of active sessions
	Size(uid string) int

	// OnOpen handler at init new session
	OnOpen(func(Session))
	// OnClose handler at close session
	OnClose(func(Session))
	// OnOpenExecute fire it if session open
	// Automatical if used Router.Sessions
	OnOpenExecute(Session)
	// OnCloseExecute fire it if session close
	// Automatical if used Router.Sessions
	OnCloseExecute(Session)
}

type SessionListImpl struct {
	sync.Mutex

	sessions map[string]string
	users    map[string][]Session

	onOpen  func(Session)
	onClose func(Session)
}

func NewSessionList() SessionList {
	return &SessionListImpl{
		sessions: make(map[string]string),
		users:    make(map[string][]Session),
	}
}

func (s *SessionListImpl) Push(session Session, uid string) int {
	s.Lock()
	defer s.Unlock()
	Loger.Debugf("%s Push session %s for user %v", logPrefix, session.ID(), uid)
	sessions, ok := s.users[uid]
	if !ok {
		sessions = make([]Session, 0)
	}
	if _, ok := s.sessions[session.ID()]; !ok {
		s.users[uid] = append(sessions, session)
		s.sessions[session.ID()] = uid
	}
	Loger.Debugf("%s Session %s for user %v pushed, total %v", logPrefix, session.ID(), uid, len(s.users[uid]))
	return len(s.users[uid])
}

func (s *SessionListImpl) Get(uid string) []Session {
	s.Lock()
	defer s.Unlock()
	sessions, ok := s.users[uid]
	if !ok {
		sessions = make([]Session, 0)
	}
	return sessions
}

func (s *SessionListImpl) Pull(session Session) bool {
	s.Lock()
	defer s.Unlock()
	uid, ok := s.sessions[session.ID()]
	if !ok {
		return false
	}
	Loger.Debugf("%s Pull session %s for user %v", logPrefix, session.ID(), uid)
	var found int
	for key, value := range s.users[uid] {
		if value.ID() == session.ID() {
			found = key
		}
	}
	s.users[uid] = append(s.users[uid][:found], s.users[uid][found+1:]...)
	delete(s.sessions, session.ID())
	Loger.Debugf("%s Session %s for user %v pulled, total %v", logPrefix, session.ID(), uid, len(s.users[uid]))
	return true
}

func (s *SessionListImpl) Size(uid string) int {
	s.Lock()
	defer s.Unlock()
	sessions, ok := s.users[uid]
	if !ok {
		return 0
	}
	return len(sessions)
}

func (s *SessionListImpl) OnOpen(handler func(session Session)) {
	s.onOpen = handler
}

func (s *SessionListImpl) OnClose(handler func(session Session)) {
	s.onClose = handler
}

// OnOpenExecute Execute open handler
// Use if custom SessionList
func (s *SessionListImpl) OnOpenExecute(session Session) {
	if s.onOpen != nil {
		s.onOpen(session)
	}
}

// OnCloseExecute Execute close handler and pull session from list
// Use if custom SessionList
func (s *SessionListImpl) OnCloseExecute(session Session) {
	if s.onClose != nil {
		s.onClose(session)
		s.Pull(session)
	}
}
