package challenge

import (
	"encoding/json"
)

type Request struct {
	Set    []string // словарь
	Answer string   // хэш ответа
}

func NewRequest(set []string, answer string) *Request {
	return &Request{
		Set:    set,
		Answer: answer,
	}
}

func (c *Request) Encode() ([]byte, error) {
	return json.Marshal(c)
}

func (c *Request) Decode(data []byte) error {
	return json.Unmarshal(data, c)
}
