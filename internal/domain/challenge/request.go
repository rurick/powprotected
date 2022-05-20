package challenge

import (
	"bytes"
	"encoding/gob"
)

const HashMethod = "SHA1"

type Request struct {
	set    []string // словарь
	answer string   // хэш ответа
	method string   // метод построение хеша
}

func NewRequest(c *Challenge) *Request {
	return &Request{
		method: HashMethod,
		set:    c.Set(),
		answer: c.Hash(),
	}
}

func (c *Request) Encode() ([]byte, error) {
	buf := bytes.Buffer{}
	enc := gob.NewEncoder(&buf)
	if err := enc.Encode(c); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}
