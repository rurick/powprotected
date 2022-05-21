package client

import (
	"crypto/sha1"
	"fmt"
	"net"
	"time"

	"github.com/rurick/powprotected/internal/app"
	"github.com/rurick/powprotected/internal/domain/challenge"
	"github.com/sirupsen/logrus"
)

type Client struct {
	conn   net.Conn
	logger *logrus.Logger
}

func New(conn net.Conn, logger *logrus.Logger) *Client {
	return &Client{
		conn:   conn,
		logger: logger,
	}
}

func (c *Client) Run() ([]byte, error) {
	task, err := c.read()
	if err != nil {
		return nil, fmt.Errorf("client run error: %v", err)
	}
	// step 1: оплучить задачу
	req := challenge.NewRequest(nil, "")
	if err = req.Decode(task); err != nil {
		return nil, fmt.Errorf("request decode error: %v", err)
	}
	// step 2: найти решение
	res := c.resolve(req)
	if res == "" {
		return nil, fmt.Errorf("solution not found in set")
	}
	// step 3: отправить решение
	if err = c.write([]byte(res)); err != nil {
		return nil, fmt.Errorf("error write to socket %v", err)
	}
	// step 4: получить услугу
	return c.read()
}

func (c *Client) resolve(req *challenge.Request) string {
	start := time.Now().UnixNano()
	defer func() {
		c.logger.Infof("resolve time %d ns", time.Now().UnixNano()-start)
	}()
	set := req.Set
	a := req.Answer
	for _, v := range set {
		if sha1hash(v) == a { // found
			return v
		}
	}
	return ""
}

func (c *Client) write(data []byte) error {
	return app.Write(c.conn, data)
}

func (c *Client) read() ([]byte, error) {
	return app.Read(c.conn)
}

func sha1hash(s string) string {
	h := sha1.New()
	h.Write([]byte(s))
	return fmt.Sprintf("%x", h.Sum(nil))
}
