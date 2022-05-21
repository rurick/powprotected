// Copyright (c) 2022 Yuriy Iovkov

package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"net"
	"time"

	"github.com/rurick/powprotected/internal/app/client"
	"github.com/rurick/powprotected/pkg/shutdown"
	"github.com/sirupsen/logrus"
)

const terminateTimeout = 10 * time.Second

func main() {
	log := logrus.New()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	addr := flag.String("a", "127.0.0.1:8888", "server address")
	flag.Parse()
	if addr == nil {
		log.Fatal("server address is not set")
	}

	sigHandler := shutdown.TermSignalTrap()
	go func() {
		if err := sigHandler.Wait(ctx); err != nil &&
			!errors.Is(err, shutdown.ErrTermSig) &&
			!errors.Is(err, context.Canceled) {
			log.Fatal(err)
		}
		cancel()
		log.Info("termination by sig")
		<-time.After(terminateTimeout)
		log.Fatal("termination timeout")
	}()

	// run client
	log.Infof("connecting to %s", *addr)
	conn, err := net.Dial("tcp", *addr)
	if err != nil {
		log.Fatal("can't context to the server")
	}
	log.Info("connected")
	cl := client.New(conn, log)
	wow, err := cl.Run()
	if err != nil {
		log.Errorf("run time error: %v", err)
	}

	fmt.Printf("Server answer: %s", string(wow))

}
