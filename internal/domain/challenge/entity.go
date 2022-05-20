package challenge

import (
	"crypto/sha1"
	"fmt"
	"math/rand"
	"time"
)

const (
	maxSize = 1024
	minSize = 100
	maxLen  = 10
	minLen  = 5
	dict    = "1234567890abcdefghijklmnopqrstuvwxyABCDEFGHIJKLMNOPQRSTUVWXYZ"
)

type Challenge struct {
	set    []string
	choice string
	size   int
}

func NewChallenge() *Challenge {
	c := &Challenge{}
	c.size = rand.Intn(maxSize-minSize) + minSize
	c.set = make([]string, c.size, c.size)
	for i := 0; i < c.size; i++ {
		c.set[i] = rndString()
	}
	c.choice = c.set[rand.Intn(c.size)]
	return c
}

func (c *Challenge) Choice() string {
	return c.choice
}

func (c *Challenge) Set() []string {
	return c.set
}

func (c *Challenge) Hash() string {
	h := sha1.New()
	h.Write([]byte(c.choice))
	return fmt.Sprintf("%x", h.Sum(nil))
}

func rndString() string {
	dictLen := len(dict)
	s := ""
	l := rand.Intn(maxLen-minLen) + minLen
	for i := 0; i < l; i++ {
		s += string(dict[rand.Intn(dictLen)])
	}
	return s
}

func init() {
	rand.Seed(time.Now().UnixNano())
}
