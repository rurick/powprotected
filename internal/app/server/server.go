package server

import (
	"context"
	"fmt"
	"io"
	"net"
	"sync"
	"time"

	"github.com/rurick/powprotected/internal/domain/challenge"
	"github.com/rurick/powprotected/internal/domain/wow"
	"github.com/sirupsen/logrus"
)

const (
	tcpBuffSize  = 2048
	ttl          = 5 * time.Second
	handlerTtl   = 10 * time.Second
	accessDenied = "Error: access denied"
)

type TCPServer struct {
	addr    string
	logger  *logrus.Logger
	l       net.Listener
	ctx     context.Context
	stopped chan interface{} // признак остановки сервера
	wg      *sync.WaitGroup
}

func New(addr string, log *logrus.Logger) *TCPServer {
	return &TCPServer{
		addr:    addr,
		logger:  log,
		stopped: make(chan interface{}),
		wg:      &sync.WaitGroup{},
	}
}

func (s *TCPServer) Start(ctx context.Context) (err error) {
	s.l, err = net.Listen("tcp", s.addr)

	if err != nil {
		return fmt.Errorf("failed to listen addr %s with error: %w", s.addr, err)
	}
	s.logger.Info("listen addr: %s", s.addr)
	s.ctx = ctx
	go s.serve()
	go func() {
		<-s.ctx.Done()
		s.Stop()
	}()
	return
}

func (s *TCPServer) Stop() {
	s.logger.Info("stopping tcp server...")
	defer s.logger.Info("tcp server stopped")
	close(s.stopped)

	if err := s.l.Close(); err != nil {
		s.logger.Error(err)
	}
	s.wg.Wait()
}

func (s *TCPServer) serve() {
	for {
		conn, err := s.l.Accept()
		if err != nil {
			select {
			case <-s.stopped:
				s.logger.Info("stop accepting new connections")
				return
			default:
				s.logger.Errorf("can't accept connection %v", err)
				continue
			}
		}
		go func() {
			if err := s.handle(conn); err != nil {
				s.logger.Errorf("handle error: %v", err)
			}

		}()
	}
}

func (s *TCPServer) handle(conn net.Conn) error {
	s.wg.Add(1)
	defer func() {
		s.wg.Done()
		if err := conn.Close(); err != nil {
			s.logger.Errorf("can't close connection: %v", err)
		}
	}()
	go func() {
		<-time.After(handlerTtl)
		_ = conn.Close()
	}()

	// step 1 - установлено соединение
	//
	// step 2 - отправка челенжа
	c := challenge.NewChallenge()
	out := challenge.NewRequest(c)
	buf, err := out.Encode()
	if err != nil {
		return fmt.Errorf("encode error: %v", err)
	}
	if err = s.write(conn, buf); err != nil {
		return fmt.Errorf("socket write error: %v", err)
	}

	// step 3 - получение ответа от клиента
	data, err := s.read(conn)
	if err != nil {
		return err
	}
	hash := string(data)

	// step 4 - поверка и отправка ответа
	answer := accessDenied
	if hash == c.Hash() {
		g := wow.New()
		answer = g.Get()
	}
	if err = s.write(conn, []byte(answer)); err != nil {
		return err
	}
	return nil
}

func (s TCPServer) write(conn net.Conn, data []byte) error {
	for {
		size := min(tcpBuffSize, len(data))
		if size == 0 {
			break
		}
		n, err := conn.Write(data[:size])
		if err != nil {
			return err
		}
		data = data[n:]
	}
	return nil
}

func (s *TCPServer) read(conn net.Conn) ([]byte, error) {
	buf := make([]byte, tcpBuffSize)
	if err := conn.SetReadDeadline(time.Now().Add(ttl)); err != nil {
		return nil, fmt.Errorf("can't set read deadline: %v", err)
	}
	for {
		n, err := conn.Read(buf)
		if err != nil && err != io.EOF {
			return nil, fmt.Errorf("read error: %v", err)
		}
		if n == 0 {
			return buf, nil
		}
		s.logger.Info("received from %v: %s", conn.RemoteAddr(), string(buf[:n]))
	}
}

func min(x, y int) int {
	if x < y {
		return x
	}
	return y
}
