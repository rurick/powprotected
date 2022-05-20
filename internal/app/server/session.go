package server

import "net"

type Session struct {
	conn net.Conn
}

func NewSessions(conn net.Conn) *Session {
	return &Session{conn: conn}
}

func (s *Session) Handle() error {
	return nil
}
