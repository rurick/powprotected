package app

import (
	"errors"
	"fmt"
	"io"
	"net"
	"time"
)

const (
	tcpBuffSize = 2048
	ttl         = 5 * time.Second
)

var (
	ErrConnClosed = errors.New("connection closed")
)

func Write(conn net.Conn, data []byte) error {
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
	if _, err := conn.Write([]byte("\n")); err != nil {
		return err
	}
	return nil
}

func Read(conn net.Conn) ([]byte, error) {
	var out []byte
	buf := make([]byte, tcpBuffSize)
	if err := conn.SetReadDeadline(time.Now().Add(ttl)); err != nil { // check timeout
		return nil, fmt.Errorf("can't set read deadline: %v", err)
	}
	for {
		n, err := conn.Read(buf)
		if err != nil {
			if err == io.EOF {
				return nil, ErrConnClosed
			}
			return nil, fmt.Errorf("read error: %v", err)
		}
		if n == 0 {
			return nil, errors.New("unexpected empty buffer")
		}
		out = append(out, buf[:n]...)
		if buf[n-1] == '\n' {
			out = out[:len(out)-1]
			return out, nil
		}
	}
}

func min(x, y int) int {
	if x < y {
		return x
	}
	return y
}
