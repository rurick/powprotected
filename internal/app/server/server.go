package server

import (
	"context"
	"fmt"
	"net"
	"sync"
	"time"

	"github.com/rurick/powprotected/internal/app"
	"github.com/rurick/powprotected/internal/domain/challenge"
	"github.com/rurick/powprotected/internal/domain/wow"
	"github.com/sirupsen/logrus"
)

const (
	handlerTtl   = 5 * time.Second
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
	s.logger.Infof("listen tcp connection to %s", s.addr)
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

// обработка установленых соединений
func (s *TCPServer) serve() {
	for {
		// step 1 - установлено соединение
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
		s.logger.Info("connect from ", conn.RemoteAddr())
		go func() {
			if err := s.handle(conn); err != nil {
				s.logger.Errorf("handle error: %v", err)
			}

		}()
	}
}

// обмен данными с клиентом. проверка доступа и передача результата
func (s *TCPServer) handle(conn net.Conn) error {
	s.wg.Add(1)
	closed := make(chan bool)
	defer func(closed chan bool) {
		s.wg.Done()
		if err := conn.Close(); err != nil {
			s.logger.Errorf("can't close connection: %v", err)
			return
		}
		close(closed)
		s.logger.Info("connection closed")
	}(closed)
	go func(closed chan bool) {
		<-time.After(handlerTtl)
		select {
		case <-closed:
		default:
			_ = conn.Close()
			s.logger.Info("connection closed by timeout")
		}
	}(closed)

	// step 2 - отправка челенжа
	c := challenge.NewChallenge()
	out := challenge.NewRequest(c.Set(), c.Hash())
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
	choice := string(data)

	// step 4 - поверка и отправка ответа
	answer := accessDenied
	if choice == c.Choice() {
		g := wow.New()
		answer = g.Get()
	}
	if err = s.write(conn, []byte(answer)); err != nil {
		return err
	}
	return nil
}

func (s TCPServer) write(conn net.Conn, data []byte) error {
	return app.Write(conn, data)
}

func (s *TCPServer) read(conn net.Conn) ([]byte, error) {
	return app.Read(conn)
}
